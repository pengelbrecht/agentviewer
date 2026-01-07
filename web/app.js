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
                        highlight: true,
                        rawTemplates: {
                            'side-by-side-file-diff': `
                                <div id="{{fileHtmlId}}" class="d2h-file-wrapper" data-lang="{{file.language}}">
                                    <div class="d2h-file-diff">
                                        <div class="d2h-code-wrapper">
                                            <table class="d2h-diff-table">
                                                <tbody class="d2h-diff-tbody">
                                                {{{diffs}}}
                                                </tbody>
                                            </table>
                                        </div>
                                    </div>
                                </div>`
                        }
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

    // Setup keyboard shortcuts
    function setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            const isMod = e.metaKey || e.ctrlKey;

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
