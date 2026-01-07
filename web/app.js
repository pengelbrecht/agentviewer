// agentviewer - Frontend Application

(function() {
    'use strict';

    // State
    let tabs = [];
    let activeTabId = null;
    let ws = null;
    let reconnectAttempts = 0;
    const maxReconnectDelay = 30000; // 30 seconds max
    let closedTabsHistory = []; // Stack of closed tabs for reopen functionality
    const maxClosedTabs = 10; // Maximum number of closed tabs to remember

    // Search state
    let searchState = {
        isOpen: false,
        query: '',
        matches: [],
        currentIndex: -1,
        originalContent: null // Store original content for restoring
    };

    // DOM Elements
    const tabsContainer = document.getElementById('tabs-container');
    const contentArea = document.getElementById('content');

    // Initialize
    function init() {
        initTheme();
        initVendorLibs();
        connectWebSocket();
        loadTabs();
        setupKeyboardShortcuts();
        setupThemeToggle();
        setupSearch();
    }

    // Theme management
    const THEME_STORAGE_KEY = 'agentviewer-theme';

    function initTheme() {
        const savedTheme = localStorage.getItem(THEME_STORAGE_KEY);
        if (savedTheme) {
            document.documentElement.setAttribute('data-theme', savedTheme);
        }
        // If no saved theme, let CSS handle it via prefers-color-scheme
    }

    function getCurrentTheme() {
        const explicitTheme = document.documentElement.getAttribute('data-theme');
        if (explicitTheme) {
            return explicitTheme;
        }
        // No explicit theme set, use system preference
        return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
    }

    function setTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem(THEME_STORAGE_KEY, theme);
        updateMermaidTheme(theme);
    }

    function toggleTheme() {
        const current = getCurrentTheme();
        const newTheme = current === 'dark' ? 'light' : 'dark';
        setTheme(newTheme);
    }

    function setupThemeToggle() {
        const toggleBtn = document.getElementById('theme-toggle');
        if (toggleBtn) {
            toggleBtn.addEventListener('click', toggleTheme);
        }

        // Listen for system theme changes
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
            // Only update if no explicit theme is set
            if (!localStorage.getItem(THEME_STORAGE_KEY)) {
                updateMermaidTheme(e.matches ? 'dark' : 'light');
            }
        });
    }

    function updateMermaidTheme(theme) {
        if (typeof mermaid !== 'undefined') {
            mermaid.initialize({
                startOnLoad: false,
                theme: theme === 'dark' ? 'dark' : 'default',
                securityLevel: 'loose'
            });
        }
    }

    // Initialize vendor libraries
    function initVendorLibs() {
        // Initialize mermaid with current theme
        updateMermaidTheme(getCurrentTheme());
    }

    // WebSocket Connection
    function connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        ws = new WebSocket(`${protocol}//${window.location.host}/ws`);

        ws.onopen = () => {
            console.log('WebSocket connected');
            reconnectAttempts = 0;
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
                // Save closed tab to history for reopen (only if not already saved locally)
                // This handles tabs deleted via external API calls
                const deletedTab = tabs.find(t => t.id === msg.id);
                if (deletedTab && !closedTabsHistory.some(t => t.id === deletedTab.id && Date.now() - t.closedAt < 1000)) {
                    saveClosedTab(deletedTab);
                }
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
                html = `<div class="content-diff">${renderDiff(tab.content, tab)}</div>`;
                break;

            case 'mermaid':
                html = `<div class="content-mermaid">${renderMermaid(tab.content)}</div>`;
                break;

            case 'image':
                html = `<div class="content-image">${renderImage(tab.content, tab.title)}</div>`;
                break;

            case 'csv':
                html = `<div class="content-csv">${renderCSV(tab.content, tab.title)}</div>`;
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
            // Highlight code blocks that weren't highlighted during render (fallback)
            if (typeof hljs !== 'undefined') {
                contentArea.querySelectorAll('pre code:not(.hljs)').forEach((block) => {
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

        if (type === 'code') {
            // Setup copy button handlers
            setupCopyButtons();
        }
    }

    // Setup copy button click handlers
    function setupCopyButtons() {
        contentArea.querySelectorAll('.code-copy-btn').forEach((btn) => {
            btn.addEventListener('click', async (e) => {
                e.preventDefault();
                const encodedContent = btn.dataset.code;
                if (!encodedContent) return;

                try {
                    // Decode the content
                    const content = decodeURIComponent(atob(encodedContent));

                    // Copy to clipboard
                    await navigator.clipboard.writeText(content);

                    // Show success state
                    const copyIcon = btn.querySelector('.copy-icon');
                    const copiedIcon = btn.querySelector('.copied-icon');

                    copyIcon.style.display = 'none';
                    copiedIcon.style.display = 'inline';
                    btn.classList.add('copied');

                    // Reset after 2 seconds
                    setTimeout(() => {
                        copyIcon.style.display = 'inline';
                        copiedIcon.style.display = 'none';
                        btn.classList.remove('copied');
                    }, 2000);
                } catch (err) {
                    console.error('Failed to copy code:', err);
                    // Fallback for older browsers
                    fallbackCopyToClipboard(decodeURIComponent(atob(encodedContent)));
                }
            });
        });
    }

    // Fallback copy method for browsers without clipboard API
    function fallbackCopyToClipboard(text) {
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.position = 'fixed';
        textarea.style.left = '-9999px';
        document.body.appendChild(textarea);
        textarea.select();
        try {
            document.execCommand('copy');
        } catch (err) {
            console.error('Fallback copy failed:', err);
        }
        document.body.removeChild(textarea);
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

    // Configure marked.js with highlight.js integration
    function configureMarked() {
        if (typeof marked === 'undefined') return;

        // Use marked.use() for extensions (works with v15+)
        marked.use({
            // Enable GFM
            gfm: true,
            breaks: false,
            pedantic: false,

            // Custom renderer using the new token-based API
            useNewRenderer: true,
            renderer: {
                // Handle code blocks with highlight.js
                code(token) {
                    const code = token.text || '';
                    const lang = token.lang || '';

                    // Mermaid diagrams
                    if (lang === 'mermaid') {
                        return `<div class="mermaid">${escapeHtml(code)}</div>`;
                    }

                    // Syntax highlighting with highlight.js
                    let highlighted = escapeHtml(code);
                    if (typeof hljs !== 'undefined') {
                        try {
                            if (lang && hljs.getLanguage(lang)) {
                                highlighted = hljs.highlight(code, { language: lang }).value;
                            } else if (code.trim()) {
                                // Auto-detect language for non-empty code
                                highlighted = hljs.highlightAuto(code).value;
                            }
                        } catch (e) {
                            console.error('Highlight.js error:', e);
                            highlighted = escapeHtml(code);
                        }
                    }

                    const langClass = lang ? `language-${escapeHtml(lang)}` : '';
                    return `<pre><code class="hljs ${langClass}">${highlighted}</code></pre>`;
                }
            }
        });
    }

    // Extract and protect math expressions before markdown processing
    // Returns { content: processed content, mathMap: map of placeholders to math }
    function extractMathExpressions(content) {
        const mathMap = {};
        let counter = 0;

        // Don't process if no $ signs present
        if (!content.includes('$')) {
            return { content: content, mathMap: mathMap };
        }

        // Extract display math: $$...$$ (can be multiline)
        content = content.replace(/\$\$([\s\S]+?)\$\$/g, function(match, math) {
            const placeholder = `%%%MATH_DISPLAY_${counter}%%%`;
            mathMap[placeholder] = { math: math.trim(), display: true };
            counter++;
            return placeholder;
        });

        // Extract inline math: $...$ (not greedy, no newlines)
        // Use negative lookbehind/lookahead to avoid matching $$
        content = content.replace(/(?<!\$)\$(?!\$)([^\$\n]+?)\$(?!\$)/g, function(match, math) {
            // Skip if math content has spaces on both ends (likely not math)
            if (/^\s|\s$/.test(math)) {
                return match;
            }
            const placeholder = `%%%MATH_INLINE_${counter}%%%`;
            mathMap[placeholder] = { math: math, display: false };
            counter++;
            return placeholder;
        });

        return { content: content, mathMap: mathMap };
    }

    // Restore math expressions as KaTeX-ready spans
    function restoreMathExpressions(html, mathMap) {
        for (const [placeholder, data] of Object.entries(mathMap)) {
            const className = data.display ? 'katex-display' : 'katex-inline';
            const escaped = escapeHtml(data.math);
            html = html.replace(placeholder, `<span class="${className}">${escaped}</span>`);
        }
        return html;
    }

    // Render markdown using marked.js
    function renderMarkdown(content) {
        // Extract math expressions before markdown processing
        const { content: processedContent, mathMap } = extractMathExpressions(content);

        let html;
        if (typeof marked !== 'undefined') {
            configureMarked();
            html = marked.parse(processedContent);
        } else {
            // Fallback: basic markdown rendering
            html = escapeHtml(processedContent);

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
        }

        // Restore math expressions as KaTeX-ready spans
        html = restoreMathExpressions(html, mathMap);

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
        const languageLabel = language ? escapeHtml(language) : '';

        // Store the raw content for copy functionality using data attribute
        const encodedContent = btoa(encodeURIComponent(content));

        return `<div class="code-header">
            <span class="code-language">${languageLabel}</span>
            <button class="code-copy-btn" data-code="${encodedContent}" title="Copy code">
                <span class="copy-icon">Copy</span>
                <span class="copied-icon" style="display:none">Copied!</span>
            </button>
        </div>
        <table class="code-table"><tbody>${
            highlightedLines.map((line, i) =>
                `<tr class="code-line"><td class="line-number">${i + 1}</td><td class="line-content">${line || ' '}</td></tr>`
            ).join('')
        }</tbody></table>`;
    }

    // Render standalone mermaid diagram (for .mmd/.mermaid files)
    function renderMermaid(content) {
        // Generate a unique ID for the mermaid diagram
        const containerId = 'mermaid-standalone-' + Date.now();

        // Schedule rendering after the element is in the DOM
        setTimeout(() => {
            const container = document.getElementById(containerId);
            if (!container) return;

            if (typeof mermaid !== 'undefined') {
                mermaid.render(`mermaid-render-${Date.now()}`, content).then(result => {
                    container.innerHTML = result.svg;
                    container.classList.add('mermaid-rendered');
                }).catch(err => {
                    console.error('Mermaid render error:', err);
                    container.innerHTML = `<div class="mermaid-error">
                        <h3>Mermaid Diagram Error</h3>
                        <p>${escapeHtml(err.message || String(err))}</p>
                        <pre>${escapeHtml(content)}</pre>
                    </div>`;
                });
            } else {
                // Mermaid library not available, show raw content
                container.innerHTML = `<div class="mermaid-fallback">
                    <p>Mermaid library not loaded</p>
                    <pre>${escapeHtml(content)}</pre>
                </div>`;
            }
        }, 0);

        return `<div id="${containerId}" class="mermaid-container"></div>`;
    }

    // Parse CSV content into a 2D array
    // Handles quoted fields, escaped quotes, and various line endings
    function parseCSV(content) {
        const rows = [];
        let currentRow = [];
        let currentField = '';
        let inQuotes = false;
        let i = 0;

        while (i < content.length) {
            const char = content[i];
            const nextChar = content[i + 1];

            if (inQuotes) {
                if (char === '"') {
                    if (nextChar === '"') {
                        // Escaped quote ("") -> single quote
                        currentField += '"';
                        i += 2;
                        continue;
                    } else {
                        // End of quoted field
                        inQuotes = false;
                        i++;
                        continue;
                    }
                } else {
                    currentField += char;
                    i++;
                    continue;
                }
            } else {
                if (char === '"' && currentField === '') {
                    // Start of quoted field
                    inQuotes = true;
                    i++;
                    continue;
                } else if (char === ',') {
                    // End of field
                    currentRow.push(currentField);
                    currentField = '';
                    i++;
                    continue;
                } else if (char === '\r' && nextChar === '\n') {
                    // CRLF line ending
                    currentRow.push(currentField);
                    rows.push(currentRow);
                    currentRow = [];
                    currentField = '';
                    i += 2;
                    continue;
                } else if (char === '\n' || char === '\r') {
                    // LF or CR line ending
                    currentRow.push(currentField);
                    rows.push(currentRow);
                    currentRow = [];
                    currentField = '';
                    i++;
                    continue;
                } else {
                    currentField += char;
                    i++;
                    continue;
                }
            }
        }

        // Handle last field/row
        if (currentField !== '' || currentRow.length > 0) {
            currentRow.push(currentField);
            rows.push(currentRow);
        }

        return rows;
    }

    // Render CSV content as an interactive table
    function renderCSV(content, title) {
        if (!content || content.trim() === '') {
            return `<div class="csv-error">
                <h3>No CSV Content</h3>
                <p>The CSV content is empty or unavailable.</p>
            </div>`;
        }

        const rows = parseCSV(content);
        if (rows.length === 0) {
            return `<div class="csv-error">
                <h3>Empty CSV</h3>
                <p>The CSV file contains no data.</p>
            </div>`;
        }

        // Assume first row is headers
        const headers = rows[0];
        const dataRows = rows.slice(1);

        // Generate unique ID for the table
        const tableId = 'csv-table-' + Date.now();

        // Build header row with sortable columns
        const headerHtml = headers.map((h, i) =>
            `<th class="csv-header" data-col="${i}" data-sort-dir="none">
                <span class="csv-header-text">${escapeHtml(h)}</span>
                <span class="csv-sort-icon">⇅</span>
            </th>`
        ).join('');

        // Build data rows
        const bodyHtml = dataRows.map((row, rowIndex) => {
            const cells = headers.map((_, colIndex) => {
                const value = row[colIndex] !== undefined ? row[colIndex] : '';
                return `<td class="csv-cell" data-col="${colIndex}">${escapeHtml(value)}</td>`;
            }).join('');
            return `<tr class="csv-row" data-row="${rowIndex}">${cells}</tr>`;
        }).join('');

        // Schedule interactive setup after DOM update
        setTimeout(() => {
            setupCSVTable(tableId, rows);
        }, 0);

        return `<div class="csv-container">
            <div class="csv-toolbar">
                <input type="text" class="csv-search" placeholder="Search..." data-table="${tableId}" />
                <span class="csv-row-count">${dataRows.length} row${dataRows.length !== 1 ? 's' : ''}</span>
            </div>
            <div class="csv-table-wrapper">
                <table id="${tableId}" class="csv-table">
                    <thead><tr>${headerHtml}</tr></thead>
                    <tbody>${bodyHtml}</tbody>
                </table>
            </div>
        </div>`;
    }

    // Setup interactive features for CSV table
    function setupCSVTable(tableId, originalRows) {
        const table = document.getElementById(tableId);
        if (!table) return;

        const container = table.closest('.csv-container');
        const searchInput = container.querySelector('.csv-search');
        const rowCountEl = container.querySelector('.csv-row-count');
        const headers = originalRows[0];
        let dataRows = originalRows.slice(1);
        let sortCol = -1;
        let sortDir = 'none'; // 'none', 'asc', 'desc'

        // Setup column sorting
        table.querySelectorAll('.csv-header').forEach(th => {
            th.addEventListener('click', () => {
                const col = parseInt(th.dataset.col, 10);

                // Update sort direction
                if (sortCol === col) {
                    sortDir = sortDir === 'none' ? 'asc' : (sortDir === 'asc' ? 'desc' : 'none');
                } else {
                    sortCol = col;
                    sortDir = 'asc';
                }

                // Update header UI
                table.querySelectorAll('.csv-header').forEach(h => {
                    h.dataset.sortDir = 'none';
                    h.querySelector('.csv-sort-icon').textContent = '⇅';
                });
                th.dataset.sortDir = sortDir;
                th.querySelector('.csv-sort-icon').textContent =
                    sortDir === 'asc' ? '↑' : (sortDir === 'desc' ? '↓' : '⇅');

                // Sort and re-render
                renderTableBody();
            });
        });

        // Setup search filtering
        searchInput.addEventListener('input', () => {
            renderTableBody();
        });

        function renderTableBody() {
            const query = searchInput.value.toLowerCase().trim();
            let rows = [...dataRows];

            // Filter
            if (query) {
                rows = rows.filter(row =>
                    row.some(cell => (cell || '').toLowerCase().includes(query))
                );
            }

            // Sort
            if (sortDir !== 'none' && sortCol >= 0) {
                rows.sort((a, b) => {
                    const valA = (a[sortCol] || '').toLowerCase();
                    const valB = (b[sortCol] || '').toLowerCase();

                    // Try numeric sort first
                    const numA = parseFloat(valA);
                    const numB = parseFloat(valB);
                    if (!isNaN(numA) && !isNaN(numB)) {
                        return sortDir === 'asc' ? numA - numB : numB - numA;
                    }

                    // Fall back to string sort
                    const cmp = valA.localeCompare(valB);
                    return sortDir === 'asc' ? cmp : -cmp;
                });
            }

            // Render
            const tbody = table.querySelector('tbody');
            tbody.innerHTML = rows.map((row, rowIndex) => {
                const cells = headers.map((_, colIndex) => {
                    const value = row[colIndex] !== undefined ? row[colIndex] : '';
                    // Highlight matching text if searching
                    let displayValue = escapeHtml(value);
                    if (query && value.toLowerCase().includes(query)) {
                        const regex = new RegExp(`(${escapeRegExp(query)})`, 'gi');
                        displayValue = displayValue.replace(regex, '<mark class="csv-match">$1</mark>');
                    }
                    return `<td class="csv-cell" data-col="${colIndex}">${displayValue}</td>`;
                }).join('');
                return `<tr class="csv-row" data-row="${rowIndex}">${cells}</tr>`;
            }).join('');

            // Update row count
            rowCountEl.textContent = `${rows.length} of ${dataRows.length} row${dataRows.length !== 1 ? 's' : ''}`;
            if (rows.length === dataRows.length) {
                rowCountEl.textContent = `${dataRows.length} row${dataRows.length !== 1 ? 's' : ''}`;
            }
        }
    }

    // Escape regex special characters
    function escapeRegExp(string) {
        return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    }

    // Render image content (expects a data URL or URL string)
    function renderImage(content, title) {
        // Handle empty content
        if (!content) {
            return `<div class="image-error">
                <h3>No Image Content</h3>
                <p>The image content is empty or unavailable.</p>
            </div>`;
        }

        // Create the image element with appropriate alt text
        const altText = title ? escapeHtml(title) : 'Image';

        // Check if content is a data URL or a regular URL
        const isDataUrl = content.startsWith('data:');
        const isUrl = content.startsWith('http://') || content.startsWith('https://') || content.startsWith('/');

        if (!isDataUrl && !isUrl) {
            // Content might be raw base64, try to wrap it
            // Attempt to detect image type from magic bytes or default to png
            return `<div class="image-error">
                <h3>Invalid Image Format</h3>
                <p>The image content is not in a recognized format (expected data URL or URL).</p>
            </div>`;
        }

        return `<figure class="image-figure">
            <img src="${content}" alt="${altText}" class="image-display" loading="lazy" />
            ${title ? `<figcaption class="image-caption">${escapeHtml(title)}</figcaption>` : ''}
        </figure>`;
    }

    // Render diff using diff2html-ui for side-by-side view with syntax highlighting
    function renderDiff(content, tab) {
        // Generate a unique container ID for this diff
        const containerId = 'diff-container-' + Date.now();
        const navId = 'diff-nav-' + Date.now();

        // Use Diff2HtmlUI if available (from diff2html-ui bundle)
        if (typeof Diff2HtmlUI !== 'undefined') {
            // Return a container element that we'll populate after it's in the DOM
            // Schedule the diff rendering after this element is added to DOM
            setTimeout(() => {
                const container = document.getElementById(containerId);
                if (!container) return;

                try {
                    // Ensure the diff content has proper format
                    const normalizedContent = normalizeUnifiedDiff(content);

                    // Detect language from tab metadata or file extension
                    const language = detectDiffLanguage(tab);

                    // Create Diff2HtmlUI instance
                    const diff2htmlUi = new Diff2HtmlUI(container, normalizedContent, {
                        drawFileList: false,
                        matching: 'lines',
                        outputFormat: 'side-by-side',
                        renderNothingWhenEmpty: false,
                        colorScheme: 'auto',
                        highlight: true
                    });

                    // Draw the diff
                    diff2htmlUi.draw();

                    // Apply syntax highlighting
                    diff2htmlUi.highlightCode();

                    // Setup collapsible sections for unchanged code
                    setupCollapsibleSections(container);

                    // Setup navigation after diff is rendered
                    setupDiffNavigation(navId, container);

                } catch (e) {
                    console.error('Diff2HtmlUI error:', e);
                    // Fallback on error - pass language for highlighting
                    const lang = detectDiffLanguage(tab);
                    container.innerHTML = renderSideBySideFallback(content, lang);
                    setupCollapsibleSections(container);
                    setupDiffNavigation(navId, container);
                }
            }, 0);

            return renderDiffNavBar(navId) + `<div id="${containerId}" class="diff2html-wrapper"></div>`;
        }

        // Fallback to Diff2Html.html (non-UI version) if available
        if (typeof Diff2Html !== 'undefined') {
            try {
                const normalizedContent = normalizeUnifiedDiff(content);
                const diffHtml = Diff2Html.html(normalizedContent, {
                    drawFileList: false,
                    matching: 'lines',
                    outputFormat: 'side-by-side',
                    renderNothingWhenEmpty: false,
                    colorScheme: 'auto'
                });
                // Setup collapsible sections and navigation after DOM is updated
                setTimeout(() => {
                    const navEl = document.getElementById(navId);
                    if (navEl) {
                        const diffContainer = navEl.parentElement.querySelector('.diff2html-wrapper, .diff-side-by-side');
                        setupCollapsibleSections(diffContainer);
                        setupDiffNavigation(navId, diffContainer);
                    }
                }, 0);
                return renderDiffNavBar(navId) + `<div class="diff2html-wrapper">${diffHtml}</div>`;
            } catch (e) {
                console.error('Diff2Html error:', e);
            }
        }

        // Fallback: custom side-by-side diff rendering with syntax highlighting
        const lang = detectDiffLanguage(tab);
        const fallbackHtml = renderSideBySideFallback(content, lang);
        // Setup collapsible sections and navigation after DOM is updated
        setTimeout(() => {
            const navEl = document.getElementById(navId);
            if (navEl) {
                const diffContainer = navEl.parentElement.querySelector('.diff-side-by-side');
                setupCollapsibleSections(diffContainer);
                setupDiffNavigation(navId, diffContainer);
            }
        }, 0);
        return renderDiffNavBar(navId) + fallbackHtml;
    }

    // Configuration for collapsible unchanged sections
    const COLLAPSE_CONFIG = {
        minUnchangedToCollapse: 8, // Minimum unchanged lines before collapsing
        contextLines: 3            // Lines to show above/below changes
    };

    // Render the diff navigation bar HTML
    function renderDiffNavBar(navId) {
        return `<div id="${navId}" class="diff-nav">
            <div class="diff-nav-info">
                <span class="diff-nav-count">0 changes</span>
                <span class="diff-nav-current"></span>
            </div>
            <div class="diff-nav-buttons">
                <button class="diff-nav-btn" data-action="prev" title="Previous change (↑)">
                    <span>↑</span> Prev
                </button>
                <button class="diff-nav-btn" data-action="next" title="Next change (↓)">
                    <span>↓</span> Next
                </button>
            </div>
        </div>`;
    }

    // Setup diff navigation functionality
    function setupDiffNavigation(navId, diffContainer) {
        const navEl = document.getElementById(navId);
        if (!navEl || !diffContainer) return;

        // Find all change rows (additions and deletions)
        // For diff2html: .d2h-ins, .d2h-del rows
        // For fallback: .diff-add, .diff-delete rows
        const changeRows = diffContainer.querySelectorAll(
            '.d2h-ins, .d2h-del, tr:has(.diff-add), tr:has(.diff-delete)'
        );

        // Group consecutive changes into hunks for navigation
        const changeHunks = [];
        let currentHunk = [];
        let lastRow = null;

        changeRows.forEach((row) => {
            // Check if this row is consecutive to the last one
            if (lastRow && lastRow.nextElementSibling === row) {
                currentHunk.push(row);
            } else {
                if (currentHunk.length > 0) {
                    changeHunks.push(currentHunk);
                }
                currentHunk = [row];
            }
            lastRow = row;
        });
        if (currentHunk.length > 0) {
            changeHunks.push(currentHunk);
        }

        // Update count display
        const countEl = navEl.querySelector('.diff-nav-count');
        const currentEl = navEl.querySelector('.diff-nav-current');
        const totalChanges = changeHunks.length;

        if (countEl) {
            countEl.textContent = `${totalChanges} ${totalChanges === 1 ? 'change' : 'changes'}`;
        }

        // Track current position
        let currentIndex = -1;

        function updateCurrentDisplay() {
            if (currentEl) {
                if (currentIndex >= 0 && totalChanges > 0) {
                    currentEl.textContent = `(${currentIndex + 1}/${totalChanges})`;
                } else {
                    currentEl.textContent = '';
                }
            }
        }

        function highlightCurrentChange() {
            // Remove previous highlight
            diffContainer.querySelectorAll('.diff-nav-highlight').forEach(el => {
                el.classList.remove('diff-nav-highlight');
            });

            // Add highlight to current hunk
            if (currentIndex >= 0 && currentIndex < changeHunks.length) {
                changeHunks[currentIndex].forEach(row => {
                    row.classList.add('diff-nav-highlight');
                });
            }
        }

        function navigateTo(index) {
            if (totalChanges === 0) return;

            // Clamp index
            currentIndex = Math.max(0, Math.min(index, totalChanges - 1));

            // Scroll to the change
            const targetRow = changeHunks[currentIndex][0];
            if (targetRow) {
                targetRow.scrollIntoView({ behavior: 'smooth', block: 'center' });
            }

            updateCurrentDisplay();
            highlightCurrentChange();
        }

        function navigateNext() {
            if (totalChanges === 0) return;
            navigateTo(currentIndex + 1 >= totalChanges ? 0 : currentIndex + 1);
        }

        function navigatePrev() {
            if (totalChanges === 0) return;
            navigateTo(currentIndex - 1 < 0 ? totalChanges - 1 : currentIndex - 1);
        }

        // Setup button handlers
        navEl.querySelectorAll('.diff-nav-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const action = btn.dataset.action;
                if (action === 'next') navigateNext();
                else if (action === 'prev') navigatePrev();
            });
        });

        // Keyboard navigation within the diff area
        diffContainer.setAttribute('tabindex', '0');
        diffContainer.addEventListener('keydown', (e) => {
            if (e.key === 'ArrowDown' || e.key === 'j') {
                e.preventDefault();
                navigateNext();
            } else if (e.key === 'ArrowUp' || e.key === 'k') {
                e.preventDefault();
                navigatePrev();
            }
        });

        // Initialize display
        updateCurrentDisplay();
    }

    // Setup collapsible unchanged sections in diff view
    function setupCollapsibleSections(diffContainer) {
        if (!diffContainer) return;

        // Find all rows in the diff table
        const tbody = diffContainer.querySelector('.d2h-diff-tbody, .diff-table tbody');
        if (!tbody) return;

        const rows = Array.from(tbody.querySelectorAll('tr'));
        if (rows.length === 0) return;

        // Identify unchanged (context) rows vs changed rows
        // diff2html: .d2h-cntx or rows without .d2h-ins/.d2h-del
        // fallback: .diff-context or rows without .diff-add/.diff-delete
        function isUnchangedRow(row) {
            // diff2html context lines
            if (row.classList.contains('d2h-cntx')) return true;
            // fallback context lines
            if (row.querySelector('.diff-context')) return true;
            // If it's not an ins/del row and not a hunk header, it's context
            if (!row.classList.contains('d2h-ins') &&
                !row.classList.contains('d2h-del') &&
                !row.querySelector('.diff-add') &&
                !row.querySelector('.diff-delete') &&
                !row.querySelector('.diff-hunk') &&
                !row.classList.contains('d2h-info')) {
                return true;
            }
            return false;
        }

        function isHunkHeader(row) {
            return row.classList.contains('d2h-info') ||
                   row.querySelector('.diff-hunk') !== null;
        }

        // Group consecutive unchanged rows
        const groups = [];
        let currentGroup = { type: 'mixed', rows: [] };

        for (const row of rows) {
            if (isHunkHeader(row)) {
                // Hunk headers break groups
                if (currentGroup.rows.length > 0) {
                    groups.push(currentGroup);
                }
                groups.push({ type: 'hunk', rows: [row] });
                currentGroup = { type: 'mixed', rows: [] };
            } else if (isUnchangedRow(row)) {
                if (currentGroup.type === 'unchanged') {
                    currentGroup.rows.push(row);
                } else {
                    if (currentGroup.rows.length > 0) {
                        groups.push(currentGroup);
                    }
                    currentGroup = { type: 'unchanged', rows: [row] };
                }
            } else {
                if (currentGroup.type === 'changed' || currentGroup.type === 'mixed') {
                    currentGroup.rows.push(row);
                    currentGroup.type = 'changed';
                } else {
                    if (currentGroup.rows.length > 0) {
                        groups.push(currentGroup);
                    }
                    currentGroup = { type: 'changed', rows: [row] };
                }
            }
        }
        if (currentGroup.rows.length > 0) {
            groups.push(currentGroup);
        }

        // Process unchanged groups to determine what to collapse
        const { minUnchangedToCollapse, contextLines } = COLLAPSE_CONFIG;

        for (let i = 0; i < groups.length; i++) {
            const group = groups[i];
            if (group.type !== 'unchanged') continue;

            const unchangedCount = group.rows.length;

            // Only collapse if we have enough unchanged lines
            if (unchangedCount <= minUnchangedToCollapse) continue;

            // Determine how many context lines to keep at top/bottom
            const prevGroup = groups[i - 1];
            const nextGroup = groups[i + 1];

            // Keep context lines at top if there's a preceding changed group
            const keepTop = (prevGroup && (prevGroup.type === 'changed' || prevGroup.type === 'hunk'))
                ? contextLines : 0;

            // Keep context lines at bottom if there's a following changed group
            const keepBottom = (nextGroup && (nextGroup.type === 'changed' || nextGroup.type === 'hunk'))
                ? contextLines : 0;

            const toCollapse = unchangedCount - keepTop - keepBottom;

            if (toCollapse <= 2) continue; // Not worth collapsing

            // Create collapsed section
            const startIdx = keepTop;
            const endIdx = unchangedCount - keepBottom;
            const collapsedRows = group.rows.slice(startIdx, endIdx);
            const hiddenCount = collapsedRows.length;

            // Hide the collapsed rows
            collapsedRows.forEach(row => {
                row.classList.add('diff-collapsed-row');
                row.style.display = 'none';
            });

            // Create expand button row
            const expandRow = document.createElement('tr');
            expandRow.className = 'diff-expand-row';

            // Determine the number of columns (side-by-side has multiple)
            const colCount = group.rows[0].querySelectorAll('td').length || 5;

            const expandTd = document.createElement('td');
            expandTd.colSpan = colCount;
            expandTd.className = 'diff-expand-cell';
            expandTd.innerHTML = `
                <button class="diff-expand-btn" data-expanded="false">
                    <span class="diff-expand-icon">▸</span>
                    <span class="diff-expand-text">${hiddenCount} lines hidden</span>
                </button>
            `;

            expandRow.appendChild(expandTd);

            // Insert expand row after the last kept top row (or at start of group)
            const insertAfter = keepTop > 0 ? group.rows[keepTop - 1] : null;
            if (insertAfter) {
                insertAfter.parentNode.insertBefore(expandRow, insertAfter.nextSibling);
            } else if (group.rows[0]) {
                group.rows[0].parentNode.insertBefore(expandRow, group.rows[0]);
            }

            // Store reference to collapsed rows on the expand row
            expandRow._collapsedRows = collapsedRows;

            // Setup click handler
            const expandBtn = expandRow.querySelector('.diff-expand-btn');
            expandBtn.addEventListener('click', () => {
                const isExpanded = expandBtn.dataset.expanded === 'true';
                const icon = expandBtn.querySelector('.diff-expand-icon');
                const text = expandBtn.querySelector('.diff-expand-text');

                if (isExpanded) {
                    // Collapse
                    expandRow._collapsedRows.forEach(row => {
                        row.style.display = 'none';
                    });
                    expandBtn.dataset.expanded = 'false';
                    icon.textContent = '▸';
                    text.textContent = `${hiddenCount} lines hidden`;
                } else {
                    // Expand
                    expandRow._collapsedRows.forEach(row => {
                        row.style.display = '';
                    });
                    expandBtn.dataset.expanded = 'true';
                    icon.textContent = '▾';
                    text.textContent = `${hiddenCount} lines shown`;
                }
            });
        }
    }

    // Detect programming language from tab metadata or diff content
    function detectDiffLanguage(tab) {
        // Check if language is explicitly set in tab metadata
        if (tab && tab.language) {
            return tab.language;
        }

        // Try to extract from diff file headers
        if (tab && tab.content) {
            const lines = tab.content.split('\n');
            for (const line of lines) {
                // Look for +++ b/path/to/file.ext pattern
                const match = line.match(/^\+\+\+ [ab]?\/?(.*)/);
                if (match && match[1]) {
                    const filePath = match[1];
                    const ext = filePath.split('.').pop();
                    return getLanguageFromExtension(ext);
                }
            }
        }

        return null;
    }

    // Map file extensions to highlight.js language names
    function getLanguageFromExtension(ext) {
        const extMap = {
            'js': 'javascript',
            'jsx': 'javascript',
            'ts': 'typescript',
            'tsx': 'typescript',
            'py': 'python',
            'rb': 'ruby',
            'go': 'go',
            'rs': 'rust',
            'java': 'java',
            'kt': 'kotlin',
            'swift': 'swift',
            'c': 'c',
            'cpp': 'cpp',
            'cc': 'cpp',
            'h': 'c',
            'hpp': 'cpp',
            'cs': 'csharp',
            'php': 'php',
            'html': 'html',
            'htm': 'html',
            'css': 'css',
            'scss': 'scss',
            'sass': 'sass',
            'less': 'less',
            'json': 'json',
            'xml': 'xml',
            'yaml': 'yaml',
            'yml': 'yaml',
            'md': 'markdown',
            'sql': 'sql',
            'sh': 'bash',
            'bash': 'bash',
            'zsh': 'bash',
            'ps1': 'powershell',
            'dockerfile': 'dockerfile',
            'makefile': 'makefile',
            'lua': 'lua',
            'r': 'r',
            'scala': 'scala',
            'perl': 'perl',
            'pl': 'perl'
        };

        return extMap[ext?.toLowerCase()] || null;
    }

    // Normalize unified diff format for diff2html
    function normalizeUnifiedDiff(content) {
        // If content doesn't start with diff markers, wrap it
        const lines = content.split('\n');

        // Check if it's already a proper unified diff
        const hasFileHeaders = lines.some(l => l.startsWith('---')) &&
                              lines.some(l => l.startsWith('+++'));
        const hasHunkHeaders = lines.some(l => l.startsWith('@@'));

        if (hasFileHeaders && hasHunkHeaders) {
            return content;
        }

        // If we only have hunk content (no headers), add minimal headers
        if (hasHunkHeaders && !hasFileHeaders) {
            return '--- a/original\n+++ b/modified\n' + content;
        }

        // If it's raw diff lines without any headers, try to wrap them
        if (lines.some(l => l.startsWith('+') || l.startsWith('-') || l.startsWith(' '))) {
            // Count the lines to estimate ranges
            let oldCount = 0, newCount = 0;
            for (const line of lines) {
                if (line.startsWith('-') && !line.startsWith('---')) oldCount++;
                else if (line.startsWith('+') && !line.startsWith('+++')) newCount++;
                else if (line.startsWith(' ')) { oldCount++; newCount++; }
            }
            return `--- a/original\n+++ b/modified\n@@ -1,${oldCount || 1} +1,${newCount || 1} @@\n${content}`;
        }

        return content;
    }

    // Fallback side-by-side diff rendering when diff2html fails
    // Accepts optional language parameter for syntax highlighting
    function renderSideBySideFallback(content, language) {
        const lines = content.split('\n');
        const leftLines = [];
        const rightLines = [];
        let leftNum = 1;
        let rightNum = 1;
        let detectedLang = language;

        for (const line of lines) {
            // Skip file headers but extract language hint
            if (line.startsWith('---') || line.startsWith('+++') || line.startsWith('diff ')) {
                // Try to detect language from file path if not already set
                if (!detectedLang) {
                    const match = line.match(/^\+\+\+ [ab]?\/?(.*)/);
                    if (match && match[1]) {
                        const ext = match[1].split('.').pop();
                        detectedLang = getLanguageFromExtension(ext);
                    }
                }
                continue;
            }

            // Skip hunk headers but use them to reset line numbers
            if (line.startsWith('@@')) {
                const match = line.match(/@@ -(\d+),?\d* \+(\d+),?\d* @@/);
                if (match) {
                    leftNum = parseInt(match[1], 10);
                    rightNum = parseInt(match[2], 10);
                }
                // Add hunk header row
                leftLines.push({ type: 'hunk', content: line, num: null });
                rightLines.push({ type: 'hunk', content: '', num: null });
                continue;
            }

            if (line.startsWith('-')) {
                leftLines.push({ type: 'delete', content: line.substring(1), num: leftNum++ });
                rightLines.push({ type: 'empty', content: '', num: null });
            } else if (line.startsWith('+')) {
                leftLines.push({ type: 'empty', content: '', num: null });
                rightLines.push({ type: 'add', content: line.substring(1), num: rightNum++ });
            } else if (line.startsWith(' ') || line === '') {
                const actualLine = line.startsWith(' ') ? line.substring(1) : '';
                leftLines.push({ type: 'context', content: actualLine, num: leftNum++ });
                rightLines.push({ type: 'context', content: actualLine, num: rightNum++ });
            }
        }

        // Apply syntax highlighting to content if highlight.js is available
        function highlightLine(text, lang) {
            if (!text || text.trim() === '') return escapeHtml(text);
            if (typeof hljs === 'undefined') return escapeHtml(text);

            try {
                if (lang && hljs.getLanguage(lang)) {
                    return hljs.highlight(text, { language: lang }).value;
                } else if (text.trim()) {
                    // Auto-detect language
                    return hljs.highlightAuto(text).value;
                }
            } catch (e) {
                console.error('Highlight error in diff fallback:', e);
            }
            return escapeHtml(text);
        }

        // Build the table
        const maxRows = Math.max(leftLines.length, rightLines.length);
        let rows = '';

        for (let i = 0; i < maxRows; i++) {
            const left = leftLines[i] || { type: 'empty', content: '', num: null };
            const right = rightLines[i] || { type: 'empty', content: '', num: null };

            const leftClass = `diff-line diff-${left.type}`;
            const rightClass = `diff-line diff-${right.type}`;
            const leftNumStr = left.num !== null ? left.num : '';
            const rightNumStr = right.num !== null ? right.num : '';

            // Apply syntax highlighting to code content (not hunks or empty lines)
            const leftContent = left.type === 'hunk' ? escapeHtml(left.content) : highlightLine(left.content, detectedLang);
            const rightContent = right.type === 'hunk' ? escapeHtml(right.content) : highlightLine(right.content, detectedLang);

            rows += `<tr>
                <td class="diff-line-num ${leftClass}">${leftNumStr}</td>
                <td class="${leftClass}">${leftContent}</td>
                <td class="diff-gutter"></td>
                <td class="diff-line-num ${rightClass}">${rightNumStr}</td>
                <td class="${rightClass}">${rightContent}</td>
            </tr>`;
        }

        return `<div class="diff-side-by-side">
            <div class="diff-header">
                <div class="diff-header-left">Original</div>
                <div class="diff-header-right">Modified</div>
            </div>
            <table class="diff-table"><tbody>${rows}</tbody></table>
        </div>`;
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
        // Find and save the tab before closing (local save in case WS message is slow)
        const tabToClose = tabs.find(t => t.id === id);
        if (tabToClose) {
            saveClosedTab(tabToClose);
        }

        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'close_tab', id: id }));
        }
    }

    // Save a closed tab to history for reopen functionality
    function saveClosedTab(tab) {
        // Create a copy of the tab to preserve its state
        const tabCopy = {
            id: tab.id,
            title: tab.title,
            type: tab.type,
            content: tab.content,
            language: tab.language,
            closedAt: Date.now()
        };

        // Add to history, limit size
        closedTabsHistory.push(tabCopy);
        if (closedTabsHistory.length > maxClosedTabs) {
            closedTabsHistory.shift(); // Remove oldest
        }
    }

    // Reopen the most recently closed tab
    async function reopenClosedTab() {
        if (closedTabsHistory.length === 0) {
            return; // No tabs to reopen
        }

        const tab = closedTabsHistory.pop();

        // Send create request to server
        try {
            const response = await fetch('/api/tabs', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    id: tab.id + '-' + Date.now(), // New unique ID to avoid conflicts
                    title: tab.title,
                    type: tab.type,
                    content: tab.content,
                    language: tab.language
                })
            });

            if (!response.ok) {
                // Put the tab back if failed
                closedTabsHistory.push(tab);
                console.error('Failed to reopen tab:', response.statusText);
            }
        } catch (error) {
            // Put the tab back if failed
            closedTabsHistory.push(tab);
            console.error('Failed to reopen tab:', error);
        }
    }

    // ========== Search functionality ==========

    // Setup search bar and event handlers
    function setupSearch() {
        const searchBar = document.getElementById('search-bar');
        const searchInput = document.getElementById('search-input');
        const searchCount = document.getElementById('search-count');
        const searchPrev = document.getElementById('search-prev');
        const searchNext = document.getElementById('search-next');
        const searchClose = document.getElementById('search-close');

        if (!searchBar || !searchInput) return;

        // Input handler for searching
        searchInput.addEventListener('input', (e) => {
            const query = e.target.value;
            searchState.query = query;
            if (query.length > 0) {
                performSearch(query);
            } else {
                clearSearchHighlights();
                updateSearchCount(0, 0);
            }
        });

        // Keyboard shortcuts within search input
        searchInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                if (e.shiftKey) {
                    navigateSearchPrev();
                } else {
                    navigateSearchNext();
                }
            } else if (e.key === 'Escape') {
                e.preventDefault();
                closeSearch();
            }
        });

        // Navigation buttons
        searchPrev.addEventListener('click', navigateSearchPrev);
        searchNext.addEventListener('click', navigateSearchNext);
        searchClose.addEventListener('click', closeSearch);
    }

    // Open the search bar
    function openSearch() {
        const searchBar = document.getElementById('search-bar');
        const searchInput = document.getElementById('search-input');

        if (!searchBar || !searchInput) return;

        searchBar.classList.remove('hidden');
        searchState.isOpen = true;

        // Focus and select existing text
        searchInput.focus();
        searchInput.select();

        // If there's already a query, re-perform search
        if (searchState.query) {
            performSearch(searchState.query);
        }
    }

    // Close the search bar
    function closeSearch() {
        const searchBar = document.getElementById('search-bar');

        if (!searchBar) return;

        searchBar.classList.add('hidden');
        searchState.isOpen = false;

        // Clear highlights when closing
        clearSearchHighlights();
    }

    // Perform search and highlight matches
    function performSearch(query) {
        if (!query || query.length === 0) {
            clearSearchHighlights();
            return;
        }

        // Clear previous highlights
        clearSearchHighlights();

        // Find all text nodes in the content area
        const contentArea = document.getElementById('content');
        if (!contentArea) return;

        // Create a TreeWalker to find text nodes
        const walker = document.createTreeWalker(
            contentArea,
            NodeFilter.SHOW_TEXT,
            {
                acceptNode: function(node) {
                    // Skip script and style tags
                    const parent = node.parentNode;
                    if (parent.tagName === 'SCRIPT' || parent.tagName === 'STYLE') {
                        return NodeFilter.FILTER_REJECT;
                    }
                    // Skip empty text nodes
                    if (node.textContent.trim().length === 0) {
                        return NodeFilter.FILTER_REJECT;
                    }
                    return NodeFilter.FILTER_ACCEPT;
                }
            }
        );

        const textNodes = [];
        let node;
        while (node = walker.nextNode()) {
            textNodes.push(node);
        }

        // Search and highlight within text nodes
        const matches = [];
        const queryLower = query.toLowerCase();

        textNodes.forEach(textNode => {
            const text = textNode.textContent;
            const textLower = text.toLowerCase();
            let startIndex = 0;
            let foundIndex;

            // Find all occurrences in this text node
            const nodeMatches = [];
            while ((foundIndex = textLower.indexOf(queryLower, startIndex)) !== -1) {
                nodeMatches.push({
                    textNode: textNode,
                    start: foundIndex,
                    end: foundIndex + query.length
                });
                startIndex = foundIndex + 1;
            }

            if (nodeMatches.length > 0) {
                // Process matches in reverse order to avoid index shifting issues
                nodeMatches.reverse().forEach(match => {
                    matches.unshift(highlightMatch(match.textNode, match.start, match.end));
                });
            }
        });

        // Store matches and update state
        searchState.matches = matches;
        searchState.currentIndex = matches.length > 0 ? 0 : -1;

        // Update count display
        updateSearchCount(searchState.currentIndex + 1, matches.length);

        // Highlight current match
        if (matches.length > 0) {
            highlightCurrentMatch();
        }
    }

    // Highlight a single match and return the highlight element
    function highlightMatch(textNode, start, end) {
        const text = textNode.textContent;
        const before = text.substring(0, start);
        const match = text.substring(start, end);
        const after = text.substring(end);

        // Create highlight span
        const highlightSpan = document.createElement('span');
        highlightSpan.className = 'search-highlight';
        highlightSpan.textContent = match;

        // Create document fragment with before text, highlight, and after text
        const fragment = document.createDocumentFragment();
        if (before) {
            fragment.appendChild(document.createTextNode(before));
        }
        fragment.appendChild(highlightSpan);
        if (after) {
            fragment.appendChild(document.createTextNode(after));
        }

        // Replace the text node with the fragment
        textNode.parentNode.replaceChild(fragment, textNode);

        return highlightSpan;
    }

    // Update the current match highlight
    function highlightCurrentMatch() {
        // Remove current class from all highlights
        document.querySelectorAll('.search-highlight.current').forEach(el => {
            el.classList.remove('current');
        });

        // Add current class to current match
        if (searchState.currentIndex >= 0 && searchState.currentIndex < searchState.matches.length) {
            const currentMatch = searchState.matches[searchState.currentIndex];
            if (currentMatch) {
                currentMatch.classList.add('current');
                // Scroll into view
                currentMatch.scrollIntoView({ behavior: 'smooth', block: 'center' });
            }
        }

        // Update count
        updateSearchCount(searchState.currentIndex + 1, searchState.matches.length);
    }

    // Navigate to next match
    function navigateSearchNext() {
        if (searchState.matches.length === 0) return;

        searchState.currentIndex = (searchState.currentIndex + 1) % searchState.matches.length;
        highlightCurrentMatch();
    }

    // Navigate to previous match
    function navigateSearchPrev() {
        if (searchState.matches.length === 0) return;

        searchState.currentIndex = searchState.currentIndex - 1;
        if (searchState.currentIndex < 0) {
            searchState.currentIndex = searchState.matches.length - 1;
        }
        highlightCurrentMatch();
    }

    // Update the search count display
    function updateSearchCount(current, total) {
        const searchCount = document.getElementById('search-count');
        if (!searchCount) return;

        if (total === 0) {
            if (searchState.query && searchState.query.length > 0) {
                searchCount.textContent = 'No results';
                searchCount.classList.add('no-results');
            } else {
                searchCount.textContent = '';
                searchCount.classList.remove('no-results');
            }
        } else {
            searchCount.textContent = `${current} of ${total}`;
            searchCount.classList.remove('no-results');
        }
    }

    // Clear all search highlights
    function clearSearchHighlights() {
        searchState.matches = [];
        searchState.currentIndex = -1;

        // Find all highlight spans and replace them with their text content
        const highlights = document.querySelectorAll('.search-highlight');
        highlights.forEach(highlight => {
            const textNode = document.createTextNode(highlight.textContent);
            highlight.parentNode.replaceChild(textNode, highlight);
        });

        // Normalize text nodes (merge adjacent text nodes)
        const contentArea = document.getElementById('content');
        if (contentArea) {
            contentArea.normalize();
        }

        updateSearchCount(0, 0);
    }

    // Setup keyboard shortcuts
    function setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            const isMod = e.metaKey || e.ctrlKey;

            // Don't trigger shortcuts when typing in search input
            if (e.target.id === 'search-input') return;

            // Cmd/Ctrl + F or '/' to open search
            if ((isMod && e.key === 'f') || (!isMod && e.key === '/' && e.target.tagName !== 'INPUT' && e.target.tagName !== 'TEXTAREA')) {
                e.preventDefault();
                openSearch();
                return;
            }

            // Escape to close search (when not in search input)
            if (e.key === 'Escape' && searchState.isOpen) {
                e.preventDefault();
                closeSearch();
                return;
            }

            // Cmd/Ctrl + Shift + T to reopen closed tab
            if (isMod && e.shiftKey && (e.key === 't' || e.key === 'T')) {
                e.preventDefault();
                reopenClosedTab();
                return;
            }

            // Cmd/Ctrl + 1-9 to switch tabs
            if (isMod && e.key >= '1' && e.key <= '9') {
                e.preventDefault();
                const idx = parseInt(e.key) - 1;
                if (idx < tabs.length) {
                    activateTab(tabs[idx].id);
                }
                return;
            }

            // Cmd/Ctrl + W to close tab
            if (isMod && e.key === 'w') {
                e.preventDefault();
                if (activeTabId) {
                    closeTab(activeTabId);
                }
                return;
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
