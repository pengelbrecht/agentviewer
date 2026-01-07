// agentviewer - Frontend Application

(function() {
    'use strict';

    // State
    let tabs = [];
    let activeTabId = null;
    let ws = null;

    // DOM Elements
    const tabsContainer = document.getElementById('tabs-container');
    const contentArea = document.getElementById('content');

    // Initialize
    function init() {
        connectWebSocket();
        loadTabs();
        setupKeyboardShortcuts();
    }

    // WebSocket Connection
    function connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        ws = new WebSocket(`${protocol}//${window.location.host}/ws`);

        ws.onopen = () => {
            console.log('WebSocket connected');
        };

        ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            handleWSMessage(msg);
        };

        ws.onclose = () => {
            console.log('WebSocket disconnected, reconnecting...');
            setTimeout(connectWebSocket, 2000);
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
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
                html = `<pre>${escapeHtml(tab.content)}</pre>`;
        }

        contentArea.innerHTML = html;
    }

    // Render markdown (basic implementation - will be enhanced with marked.js)
    function renderMarkdown(content) {
        // Basic markdown rendering - will use marked.js when vendor libs are added
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

    // Render code (basic implementation - will be enhanced with highlight.js)
    function renderCode(content, language) {
        // Basic code rendering - will use highlight.js when vendor libs are added
        const lines = escapeHtml(content).split('\n');
        return lines.map((line, i) =>
            `<div class="code-line"><span class="line-number">${i + 1}</span><span class="line-content">${line}</span></div>`
        ).join('');
    }

    // Render diff (basic implementation - will be enhanced with diff2html)
    function renderDiff(content) {
        // Basic diff rendering - will use diff2html when vendor libs are added
        const lines = escapeHtml(content).split('\n');
        return lines.map(line => {
            let className = 'diff-line';
            if (line.startsWith('+') && !line.startsWith('+++')) {
                className += ' diff-add';
            } else if (line.startsWith('-') && !line.startsWith('---')) {
                className += ' diff-remove';
            } else if (line.startsWith('@@')) {
                className += ' diff-hunk';
            }
            return `<div class="${className}">${line}</div>`;
        }).join('');
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
