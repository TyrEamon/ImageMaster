// ==UserScript==
// @name         18comic Album Download Helper(JM本子文件助手)
// @namespace    https://local.tampermonkey/
// @version      0.1.2
// @description  Keep 18comic album download pages on-site, assist manual captcha flow, and save official ZIPs into per-album folders.
// @match        https://18comic.vip/album_download/*
// @grant        GM_download
// @grant        GM_notification
// @grant        GM_registerMenuCommand
// @grant        GM_xmlhttpRequest
// @connect      *
// @run-at       document-start
// ==/UserScript==

(function () {
    'use strict';

    const PANEL_ID = 'tm-18comic-download-helper';
    const DB_NAME = 'tm-18comic-download-helper';
    const DB_STORE = 'kv';
    const DB_VERSION = 1;
    const BASE_DIR_HANDLE_KEY = 'baseDirHandle';
    const BASE_DIR_LABEL_KEY = 'baseDirLabel';
    const AUTO_CAPTURE_KEY = 'autoCaptureEnabled';
    const PANEL_POSITION_KEY = 'panelPosition';
    const PANEL_COLLAPSED_KEY = 'panelCollapsed';
    const DEBUG_LOG_LIMIT = 120;

    const state = {
        adGuardUntil: 0,
        panel: null,
        stats: {},
        baseDirHandle: null,
        baseDirLabel: localStorage.getItem(BASE_DIR_LABEL_KEY) || '',
        autoCaptureEnabled: localStorage.getItem(AUTO_CAPTURE_KEY) !== '0',
        captureInProgress: false,
        nativeSubmitBypass: new WeakSet(),
        albumId: '',
        albumTitle: '',
        lastArchiveName: '',
        lastSavedLocation: '',
        message: '准备中',
        tone: 'info',
        debugLogs: [],
        lastStatusKey: '',
        scanTimer: 0,
        panelCollapsed: localStorage.getItem(PANEL_COLLAPSED_KEY) === '1',
        panelPosition: null,
        draggingPanel: false,
    };

    const PANEL_STYLE = `
#${PANEL_ID} {
    position: fixed;
    top: 16px;
    right: 16px;
    z-index: 2147483647;
    width: 420px;
    max-width: calc(100vw - 32px);
    border: 1px solid rgba(255, 255, 255, 0.14);
    border-radius: 14px;
    background: rgba(15, 23, 42, 0.96);
    box-shadow: 0 14px 38px rgba(0, 0, 0, 0.32);
    color: #e5eef8;
    font: 13px/1.5 -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    backdrop-filter: blur(12px);
}
#${PANEL_ID} * {
    box-sizing: border-box;
}
#${PANEL_ID} .tm-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    cursor: grab;
    user-select: none;
}
#${PANEL_ID}.tm-dragging .tm-head {
    cursor: grabbing;
}
#${PANEL_ID} .tm-head-main {
    min-width: 0;
    flex: 1;
}
#${PANEL_ID} .tm-title {
    font-weight: 700;
    font-size: 14px;
}
#${PANEL_ID} .tm-status {
    font-size: 12px;
    word-break: break-word;
}
#${PANEL_ID} .tm-head-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
}
#${PANEL_ID} .tm-icon-btn {
    min-width: 0;
    padding: 6px 10px;
    border-radius: 999px;
    background: rgba(37, 99, 235, 0.18);
    color: #dbeafe;
}
#${PANEL_ID} .tm-icon-btn:hover {
    background: rgba(59, 130, 246, 0.3);
}
#${PANEL_ID}[data-tone="ok"] .tm-status {
    color: #a7f3d0;
}
#${PANEL_ID}[data-tone="warn"] .tm-status {
    color: #fcd34d;
}
#${PANEL_ID}[data-tone="error"] .tm-status {
    color: #fda4af;
}
#${PANEL_ID}[data-tone="info"] .tm-status {
    color: #93c5fd;
}
#${PANEL_ID} .tm-body {
    padding: 12px 14px 14px;
}
#${PANEL_ID}[data-collapsed="1"] .tm-body {
    display: none;
}
#${PANEL_ID}[data-collapsed="1"] .tm-head {
    border-bottom: 0;
    padding-bottom: 12px;
}
#${PANEL_ID} .tm-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
    margin-bottom: 10px;
}
#${PANEL_ID} .tm-stat {
    padding: 8px 10px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.06);
}
#${PANEL_ID} .tm-stat-label {
    display: block;
    color: #9fb0c2;
    font-size: 11px;
}
#${PANEL_ID} .tm-stat-value {
    display: block;
    margin-top: 3px;
    font-size: 13px;
    font-weight: 600;
    word-break: break-word;
}
#${PANEL_ID} .tm-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 10px;
}
#${PANEL_ID} button {
    appearance: none;
    border: 0;
    border-radius: 10px;
    padding: 8px 10px;
    cursor: pointer;
    background: #2563eb;
    color: #eff6ff;
    font: inherit;
}
#${PANEL_ID} button[data-action="pick-dir"] {
    background: #0f766e;
}
#${PANEL_ID} button[data-action="toggle-capture"] {
    background: #7c3aed;
}
#${PANEL_ID} button[data-action="focus-captcha"] {
    background: #b45309;
}
#${PANEL_ID} button[data-action="copy-log"] {
    background: #1d4ed8;
}
#${PANEL_ID} button[data-action="clear-log"] {
    background: #475569;
}
#${PANEL_ID} .tm-note {
    color: #9fb0c2;
    font-size: 11px;
    white-space: pre-wrap;
}
#${PANEL_ID} .tm-log-head {
    margin: 10px 0 6px;
    color: #9fb0c2;
    font-size: 11px;
    font-weight: 600;
}
#${PANEL_ID} .tm-log {
    max-height: 180px;
    overflow: auto;
    padding: 10px;
    border-radius: 10px;
    background: rgba(2, 6, 23, 0.72);
    color: #dbeafe;
    font: 11px/1.45 Consolas, "SFMono-Regular", Menlo, monospace;
    white-space: pre-wrap;
    word-break: break-word;
}
`;

    const originalWindowOpen = window.open ? window.open.bind(window) : null;

    function wait(ms) {
        return new Promise((resolve) => setTimeout(resolve, ms));
    }

    function sanitizeText(value) {
        return String(value ?? '').replace(/\s+/g, ' ').trim();
    }

    function toAbsoluteUrl(url, base) {
        try {
            return new URL(url, base || location.href).toString();
        } catch {
            return '';
        }
    }

    function sanitizePathSegment(value, fallback = 'untitled') {
        const cleaned = String(value || '')
            .replace(/[<>:"/\\|?*\u0000-\u001F]/g, '_')
            .replace(/\s+/g, ' ')
            .replace(/[. ]+$/g, '')
            .trim()
            .slice(0, 120);
        return cleaned || fallback;
    }

    function parseHeaders(rawHeaders) {
        const result = {};
        String(rawHeaders || '')
            .split(/\r?\n/)
            .map((line) => line.trim())
            .filter(Boolean)
            .forEach((line) => {
                const index = line.indexOf(':');
                if (index <= 0) {
                    return;
                }
                const key = line.slice(0, index).trim().toLowerCase();
                const value = line.slice(index + 1).trim();
                if (!result[key]) {
                    result[key] = value;
                }
            });
        return result;
    }

    function decodeArrayBuffer(buffer, contentType = '') {
        try {
            const charsetMatch = String(contentType).match(/charset=([^;]+)/i);
            const charset = charsetMatch ? charsetMatch[1].trim() : 'utf-8';
            return new TextDecoder(charset).decode(new Uint8Array(buffer));
        } catch {
            return new TextDecoder('utf-8').decode(new Uint8Array(buffer));
        }
    }

    function extractFilenameFromDisposition(disposition) {
        const text = String(disposition || '');
        const utf8Match = text.match(/filename\*\s*=\s*UTF-8''([^;]+)/i);
        if (utf8Match) {
            try {
                return decodeURIComponent(utf8Match[1]);
            } catch {
                return utf8Match[1];
            }
        }

        const plainMatch = text.match(/filename\s*=\s*"([^"]+)"/i) || text.match(/filename\s*=\s*([^;]+)/i);
        return plainMatch ? plainMatch[1].trim() : '';
    }

    function getAlbumId() {
        const match = location.pathname.match(/\/album_download\/(\d+)/i);
        return match ? match[1] : '';
    }

    function normalizeTitle(text) {
        return sanitizeText(text)
            .replace(/\s*\|\s*18comic.*$/i, '')
            .replace(/\s*-\s*18comic.*$/i, '')
            .replace(/\s+\d+(?:\.\d+)?\s*(?:KB|MB|GB)\s*$/i, '')
            .trim();
    }

    function getAlbumTitle() {
        const selectors = [
            'meta[property="og:title"]',
            'meta[name="title"]',
            'h1',
            '.panel-heading',
            '.book-name',
            '.episode-name',
            '.album-name',
        ];

        for (const selector of selectors) {
            const node = document.querySelector(selector);
            const value = normalizeTitle(node?.getAttribute?.('content') || node?.textContent || '');
            if (value && !/18comic/i.test(value)) {
                return value;
            }
        }

        const title = normalizeTitle(document.title);
        return title || `JM${getAlbumId() || 'download'}`;
    }

    function shouldBlockGuardedNavigation(url) {
        const absoluteUrl = toAbsoluteUrl(url);
        return Boolean(Date.now() < state.adGuardUntil && absoluteUrl && !absoluteUrl.startsWith(location.origin));
    }

    function notify(message) {
        if (typeof GM_notification !== 'function') {
            return;
        }

        GM_notification({
            title: '18comic Download Helper',
            text: message,
            timeout: 4000,
        });
    }

    function formatLogDetail(detail) {
        if (!detail) {
            return '';
        }

        const text = typeof detail === 'string' ? detail : JSON.stringify(detail);
        return text.length > 220 ? `${text.slice(0, 220)}...` : text;
    }

    function refreshLogPanel() {
        if (!state.stats.debugLog) {
            return;
        }

        state.stats.debugLog.textContent = state.debugLogs.length
            ? state.debugLogs.join('\n')
            : '[no logs yet]';
    }

    function logStep(step, detail = '') {
        const line = `[${new Date().toLocaleTimeString('zh-CN', { hour12: false })}] ${step}${detail ? ` | ${formatLogDetail(detail)}` : ''}`;
        if (state.debugLogs[state.debugLogs.length - 1] === line) {
            return;
        }

        state.debugLogs.push(line);
        if (state.debugLogs.length > DEBUG_LOG_LIMIT) {
            state.debugLogs.splice(0, state.debugLogs.length - DEBUG_LOG_LIMIT);
        }

        refreshLogPanel();
        console.info('[18comic helper]', step, detail || '');
    }

    function isInsidePanel(node) {
        return Boolean(
            state.panel &&
            node &&
            typeof node === 'object' &&
            'nodeType' in node &&
            state.panel.contains(node)
        );
    }

    function scheduleScan(reason = 'scheduled') {
        if (state.scanTimer) {
            clearTimeout(state.scanTimer);
        }

        state.scanTimer = window.setTimeout(() => {
            state.scanTimer = 0;
            scanPageState(reason);
        }, 120);
    }

    function loadPanelPosition() {
        try {
            const raw = localStorage.getItem(PANEL_POSITION_KEY);
            if (!raw) {
                return null;
            }

            const parsed = JSON.parse(raw);
            if (!parsed || !Number.isFinite(parsed.left) || !Number.isFinite(parsed.top)) {
                return null;
            }

            return {
                left: parsed.left,
                top: parsed.top,
            };
        } catch {
            return null;
        }
    }

    function clampPanelPosition(left, top) {
        if (!state.panel) {
            return { left, top };
        }

        const panelWidth = state.panel.offsetWidth || 420;
        const panelHeight = state.panel.offsetHeight || 240;
        const maxLeft = Math.max(8, window.innerWidth - panelWidth - 8);
        const maxTop = Math.max(8, window.innerHeight - panelHeight - 8);
        return {
            left: Math.min(Math.max(8, left), maxLeft),
            top: Math.min(Math.max(8, top), maxTop),
        };
    }

    function applyPanelPosition(position) {
        if (!state.panel) {
            return;
        }

        if (!position || !Number.isFinite(position.left) || !Number.isFinite(position.top)) {
            state.panel.style.left = '';
            state.panel.style.top = '';
            state.panel.style.right = '16px';
            state.panel.style.bottom = '';
            state.panelPosition = null;
            return;
        }

        const clamped = clampPanelPosition(position.left, position.top);
        state.panel.style.left = `${clamped.left}px`;
        state.panel.style.top = `${clamped.top}px`;
        state.panel.style.right = 'auto';
        state.panel.style.bottom = 'auto';
        state.panelPosition = clamped;
    }

    function persistPanelPosition(position) {
        if (!position) {
            localStorage.removeItem(PANEL_POSITION_KEY);
            return;
        }

        localStorage.setItem(PANEL_POSITION_KEY, JSON.stringify(position));
    }

    function setPanelCollapsed(nextValue) {
        state.panelCollapsed = Boolean(nextValue);
        localStorage.setItem(PANEL_COLLAPSED_KEY, state.panelCollapsed ? '1' : '0');

        if (state.panel) {
            state.panel.dataset.collapsed = state.panelCollapsed ? '1' : '0';
        }

        if (state.stats.collapseButton) {
            state.stats.collapseButton.textContent = state.panelCollapsed ? '展开' : '收起';
            state.stats.collapseButton.title = state.panelCollapsed ? '展开面板' : '收起面板';
        }
    }

    async function copyText(text) {
        try {
            if (navigator.clipboard?.writeText) {
                await navigator.clipboard.writeText(text);
                return true;
            }
        } catch {
            // ignore clipboard api failure
        }

        try {
            const textarea = document.createElement('textarea');
            textarea.value = text;
            textarea.setAttribute('readonly', 'readonly');
            textarea.style.position = 'fixed';
            textarea.style.left = '-9999px';
            document.body.appendChild(textarea);
            textarea.select();
            document.execCommand('copy');
            textarea.remove();
            return true;
        } catch {
            return false;
        }
    }

    function updateStatus(message, tone = 'info') {
        const statusKey = `${tone}:${message}`;
        if (state.lastStatusKey === statusKey) {
            return;
        }

        state.lastStatusKey = statusKey;
        state.message = message;
        state.tone = tone;
        logStep(`status:${tone}`, message);

        if (state.panel) {
            state.panel.dataset.tone = tone;
        }

        if (state.stats.status) {
            state.stats.status.textContent = message;
        }
    }

    function refreshPanel() {
        state.albumId = getAlbumId();
        state.albumTitle = getAlbumTitle();

        if (!state.panel) {
            return;
        }

        state.panel.dataset.tone = state.tone;
        state.stats.albumId.textContent = state.albumId || 'unknown';
        state.stats.albumTitle.textContent = state.albumTitle || 'unknown';
        state.stats.baseDir.textContent = state.baseDirLabel || '未设置';
        state.stats.capture.textContent = state.autoCaptureEnabled ? '已开启' : '已关闭';
        state.stats.lastArchive.textContent = state.lastArchiveName || '暂无';
        state.stats.lastSaved.textContent = state.lastSavedLocation || '暂无';
        state.stats.status.textContent = state.message;
        state.stats.toggleButton.textContent = state.autoCaptureEnabled ? '关闭自动接管' : '开启自动接管';
        refreshLogPanel();
    }

    function ensurePanel() {
        if (state.panel || !document.body) {
            return;
        }

        const style = document.createElement('style');
        style.textContent = PANEL_STYLE;
        document.documentElement.appendChild(style);

        const panel = document.createElement('section');
        panel.id = PANEL_ID;
        panel.dataset.tone = state.tone;
        panel.dataset.collapsed = state.panelCollapsed ? '1' : '0';
        panel.innerHTML = `
            <div class="tm-head">
                <div class="tm-head-main">
                    <div class="tm-title">18comic 下载助手</div>
                    <div class="tm-status"></div>
                </div>
                <div class="tm-head-actions">
                    <button type="button" class="tm-icon-btn" data-action="toggle-panel">收起</button>
                </div>
            </div>
            <div class="tm-body">
                <div class="tm-grid">
                    <div class="tm-stat">
                        <span class="tm-stat-label">Album ID</span>
                        <span class="tm-stat-value" data-field="album-id"></span>
                    </div>
                    <div class="tm-stat">
                        <span class="tm-stat-label">自动接管</span>
                        <span class="tm-stat-value" data-field="capture"></span>
                    </div>
                    <div class="tm-stat">
                        <span class="tm-stat-label">本子标题</span>
                        <span class="tm-stat-value" data-field="album-title"></span>
                    </div>
                    <div class="tm-stat">
                        <span class="tm-stat-label">保存目录</span>
                        <span class="tm-stat-value" data-field="base-dir"></span>
                    </div>
                    <div class="tm-stat">
                        <span class="tm-stat-label">最近 ZIP</span>
                        <span class="tm-stat-value" data-field="last-archive"></span>
                    </div>
                    <div class="tm-stat">
                        <span class="tm-stat-label">最近落盘</span>
                        <span class="tm-stat-value" data-field="last-saved"></span>
                    </div>
                </div>
                <div class="tm-actions">
                    <button type="button" data-action="pick-dir">选择保存目录</button>
                    <button type="button" data-action="toggle-capture"></button>
                    <button type="button" data-action="focus-captcha">聚焦答题框</button>
                    <button type="button" data-action="copy-log">Copy Log</button>
                    <button type="button" data-action="clear-log">Clear Log</button>
                </div>
                <div class="tm-note">标题栏可以拖动，右上角可以收起。点击“免等待，点我直接下载”时，脚本会尽量拦住广告跳转并把你留在原下载页。数学题由你手动填写，提交后脚本会优先接管 ZIP 下载并按本子名建文件夹保存。</div>
                <div class="tm-log-head">Debug Log</div>
                <div class="tm-log" data-field="debug-log"></div>
            </div>
        `;

        state.panel = panel;
        state.stats = {
            status: panel.querySelector('.tm-status'),
            albumId: panel.querySelector('[data-field="album-id"]'),
            albumTitle: panel.querySelector('[data-field="album-title"]'),
            baseDir: panel.querySelector('[data-field="base-dir"]'),
            capture: panel.querySelector('[data-field="capture"]'),
            lastArchive: panel.querySelector('[data-field="last-archive"]'),
            lastSaved: panel.querySelector('[data-field="last-saved"]'),
            debugLog: panel.querySelector('[data-field="debug-log"]'),
            toggleButton: panel.querySelector('[data-action="toggle-capture"]'),
            collapseButton: panel.querySelector('[data-action="toggle-panel"]'),
        };

        panel.querySelector('[data-action="toggle-panel"]').addEventListener('click', (event) => {
            event.preventDefault();
            event.stopPropagation();
            setPanelCollapsed(!state.panelCollapsed);
            if (!state.panelCollapsed) {
                scheduleScan('panel-expand');
            }
        });

        panel.querySelector('[data-action="pick-dir"]').addEventListener('click', () => {
            chooseBaseDirectory();
        });

        panel.querySelector('[data-action="toggle-capture"]').addEventListener('click', () => {
            state.autoCaptureEnabled = !state.autoCaptureEnabled;
            localStorage.setItem(AUTO_CAPTURE_KEY, state.autoCaptureEnabled ? '1' : '0');
            updateStatus(state.autoCaptureEnabled ? '已开启自动接管 ZIP 下载' : '已关闭自动接管 ZIP 下载', 'info');
            refreshPanel();
        });

        panel.querySelector('[data-action="focus-captcha"]').addEventListener('click', () => {
            if (focusCaptchaInput()) {
                updateStatus('已聚焦到数学题输入框', 'ok');
            } else {
                updateStatus('暂时还没找到答题输入框', 'warn');
            }
        });

        panel.querySelector('[data-action="copy-log"]').addEventListener('click', async () => {
            const ok = await copyText(state.debugLogs.join('\n'));
            updateStatus(ok ? '日志已复制到剪贴板' : '复制日志失败', ok ? 'ok' : 'warn');
        });

        panel.querySelector('[data-action="clear-log"]').addEventListener('click', () => {
            state.debugLogs = [];
            refreshLogPanel();
            logStep('log:cleared');
        });

        document.body.appendChild(panel);
        applyPanelPosition(state.panelPosition || loadPanelPosition());
        installPanelDrag();
        setPanelCollapsed(state.panelCollapsed);
        refreshPanel();
        logStep('panel:ready');
    }

    function installPanelDrag() {
        if (!state.panel) {
            return;
        }

        const head = state.panel.querySelector('.tm-head');
        if (!head) {
            return;
        }

        let dragState = null;

        const onMouseMove = (event) => {
            if (!dragState || !state.panel) {
                return;
            }

            const nextLeft = dragState.startLeft + (event.clientX - dragState.startX);
            const nextTop = dragState.startTop + (event.clientY - dragState.startY);
            applyPanelPosition({ left: nextLeft, top: nextTop });
        };

        const finishDrag = () => {
            if (!dragState || !state.panel) {
                return;
            }

            state.draggingPanel = false;
            state.panel.classList.remove('tm-dragging');
            persistPanelPosition(state.panelPosition);
            dragState = null;
            document.removeEventListener('mousemove', onMouseMove, true);
            document.removeEventListener('mouseup', finishDrag, true);
        };

        head.addEventListener('mousedown', (event) => {
            if (event.button !== 0) {
                return;
            }

            const target = event.target;
            if (target instanceof Element && target.closest('button, a, input, textarea, select, label')) {
                return;
            }

            const rect = state.panel.getBoundingClientRect();
            applyPanelPosition({ left: rect.left, top: rect.top });
            dragState = {
                startX: event.clientX,
                startY: event.clientY,
                startLeft: state.panelPosition?.left ?? rect.left,
                startTop: state.panelPosition?.top ?? rect.top,
            };
            state.draggingPanel = true;
            state.panel.classList.add('tm-dragging');
            document.addEventListener('mousemove', onMouseMove, true);
            document.addEventListener('mouseup', finishDrag, true);
            event.preventDefault();
        });

        window.addEventListener('resize', () => {
            if (state.panelPosition) {
                applyPanelPosition(state.panelPosition);
                persistPanelPosition(state.panelPosition);
            }
        });
    }

    function openDb() {
        return new Promise((resolve, reject) => {
            const request = indexedDB.open(DB_NAME, DB_VERSION);

            request.onupgradeneeded = () => {
                const db = request.result;
                if (!db.objectStoreNames.contains(DB_STORE)) {
                    db.createObjectStore(DB_STORE);
                }
            };

            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async function idbGet(key) {
        const db = await openDb();
        return new Promise((resolve, reject) => {
            const transaction = db.transaction(DB_STORE, 'readonly');
            const store = transaction.objectStore(DB_STORE);
            const request = store.get(key);
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async function idbSet(key, value) {
        const db = await openDb();
        return new Promise((resolve, reject) => {
            const transaction = db.transaction(DB_STORE, 'readwrite');
            const store = transaction.objectStore(DB_STORE);
            const request = store.put(value, key);
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }

    async function loadPersistedDirectoryHandle(interactive = false) {
        if (!window.showDirectoryPicker || !window.isSecureContext) {
            logStep('dir:restore-skip', 'showDirectoryPicker unavailable');
            return null;
        }

        try {
            const handle = await idbGet(BASE_DIR_HANDLE_KEY);
            if (!handle) {
                return null;
            }

            const permissionOptions = { mode: 'readwrite' };
            const current = await handle.queryPermission(permissionOptions);
            if (current === 'granted') {
                state.baseDirHandle = handle;
                if (!state.baseDirLabel) {
                    state.baseDirLabel = handle.name || '';
                }
                logStep('dir:restored', state.baseDirLabel || handle.name || 'unknown');
                return handle;
            }

            if (interactive) {
                const next = await handle.requestPermission(permissionOptions);
                if (next === 'granted') {
                    state.baseDirHandle = handle;
                    if (!state.baseDirLabel) {
                        state.baseDirLabel = handle.name || '';
                    }
                    logStep('dir:permission-granted', state.baseDirLabel || handle.name || 'unknown');
                    return handle;
                }
            }
        } catch (error) {
            console.warn('[18comic helper] failed to restore directory handle', error);
            logStep('dir:restore-error', error?.message || String(error));
        }

        return null;
    }

    async function chooseBaseDirectory() {
        logStep('dir:pick-click');
        if (!window.showDirectoryPicker || !window.isSecureContext) {
            updateStatus('当前浏览器不支持目录写入，只能回退到浏览器默认下载目录', 'warn');
            notify('当前浏览器不支持目录写入');
            return;
        }

        try {
            const handle = await window.showDirectoryPicker({ mode: 'readwrite' });
            state.baseDirHandle = handle;
            state.baseDirLabel = handle.name || '已选择目录';
            localStorage.setItem(BASE_DIR_LABEL_KEY, state.baseDirLabel);
            await idbSet(BASE_DIR_HANDLE_KEY, handle);
            logStep('dir:picked', state.baseDirLabel);
            updateStatus(`已设置保存目录：${state.baseDirLabel}`, 'ok');
            refreshPanel();
        } catch (error) {
            if (error?.name === 'AbortError') {
                logStep('dir:pick-cancelled');
                updateStatus('目录选择已取消', 'warn');
                return;
            }

            console.error('[18comic helper] choose directory failed', error);
            logStep('dir:pick-error', error?.message || String(error));
            updateStatus('目录选择失败', 'error');
        }
    }

    async function ensureBaseDirHandle(allowInteractiveRetry = false) {
        if (state.baseDirHandle) {
            return state.baseDirHandle;
        }

        const passiveHandle = await loadPersistedDirectoryHandle(false);
        if (passiveHandle || !allowInteractiveRetry || !state.baseDirLabel) {
            return passiveHandle;
        }

        logStep('dir:permission-retry', state.baseDirLabel);
        updateStatus(`正在重新获取目录权限：${state.baseDirLabel}`, 'info');
        const retriedHandle = await loadPersistedDirectoryHandle(true);
        if (!retriedHandle) {
            logStep('dir:permission-missing', state.baseDirLabel);
        }
        return retriedHandle;
    }

    async function ensureAlbumFolderReady() {
        const cleanAlbumTitle = sanitizePathSegment(state.albumTitle || `JM${state.albumId}`);

        const baseHandle = await ensureBaseDirHandle(true);
        if (!baseHandle) {
            logStep('folder:base-missing', state.baseDirLabel || 'unset');
            return null;
        }

        try {
            await baseHandle.getDirectoryHandle(cleanAlbumTitle, { create: true });
            const folderPath = `${state.baseDirLabel || baseHandle.name}\\${cleanAlbumTitle}`;
            state.lastSavedLocation = folderPath;
            refreshPanel();
            logStep('folder:ready', folderPath);
            return {
                albumTitle: cleanAlbumTitle,
                folderPath,
            };
        } catch (error) {
            console.error('[18comic helper] prepare album folder failed', error);
            logStep('folder:error', error?.message || String(error));
            return null;
        }
    }

    async function saveBlobToConfiguredDirectory(blob, suggestedName) {
        const cleanAlbumTitle = sanitizePathSegment(state.albumTitle || `JM${state.albumId}`);
        const cleanFileName = sanitizePathSegment(suggestedName || `${cleanAlbumTitle}.zip`);
        logStep('save:start', `${cleanAlbumTitle}/${cleanFileName}`);

        if (state.baseDirHandle || await loadPersistedDirectoryHandle(false)) {
            const baseHandle = state.baseDirHandle;
            const folderHandle = await baseHandle.getDirectoryHandle(cleanAlbumTitle, { create: true });
            const fileHandle = await folderHandle.getFileHandle(cleanFileName, { create: true });
            const writable = await fileHandle.createWritable();
            await writable.write(blob);
            await writable.close();

            state.lastArchiveName = cleanFileName;
            state.lastSavedLocation = `${state.baseDirLabel || baseHandle.name}\\${cleanAlbumTitle}\\${cleanFileName}`;
            refreshPanel();
            logStep('save:filesystem-ok', state.lastSavedLocation);
            updateStatus(`ZIP 已保存到 ${state.lastSavedLocation}`, 'ok');
            notify(`已保存：${cleanFileName}`);
            return true;
        }

        const blobUrl = URL.createObjectURL(blob);
        const fallbackName = `${cleanAlbumTitle}/${cleanFileName}`;

        try {
            if (typeof GM_download === 'function') {
                GM_download({
                    url: blobUrl,
                    name: fallbackName,
                    saveAs: false,
                    onload: () => setTimeout(() => URL.revokeObjectURL(blobUrl), 2000),
                    onerror: () => setTimeout(() => URL.revokeObjectURL(blobUrl), 2000),
                    ontimeout: () => setTimeout(() => URL.revokeObjectURL(blobUrl), 2000),
                });
            } else {
                const anchor = document.createElement('a');
                anchor.href = blobUrl;
                anchor.download = cleanFileName;
                document.body.appendChild(anchor);
                anchor.click();
                anchor.remove();
                setTimeout(() => URL.revokeObjectURL(blobUrl), 2000);
            }

            state.lastArchiveName = cleanFileName;
            state.lastSavedLocation = `浏览器默认下载目录\\${fallbackName}`;
            refreshPanel();
            logStep('save:fallback-download', state.lastSavedLocation);
            updateStatus('已回退到浏览器默认下载目录', 'warn');
            notify(`已触发下载：${cleanFileName}`);
            return true;
        } catch (error) {
            logStep('save:error', error?.message || String(error));
            URL.revokeObjectURL(blobUrl);
            throw error;
        }
    }

    function gmRequest(options) {
        return new Promise((resolve, reject) => {
            if (typeof GM_xmlhttpRequest !== 'function') {
                logStep('request:unavailable');
                reject(new Error('GM_xmlhttpRequest is unavailable'));
                return;
            }

            logStep('request:start', `${options.method || 'GET'} ${options.url}`);

            GM_xmlhttpRequest({
                method: options.method || 'GET',
                url: options.url,
                headers: options.headers || {},
                data: options.data,
                timeout: options.timeout || 120000,
                responseType: options.responseType || 'arraybuffer',
                anonymous: false,
                onload: (response) => {
                    logStep('request:load', `${response.status || 0} ${response.finalUrl || options.url}`);
                    resolve(response);
                },
                onerror: (error) => {
                    logStep('request:error', error?.error || error?.message || options.url);
                    reject(error);
                },
                ontimeout: (error) => {
                    logStep('request:timeout', options.url);
                    reject(error);
                },
            });
        });
    }

    function headersToRaw(headers) {
        const lines = [];
        if (!headers?.forEach) {
            return '';
        }

        headers.forEach((value, key) => {
            lines.push(`${key}: ${value}`);
        });
        return lines.join('\n');
    }

    async function pageRequest(options) {
        const method = options.method || 'GET';
        logStep('request:page-start', `${method} ${options.url}`);

        const response = await fetch(options.url, {
            method,
            headers: options.headers || {},
            body: options.data,
            credentials: 'include',
            redirect: 'follow',
            cache: 'no-store',
        });

        logStep('request:page-load', `${response.status || 0} ${response.url || options.url}`);

        return {
            status: response.status,
            finalUrl: response.url || options.url,
            responseHeaders: headersToRaw(response.headers),
            response: await response.arrayBuffer(),
        };
    }

    function looksLikeArchiveUrl(url) {
        return /\.(zip|rar|7z)(?:$|[?#])/i.test(String(url || ''));
    }

    function looksLikeArchiveResponse(response, finalUrl) {
        const headers = parseHeaders(response.responseHeaders);
        const disposition = headers['content-disposition'] || '';
        const contentType = headers['content-type'] || '';
        const fileName = extractFilenameFromDisposition(disposition);

        return Boolean(
            /attachment/i.test(disposition) ||
            /(zip|octet-stream|x-rar|x-7z|gzip|compressed)/i.test(contentType) ||
            looksLikeArchiveUrl(fileName) ||
            looksLikeArchiveUrl(finalUrl)
        );
    }

    function extractArchiveUrlFromHtml(html, baseUrl) {
        const directMatch = String(html).match(/https?:\/\/[^"'\\s>]+?\.(zip|rar|7z)(?:\?[^"'\\s>]*)?/i);
        if (directMatch) {
            return directMatch[0];
        }

        const parser = new DOMParser();
        const doc = parser.parseFromString(String(html || ''), 'text/html');
        const nodes = Array.from(doc.querySelectorAll('a[href], form[action]'));

        for (const node of nodes) {
            const candidate = node.getAttribute('href') || node.getAttribute('action') || '';
            const text = sanitizeText(node.textContent || '');
            const url = toAbsoluteUrl(candidate || baseUrl, baseUrl);
            if (!url) {
                continue;
            }

            if (looksLikeArchiveUrl(url)) {
                return url;
            }

            if (/下载|download/i.test(text) && /zip|rar|7z|download/i.test(url)) {
                return url;
            }
        }

        return '';
    }

    function inferArchiveName(url, responseHeaders) {
        const parsedHeaders = parseHeaders(responseHeaders);
        const fromDisposition = extractFilenameFromDisposition(parsedHeaders['content-disposition']);
        if (fromDisposition) {
            return fromDisposition;
        }

        try {
            const pathname = new URL(url, location.href).pathname;
            const last = pathname.split('/').filter(Boolean).pop() || '';
            if (last && /\.[a-z0-9]+$/i.test(last)) {
                return last;
            }
        } catch {
            // ignore
        }

        return `${sanitizePathSegment(state.albumTitle || `JM${state.albumId}`)}.zip`;
    }

    async function handleArchiveResponse(response, fallbackUrl) {
        if (!looksLikeArchiveResponse(response, response.finalUrl || fallbackUrl)) {
            logStep('archive:not-detected', response.finalUrl || fallbackUrl || 'unknown');
            return false;
        }

        const finalUrl = response.finalUrl || fallbackUrl;
        const fileName = inferArchiveName(finalUrl, response.responseHeaders);
        const contentType = parseHeaders(response.responseHeaders)['content-type'] || 'application/octet-stream';
        logStep('archive:detected', `${fileName} | ${contentType}`);
        const blob = new Blob([response.response], { type: contentType });
        await saveBlobToConfiguredDirectory(blob, fileName);
        return true;
    }

    async function downloadArchiveFromUrl(url) {
        const finalUrl = toAbsoluteUrl(url);
        if (!finalUrl) {
            logStep('archive:url-invalid', url || 'empty');
            updateStatus('没拿到有效 ZIP 地址', 'error');
            return false;
        }

        logStep('archive:download-start', finalUrl);
        updateStatus('正在下载官方 ZIP...', 'info');

        const response = await gmRequest({
            method: 'GET',
            url: finalUrl,
            responseType: 'arraybuffer',
        });

        if (await handleArchiveResponse(response, finalUrl)) {
            return true;
        }

        const html = decodeArrayBuffer(response.response, parseHeaders(response.responseHeaders)['content-type']);
        const nestedArchiveUrl = extractArchiveUrlFromHtml(html, response.finalUrl || finalUrl);
        if (nestedArchiveUrl) {
            logStep('archive:nested-url', nestedArchiveUrl);
            const nestedResponse = await gmRequest({
                method: 'GET',
                url: nestedArchiveUrl,
                responseType: 'arraybuffer',
            });
            return handleArchiveResponse(nestedResponse, nestedArchiveUrl);
        }

        logStep('archive:not-found-in-response', finalUrl);
        updateStatus('没有在响应里识别出 ZIP 下载', 'warn');
        return false;
    }

    function findDownloadForm() {
        const exact = document.querySelector('form#album_down');
        if (exact instanceof HTMLFormElement) {
            return exact;
        }

        const fallback = Array.from(document.forms).find((form) => {
            const hasAlbumId = Boolean(form.querySelector('input[name="album_id"]'));
            const hasVerification = Boolean(form.querySelector('input[name="verification"], input#invite_verification'));
            const hasSubmitter = Boolean(form.querySelector('#download_submit, button[name="download_submit"], input[name="download_submit"]'));
            return hasAlbumId && hasVerification && hasSubmitter;
        });

        return fallback || null;
    }

    function findCaptchaInput() {
        const downloadForm = findDownloadForm();
        const exact = downloadForm?.querySelector('#invite_verification, input[name="verification"]');
        if (exact instanceof HTMLInputElement && !exact.disabled) {
            return exact;
        }

        const candidates = Array.from(document.querySelectorAll('input'))
            .filter((input) => input instanceof HTMLInputElement)
            .filter((input) => !input.disabled && input.type !== 'hidden')
            .filter((input) => {
                const placeholder = sanitizeText(input.placeholder || '');
                const nearbyText = sanitizeText(input.closest('form, div, td')?.textContent || '');
                return /数字|算|结果|验证码/i.test(placeholder) || /图片中的数字|可以下载|下载须知|计算/i.test(nearbyText);
            });

        return candidates[0] || null;
    }

    function focusCaptchaInput() {
        const input = findCaptchaInput();
        if (!input) {
            logStep('captcha:focus-miss');
            return false;
        }

        input.focus();
        input.select?.();
        logStep('captcha:focused', input.id || input.name || 'unknown');
        return true;
    }

    function getNodeText(node) {
        if (!node) {
            return '';
        }

        if (node instanceof HTMLInputElement) {
            return sanitizeText(node.value || node.placeholder || '');
        }

        return sanitizeText(node.textContent || '');
    }

    function isLikelyDirectDownloadTrigger(node, href, text) {
        const normalizedHref = String(href || '');
        const normalizedText = String(text || '');
        const selectorHints = node instanceof Element
            ? `${node.id ? `#${node.id}` : ''} ${node.getAttribute('onclick') || ''} ${node.getAttribute('href') || ''}`
            : '';

        return /免等待|直接下载|点我直接下载/.test(normalizedText) ||
            /click_fl1|click_fl2|shunt_modal_display|#shunt-modal/i.test(selectorHints) ||
            (/广告|sponsor/i.test(normalizedText) && Boolean(normalizedHref)) ||
            (/^https?:\/\//i.test(normalizedHref) && !normalizedHref.startsWith(location.origin) && /download|ad|sponsor|jump/i.test(normalizedHref));
    }

    function isLikelyDownloadForm(form) {
        if (!(form instanceof HTMLFormElement)) {
            return false;
        }

        const exactForm = findDownloadForm();
        if (exactForm && form === exactForm) {
            return true;
        }

        const submitters = Array.from(form.querySelectorAll('button, input[type="submit"]'));
        const hasDownloadSubmitter = submitters.some((node) => /下载|download/i.test(getNodeText(node)));
        const hasCaptcha = Boolean(findCaptchaInput()) || /图片中的数字|可以下载|下载须知|数学/i.test(sanitizeText(form.textContent || ''));

        return hasDownloadSubmitter && hasCaptcha;
    }

    function buildRequestFromForm(form, submitter) {
        const method = String(form.getAttribute('method') || 'GET').toUpperCase();
        const action = toAbsoluteUrl(form.getAttribute('action') || location.href);
        const enctype = String(form.enctype || 'application/x-www-form-urlencoded').toLowerCase();
        const formData = new FormData(form);

        if (submitter?.name && !formData.has(submitter.name)) {
            formData.append(submitter.name, submitter.value || '');
        }

        if (enctype.includes('multipart/form-data')) {
            return null;
        }

        if (method === 'GET') {
            const target = new URL(action);
            for (const [key, value] of formData.entries()) {
                if (typeof value === 'string') {
                    target.searchParams.append(key, value);
                }
            }

            return {
                method,
                url: target.toString(),
                headers: {},
            };
        }

        const params = new URLSearchParams();
        for (const [key, value] of formData.entries()) {
            if (typeof value === 'string') {
                params.append(key, value);
            }
        }

        return {
            method,
            url: action,
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
            },
            data: params.toString(),
        };
    }

    function getSafeSelectorId(value) {
        if (!value) {
            return '';
        }

        if (window.CSS?.escape) {
            return `#${window.CSS.escape(value)}`;
        }

        return `#${String(value).replace(/([ !"#$%&'()*+,./:;<=>?@[\\\]^`{|}~])/g, '\\$1')}`;
    }

    function findNativeSubmitTarget(form, submitter) {
        const candidates = [];

        if (submitter) {
            if (submitter.id) {
                candidates.push(getSafeSelectorId(submitter.id));
            }

            if (submitter.name) {
                candidates.push(`[name="${String(submitter.name).replace(/"/g, '\\"')}"]`);
            }
        }

        candidates.push('#download_submit');
        candidates.push('button[type="submit"]');
        candidates.push('input[type="submit"]');
        candidates.push('button');

        const matchesFormTarget = (node) => {
            if (!(node instanceof HTMLElement)) {
                return false;
            }

            if (node.form === form) {
                return true;
            }

            if (form.contains(node)) {
                return true;
            }

            return Boolean(form.id && node.getAttribute('form') === form.id);
        };

        for (const selector of candidates) {
            if (!selector) {
                continue;
            }

            const localNode = form.querySelector(selector);
            if (matchesFormTarget(localNode)) {
                return localNode;
            }

            const globalNodes = Array.from(document.querySelectorAll(selector));
            const globalNode = globalNodes.find((node) => matchesFormTarget(node));
            if (globalNode) {
                return globalNode;
            }
        }

        return null;
    }

    function fallbackToNativeSubmit(form, submitter, reason = 'fallback') {
        state.nativeSubmitBypass.add(form);
        logStep('submit:fallback-native', `${reason} | ${submitter?.name || submitter?.id || form.id || form.action || 'unknown'}`);

        if (reason === 'native-required') {
            updateStatus('该表单需要走浏览器原生提交，已切回原站流程', 'info');
        } else if (reason === 'native-browser') {
            // Preserve the folder-ready status message set by the caller.
        } else {
            updateStatus('接管失败，已回退原站下载流程', 'warn');
        }

        window.setTimeout(() => {
            try {
                const nativeTarget = findNativeSubmitTarget(form, submitter);
                if (nativeTarget) {
                    if ('disabled' in nativeTarget && nativeTarget.disabled) {
                        nativeTarget.disabled = false;
                    }
                    nativeTarget.removeAttribute?.('disabled');
                    nativeTarget.removeAttribute?.('aria-disabled');
                    logStep('submit:native-target', `${nativeTarget.tagName.toLowerCase()}#${nativeTarget.id || ''}.${nativeTarget.className || ''}`);
                    logStep('submit:native-click', nativeTarget.id || nativeTarget.getAttribute('name') || nativeTarget.tagName.toLowerCase());
                    nativeTarget.click();
                    return;
                }

                if (submitter?.name) {
                    const hidden = document.createElement('input');
                    hidden.type = 'hidden';
                    hidden.name = submitter.name;
                    hidden.value = submitter.value || '';
                    hidden.dataset.tmSubmitShim = '1';
                    form.appendChild(hidden);
                    logStep('submit:native-shim', `${hidden.name}=${hidden.value}`);
                    window.setTimeout(() => hidden.remove(), 3000);
                }

                logStep('submit:native-submit', form.id || form.action || 'unknown');
                HTMLFormElement.prototype.submit.call(form);
            } catch (error) {
                logStep('submit:native-error', error?.message || String(error));
                throw error;
            }
        }, 0);
    }

    async function interceptDownloadSubmit(form, submitter) {
        if (state.captureInProgress) {
            logStep('submit:busy');
            updateStatus('ZIP 接管中，请稍等', 'warn');
            return;
        }

        const exactForm = findDownloadForm();
        if (exactForm && form === exactForm) {
            const preparedFolder = await ensureAlbumFolderReady();
            if (preparedFolder) {
                updateStatus(`已创建文件夹：${preparedFolder.folderPath}，浏览器下载时把 ZIP 存进去`, 'ok');
            } else if (state.baseDirLabel) {
                updateStatus('没能提前建好本子文件夹，将继续浏览器原生下载', 'warn');
            } else {
                updateStatus('未设置保存目录，将继续浏览器原生下载', 'warn');
            }

            logStep('submit:native-browser-download', preparedFolder?.folderPath || 'folder-not-ready');
            fallbackToNativeSubmit(form, submitter, 'native-browser');
            return;
        }

        const request = buildRequestFromForm(form, submitter);
        if (!request) {
            logStep('submit:request-build-failed', form.id || form.action || 'unknown');
            updateStatus('表单编码不适合自动接管，已回退原站流程', 'warn');
            fallbackToNativeSubmit(form, submitter);
            return;
        }

        state.captureInProgress = true;
        logStep('submit:intercept-start', `${request.method} ${request.url}`);
        updateStatus('正在接管官方 ZIP 下载...', 'info');

        try {
            const response = await gmRequest({
                method: request.method,
                url: request.url,
                headers: request.headers,
                data: request.data,
                responseType: 'arraybuffer',
            });

            const saved = await handleArchiveResponse(response, request.url);
            if (saved) {
                return;
            }

            const headers = parseHeaders(response.responseHeaders);
            const html = decodeArrayBuffer(response.response, headers['content-type']);
            const archiveUrl = extractArchiveUrlFromHtml(html, response.finalUrl || request.url);

            if (archiveUrl) {
                logStep('submit:archive-url-found', archiveUrl);
                const downloaded = await downloadArchiveFromUrl(archiveUrl);
                if (downloaded) {
                    return;
                }
            }

            fallbackToNativeSubmit(form, submitter);
        } catch (error) {
            console.error('[18comic helper] submit interception failed', error);
            logStep('submit:intercept-error', error?.message || String(error));
            fallbackToNativeSubmit(form, submitter);
        } finally {
            state.captureInProgress = false;
            logStep('submit:intercept-finished');
        }
    }

    function scanPageState(reason = 'manual') {
        state.albumId = getAlbumId();
        state.albumTitle = getAlbumTitle();
        refreshPanel();

        const downloadForm = findDownloadForm();
        const captcha = findCaptchaInput();
        if (captcha && document.hasFocus()) {
            setTimeout(() => {
                if (document.activeElement !== captcha) {
                    focusCaptchaInput();
                }
            }, 80);
        }

        const archiveAnchor = Array.from(document.querySelectorAll('a[href]'))
            .map((anchor) => ({ anchor, href: toAbsoluteUrl(anchor.getAttribute('href')) }))
            .find(({ href, anchor }) => looksLikeArchiveUrl(href) || (/下载|download/i.test(getNodeText(anchor)) && /zip|rar|7z|download/i.test(href)));

        if (archiveAnchor) {
            updateStatus('页面上已经出现 ZIP 下载入口', 'ok');
            return;
        }

        if (downloadForm && captcha) {
            updateStatus('已找到官方下载表单，等待你手动答题', 'info');
        }
    }

    function installMutationObserver() {
        const observer = new MutationObserver((mutations) => {
            const hasExternalMutation = mutations.some((mutation) => {
                if (!isInsidePanel(mutation.target)) {
                    return true;
                }

                const nodes = [
                    ...Array.from(mutation.addedNodes || []),
                    ...Array.from(mutation.removedNodes || []),
                ];

                return nodes.some((node) => !isInsidePanel(node));
            });

            if (!hasExternalMutation) {
                return;
            }

            scheduleScan('mutation');
        });

        observer.observe(document.documentElement, {
            childList: true,
            subtree: true,
        });
    }

    function installWindowOpenGuard() {
        if (!originalWindowOpen) {
            logStep('guard:window-open-skip');
            return;
        }

        window.open = function (...args) {
            if (shouldBlockGuardedNavigation(args[0])) {
                logStep('guard:window-open-blocked', args[0]);
                updateStatus('已拦截广告弹窗，留在原下载页继续', 'ok');
                return null;
            }

            logStep('guard:window-open-pass', args[0] || 'empty');
            return originalWindowOpen(...args);
        };
    }

    function onClickCapture(event) {
        const target = event.target instanceof Element ? event.target.closest('a, button, input[type="button"], input[type="submit"]') : null;
        if (!target) {
            return;
        }

        const anchor = target.closest('a[href]');
        const href = toAbsoluteUrl(anchor?.getAttribute('href') || '');
        const text = getNodeText(target);
        logStep('click', `${target.tagName.toLowerCase()} | ${text || '[no text]'} | ${href || '[no href]'}`);

        if (anchor && looksLikeArchiveUrl(href) && state.autoCaptureEnabled) {
            event.preventDefault();
            logStep('click:archive-direct', href);
            downloadArchiveFromUrl(href).catch((error) => {
                console.error('[18comic helper] archive click interception failed', error);
                logStep('click:archive-direct-error', error?.message || String(error));
                updateStatus('ZIP 点击接管失败，已保留原页', 'warn');
            });
            return;
        }

        if (isLikelyDirectDownloadTrigger(target, href, text)) {
            state.adGuardUntil = Date.now() + 5000;
            logStep('click:direct-trigger', href || text || target.id || 'unknown');

            if (anchor && href && !href.startsWith(location.origin)) {
                event.preventDefault();
                logStep('click:blocked-external', href);
            }

            updateStatus('已接管直接下载点击，优先保留站内流程并拦外跳', 'ok');
            scheduleScan('direct-trigger');
            setTimeout(focusCaptchaInput, 300);
        }
    }

    function onSubmitCapture(event) {
        const form = event.target;
        if (!(form instanceof HTMLFormElement)) {
            return;
        }

        logStep('submit:event', `${form.id || '[no id]'} | auto=${state.autoCaptureEnabled ? 'on' : 'off'}`);

        if (state.nativeSubmitBypass.has(form)) {
            state.nativeSubmitBypass.delete(form);
            logStep('submit:native-bypass', form.id || form.action || 'unknown');
            return;
        }

        if (!state.autoCaptureEnabled || !isLikelyDownloadForm(form)) {
            logStep('submit:ignored', form.id || form.action || 'unknown');
            return;
        }

        event.preventDefault();
        interceptDownloadSubmit(form, event.submitter).catch((error) => {
            console.error('[18comic helper] form interception crashed', error);
            fallbackToNativeSubmit(form, event.submitter);
        });
    }

    function registerMenuCommands() {
        if (typeof GM_registerMenuCommand !== 'function') {
            return;
        }

        GM_registerMenuCommand('18comic: 选择保存目录', () => {
            chooseBaseDirectory();
        });

        GM_registerMenuCommand('18comic: 聚焦答题框', () => {
            focusCaptchaInput();
        });

        GM_registerMenuCommand('18comic: 复制调试日志', async () => {
            const ok = await copyText(state.debugLogs.join('\n'));
            updateStatus(ok ? '日志已复制到剪贴板' : '复制日志失败', ok ? 'ok' : 'warn');
        });

        GM_registerMenuCommand('18comic: 重新扫描下载页', () => {
            scanPageState('menu');
        });
    }

    async function init() {
        logStep('init:start', location.href);
        await wait(0);

        if (document.readyState === 'loading') {
            await new Promise((resolve) => document.addEventListener('DOMContentLoaded', resolve, { once: true }));
        }

        ensurePanel();
        await loadPersistedDirectoryHandle(false);
        refreshPanel();
        scanPageState('init');
        installMutationObserver();
        registerMenuCommands();

        if (state.baseDirHandle) {
            updateStatus(`已连接保存目录：${state.baseDirLabel || state.baseDirHandle.name}`, 'ok');
        } else {
            updateStatus('先点一次“选择保存目录”，后面就能按本子名自动建文件夹', 'info');
        }

        logStep('init:ready', state.albumId || 'unknown');
    }

    installWindowOpenGuard();
    document.addEventListener('click', onClickCapture, true);
    document.addEventListener('submit', onSubmitCapture, true);
    init().catch((error) => {
        console.error('[18comic helper] init failed', error);
    });
})();
