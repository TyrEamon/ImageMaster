package request

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	tls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

type uTransport struct {
	tr1 *http.Transport
	tr2 *http2.Transport
}

func (*uTransport) newSpec() *tls.ClientHelloSpec {
	return &tls.ClientHelloSpec{
		TLSVersMax:         tls.VersionTLS13,
		TLSVersMin:         tls.VersionTLS12,
		CipherSuites:       []uint16{tls.GREASE_PLACEHOLDER, 0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8, 0xc013, 0xc014, 0x009c, 0x009d, 0x002f, 0x0035},
		CompressionMethods: []uint8{0x0},
		Extensions: []tls.TLSExtension{
			&tls.UtlsGREASEExtension{},
			&tls.SNIExtension{},
			&tls.ExtendedMasterSecretExtension{},
			&tls.RenegotiationInfoExtension{},
			&tls.SupportedCurvesExtension{Curves: []tls.CurveID{tls.GREASE_PLACEHOLDER, tls.X25519, tls.CurveP256, tls.CurveP384}},
			&tls.SupportedPointsExtension{SupportedPoints: []byte{0x0}},
			&tls.SessionTicketExtension{},
			&tls.ALPNExtension{AlpnProtocols: []string{"http/1.1"}},
			&tls.StatusRequestExtension{},
			&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{0x0403, 0x0804, 0x0401, 0x0503, 0x0805, 0x0501, 0x0806, 0x0601}},
			&tls.SCTExtension{},
			&tls.KeyShareExtension{KeyShares: []tls.KeyShare{
				{Group: tls.CurveID(tls.GREASE_PLACEHOLDER), Data: []byte{0}},
				{Group: tls.X25519},
			}},
			&tls.PSKKeyExchangeModesExtension{Modes: []uint8{tls.PskModeDHE}},
			&tls.SupportedVersionsExtension{Versions: []uint16{tls.GREASE_PLACEHOLDER, tls.VersionTLS13, tls.VersionTLS12}},
			&tls.UtlsCompressCertExtension{Algorithms: []tls.CertCompressionAlgo{tls.CertCompressionBrotli}},
			&tls.ApplicationSettingsExtension{SupportedProtocols: []string{"h2"}},
			&tls.UtlsGREASEExtension{},
			&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
		},
		GetSessionID: nil,
	}
}

func (u *uTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == "http" {
		// 对于 http 请求，直接使用普通 RoundTrip 处理
		return u.tr1.RoundTrip(req)
	} else if req.URL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme: %s", req.URL.Scheme)
	}

	// 检查是否配置了代理
	var proxyURL *url.URL
	var err error
	if u.tr1.Proxy != nil {
		proxyURL, err = u.tr1.Proxy(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get proxy URL: %v", err)
		}
	}

	// 如果没有配置代理，直接使用标准 transport
	if proxyURL == nil {
		return u.tr1.RoundTrip(req)
	}

	// 连接到 HTTP 代理
	conn, err := net.DialTimeout("tcp", proxyURL.Host, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %v", err)
	}

	// 发送 CONNECT 请求
	dest := req.URL.Host
	if req.URL.Port() == "" {
		dest += ":443"
	}
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", dest, req.URL.Host)
	_, err = conn.Write([]byte(connectReq))
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send CONNECT request: %v", err)
	}

	// 读取代理响应
	respReader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(respReader, req)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read proxy response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		conn.Close()
		return nil, fmt.Errorf("proxy CONNECT failed with status: %s", resp.Status)
	}

	// 代理建立了隧道，创建 TLS 连接
	tlsConn := tls.UClient(conn, &tls.Config{ServerName: req.URL.Hostname()}, tls.HelloCustom)
	if err = tlsConn.ApplyPreset(u.newSpec()); err != nil {
		tlsConn.Close()
		return nil, fmt.Errorf("uConn.ApplyPreset() error: %+v", err)
	}
	if err = tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		return nil, fmt.Errorf("TLS handshake failed: %+v", err)
	}

	// 处理 ALPN (HTTP/2 或 HTTP/1.1)
	alpn := tlsConn.ConnectionState().NegotiatedProtocol
	switch alpn {
	case "h2":
		req.Proto = "HTTP/2.0"
		req.ProtoMajor = 2
		req.ProtoMinor = 0

		if c, err := u.tr2.NewClientConn(tlsConn); err == nil {
			return c.RoundTrip(req)
		} else {
			return nil, fmt.Errorf("http2.Transport.NewClientConn() error: %+v", err)
		}

	case "http/1.1", "":
		req.Proto = "HTTP/1.1"
		req.ProtoMajor = 1
		req.ProtoMinor = 1

		if err := req.Write(tlsConn); err == nil {
			return http.ReadResponse(bufio.NewReader(tlsConn), req)
		} else {
			return nil, fmt.Errorf("http.Request.Write() error: %+v", err)
		}

	default:
		return nil, fmt.Errorf("unsupported ALPN: %v", alpn)
	}
}
