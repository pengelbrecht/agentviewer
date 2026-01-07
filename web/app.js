// agentviewer - Frontend Application

(function() {
    'use strict';

    // State
    let tabs = [];
    let activeTabId = null;
    let ws = null;
    let reconnectAttempts = 0;
    const maxReconnectDelay = 30000; // 30 seconds max

    // DOM Elements
    const tabsContainer = document.getElementById('tabs-container');
    const contentArea = document.getElementById('content');

    // Initialize
    function init() {
        initVendorLibs();
        connectWebSocket();
        loadTabs();
        setupKeyboardShortcuts();
    }

    // Initialize vendor libraries
    function initVendorLibs() {
        // Initialize mermaid
        if (typeof mermaid !== 'undefined') {
            mermaid.initialize({
                startOnLoad: false,
                theme: 'dark',
                securityLevel: 'loose'
            });
        }
    }

    // WebSocket Connection
    function connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        ws = new WebSocket(`${protocol}//${window.location.host}/ws`);

        ws.onopen = () => {
            console.log('WebSocket connected');
            reconnectAttempts = 0;
            updateConnectionStatus(true);
        };

        ws.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                handleWSMessage(msg);
            } catch (e) {
                console.error('Failed to parse WebSocket message:', e);
            }
        };

        ws.onclose = () => {
            console.log('WebSocket disconnected');
            updateConnectionStatus(false);
            scheduleReconnect();
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    // Schedule reconnection with exponential backoff
    function scheduleReconnect() {
        reconnectAttempts++;
        // Exponential backoff: 1s, 2s, 4s, 8s, ... up to maxReconnectDelay
        const delay = Math.min(1000 * Math.pow(2, reconnectAttempts - 1), maxReconnectDelay);
        console.log(`Reconnecting in ${delay}ms (attempt ${reconnectAttempts})...`);
        setTimeout(connectWebSocket, delay);
    }

    // Update connection status indicator
    function updateConnectionStatus(connected) {
        let indicator = document.getElementById('connection-status');
        if (!indicator) {
            indicator = document.createElement('div');
            indicator.id = 'connection-status';
            document.body.appendChild(indicator);
        }

        indicator.className = connected ? 'connected' : 'disconnected';
        indicator.title = connected ? 'Connected' : 'Disconnected - reconnecting...';

        // Hide indicator after a short delay when connected
        if (connected) {
            setTimeout(() => {
                indicator.classList.add('fade-out');
            }, 2000);
        } else {
            indicator.classList.remove('fade-out');
        }
    }

    // Handle WebSocket messages
    function handleWSMessage(msg) {
        switch (msg.type) {
            case 'tab_created':
                tabs.push(msg.tab);
                renderTabs();
                activateTab(msg.tab.id);
                break;

            case 'tab_updated':
                const idx = tabs.findIndex(t => t.id === msg.tab.id);
                if (idx !== -1) {
                    tabs[idx] = msg.tab;
                    if (activeTabId === msg.tab.id) {
                        renderContent(msg.tab);
                    }
                    renderTabs();
                }
                break;

            case 'tab_deleted':
                tabs = tabs.filter(t => t.id !== msg.id);
                if (activeTabId === msg.id) {
                    activeTabId = tabs.length > 0 ? tabs[0].id : null;
                }
                renderTabs();
                renderActiveContent();
                break;

            case 'tab_activated':
                activeTabId = msg.id;
                renderTabs();
                renderActiveContent();
                break;

            case 'tabs_cleared':
                tabs = [];
                activeTabId = null;
                renderTabs();
                renderActiveContent();
                break;
        }
    }

    // Load initial tabs
    async function loadTabs() {
        try {
            const response = await fetch('/api/tabs');
            const data = await response.json();
            tabs = data.tabs || [];

            // Find active tab
            const active = tabs.find(t => t.active);
            activeTabId = active ? active.id : (tabs.length > 0 ? tabs[0].id : null);

            renderTabs();
            renderActiveContent();
        } catch (error) {
            console.error('Failed to load tabs:', error);
        }
    }

    // Render tab bar
    function renderTabs() {
        tabsContainer.innerHTML = tabs.map(tab => `
            <div class="tab ${tab.id === activeTabId ? 'active' : ''}" data-id="${tab.id}">
                <span class="tab-title">${escapeHtml(tab.title || 'Untitled')}</span>
                <span class="tab-close" data-close="${tab.id}">&times;</span>
            </div>
        `).join('');

        // Add click handlers
        tabsContainer.querySelectorAll('.tab').forEach(el => {
            el.addEventListener('click', (e) => {
                if (e.target.classList.contains('tab-close')) {
                    closeTab(e.target.dataset.close);
                } else {
                    activateTab(el.dataset.id);
                }
            });

            // Middle-click to close tab
            el.addEventListener('mousedown', (e) => {
                if (e.button === 1) { // Middle mouse button
                    e.preventDefault();
                    closeTab(el.dataset.id);
                }
            });

            // Prevent middle-click from triggering browser behavior
            el.addEventListener('auxclick', (e) => {
                if (e.button === 1) {
                    e.preventDefault();
                }
            });
        });
    }

    // Render active content
    async function renderActiveContent() {
        if (!activeTabId) {
            contentArea.innerHTML = `
                <div class="empty-state">
                    <h2>No tabs open</h2>
                    <p>Waiting for content from AI agent...</p>
                </div>
            `;
            return;
        }

        try {
            const response = await fetch(`/api/tabs/${activeTabId}`);
            const tab = await response.json();
            renderContent(tab);
        } catch (error) {
            console.error('Failed to load tab content:', error);
        }
    }

    // Render tab content
    function renderContent(tab) {
        let html = '';

        switch (tab.type) {
            case 'markdown':
                html = `<div class="content-markdown">${renderMarkdown(tab.content)}</div>`;
                break;

            case 'code':
                html = `<div class="content-code">${renderCode(tab.content, tab.language)}</div>`;
                break;

            case 'diff':
                html = `<div class="content-diff">${renderDiff(tab.content)}</div>`;
                break;

            default:
                html = `<pre class="content-plain">${escapeHtml(tab.content)}</pre>`;
        }

        contentArea.innerHTML = html;

        // Post-render hooks
        postRenderContent(tab.type);
    }

    // Post-render processing for content types
    function postRenderContent(type) {
        if (type === 'markdown') {
            // Highlight code blocks in markdown
            if (typeof hljs !== 'undefined') {
                contentArea.querySelectorAll('pre code').forEach((block) => {
                    hljs.highlightElement(block);
                });
            }

            // Render mermaid diagrams
            if (typeof mermaid !== 'undefined') {
                contentArea.querySelectorAll('.mermaid').forEach((el, i) => {
                    const code = el.textContent;
                    el.setAttribute('id', `mermaid-${i}-${Date.now()}`);
                    mermaid.render(`mermaid-graph-${i}-${Date.now()}`, code).then(result => {
                        el.innerHTML = result.svg;
                    }).catch(err => {
                        console.error('Mermaid render error:', err);
                        el.innerHTML = `<pre class="mermaid-error">Mermaid error: ${escapeHtml(err.message || String(err))}</pre>`;
                    });
                });
            }

            // Render KaTeX math
            if (typeof katex !== 'undefined') {
                renderKatexInContent();
            }
        }
    }

    // Render KaTeX math expressions
    function renderKatexInContent() {
        // Render display math: $$...$$
        contentArea.querySelectorAll('.katex-display').forEach((el) => {
            try {
                katex.render(el.textContent, el, { displayMode: true, throwOnError: false });
            } catch (e) {
                console.error('KaTeX display error:', e);
            }
        });

        // Render inline math: $...$
        contentArea.querySelectorAll('.katex-inline').forEach((el) => {
            try {
                katex.render(el.textContent, el, { displayMode: false, throwOnError: false });
            } catch (e) {
                console.error('KaTeX inline error:', e);
            }
        });
    }

    // Configure marked.js
    function configureMarked() {
        if (typeof marked === 'undefined') return;

        // Custom renderer for mermaid and math
        const renderer = new marked.Renderer();

        // Handle code blocks
        renderer.code = function(code, language) {
            // Mermaid diagrams
            if (language === 'mermaid') {
                return `<div class="mermaid">${escapeHtml(code)}</div>`;
            }

            // Regular code blocks - highlight.js will process them post-render
            const langClass = language ? `language-${escapeHtml(language)}` : '';
            return `<pre><code class="${langClass}">${escapeHtml(code)}</code></pre>`;
        };

        // Handle math in paragraphs and text
        const originalParagraph = renderer.paragraph.bind(renderer);
        renderer.paragraph = function(text) {
            // Replace display math: $$...$$
            text = text.replace(/\$\$([^$]+)\$\$/g, '<span class="katex-display">$1</span>');
            // Replace inline math: $...$
            text = text.replace(/\$([^$\n]+)\$/g, '<span class="katex-inline">$1</span>');
            return originalParagraph(text);
        };

        marked.setOptions({
            renderer: renderer,
            gfm: true,
            breaks: false,
            pedantic: false
        });
    }

    // Render markdown using marked.js
    function renderMarkdown(content) {
        if (typeof marked !== 'undefined') {
            configureMarked();
            return marked.parse(content);
        }

        // Fallback: basic markdown rendering
        let html = escapeHtml(content);

        // Headers
        html = html.replace(/^### (.+)$/gm, '<h3>$1</h3>');
        html = html.replace(/^## (.+)$/gm, '<h2>$1</h2>');
        html = html.replace(/^# (.+)$/gm, '<h1>$1</h1>');

        // Bold and italic
        html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
        html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');

        // Code blocks
        html = html.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code class="language-$1">$2</code></pre>');
        html = html.replace(/`([^`]+)`/g, '<code>$1</code>');

        // Line breaks
        html = html.replace(/\n\n/g, '</p><p>');
        html = '<p>' + html + '</p>';

        return html;
    }

    // Render code using highlight.js
    function renderCode(content, language) {
        const lines = content.split('\n');
        let highlightedCode = escapeHtml(content);

        // Use highlight.js if available
        if (typeof hljs !== 'undefined') {
            try {
                if (language && hljs.getLanguage(language)) {
                    highlightedCode = hljs.highlight(content, { language: language }).value;
                } else {
                    highlightedCode = hljs.highlightAuto(content).value;
                }
            } catch (e) {
                console.error('Highlight error:', e);
                highlightedCode = escapeHtml(content);
            }
        }

        // Wrap with line numbers
        const highlightedLines = highlightedCode.split('\n');
        return `<table class="code-table"><tbody>${
            highlightedLines.map((line, i) =>
                `<tr class="code-line"><td class="line-number">${i + 1}</td><td class="line-content">${line || ' '}</td></tr>`
            ).join('')
        }</tbody></table>`;
    }

    // Render diff using diff2html
    function renderDiff(content) {
        // Use diff2html if available
        if (typeof Diff2Html !== 'undefined') {
            try {
                return Diff2Html.html(content, {
                    drawFileList: false,
                    matching: 'lines',
                    outputFormat: 'side-by-side',
                    renderNothingWhenEmpty: false
                });
            } catch (e) {
                console.error('Diff2Html error:', e);
            }
        }

        // Fallback: basic diff rendering
        const lines = escapeHtml(content).split('\n');
        return `<div class="diff-fallback">${lines.map(line => {
            let className = 'diff-line';
            if (line.startsWith('+') && !line.startsWith('+++')) {
                className += ' diff-add';
            } else if (line.startsWith('-') && !line.startsWith('---')) {
                className += ' diff-remove';
            } else if (line.startsWith('@@')) {
                className += ' diff-hunk';
            }
            return `<div class="${className}">${line}</div>`;
        }).join('')}</div>`;
    }

    // Activate a tab
    function activateTab(id) {
        activeTabId = id;
        renderTabs();
        renderActiveContent();

        // Notify server
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'activate_tab', id: id }));
        }
    }

    // Close a tab
    function closeTab(id) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'close_tab', id: id }));
        }
    }

    // Setup keyboard shortcuts
    function setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            const isMod = e.metaKey || e.ctrlKey;

            // Cmd/Ctrl + 1-9 to switch tabs
            if (isMod && e.key >= '1' && e.key <= '9') {
                e.preventDefault();
                const idx = parseInt(e.key) - 1;
                if (idx < tabs.length) {
                    activateTab(tabs[idx].id);
                }
            }

            // Cmd/Ctrl + W to close tab
            if (isMod && e.key === 'w') {
                e.preventDefault();
                if (activeTabId) {
                    closeTab(activeTabId);
                }
            }
        });
    }

    // Escape HTML
    function escapeHtml(str) {
        if (!str) return '';
        return str
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    }

    // Start app when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
