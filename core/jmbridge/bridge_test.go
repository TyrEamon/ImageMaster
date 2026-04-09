package jmbridge

import "testing"

func TestSupports(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "vip photo", url: "https://18comic.vip/photo/123456", want: true},
		{name: "org album", url: "https://18comic.org/album/123456", want: true},
		{name: "other host", url: "https://example.com/photo/123456", want: false},
		{name: "invalid", url: "not a url", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Supports(test.url); got != test.want {
				t.Fatalf("Supports(%q) = %v, want %v", test.url, got, test.want)
			}
		})
	}
}

func TestGetRuntimeInfoDefaults(t *testing.T) {
	info := GetRuntimeInfo()
	if info.Name == "" {
		t.Fatal("expected runtime name")
	}
	if info.Engine == "" {
		t.Fatal("expected runtime engine")
	}
	if info.Upstream == "" {
		t.Fatal("expected runtime upstream")
	}
}
