package server

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN" id="htmlRoot">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Vault Reader</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500&family=Noto+Sans+SC:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
        /* ==================== Theme: Light (default) ==================== */
        :root, [data-theme="light"] {
            --bg: #ffffff;
            --bg-secondary: #f8f9fa;
            --sidebar-bg: #fafbfc;
            --sidebar-hover: #f0f1f3;
            --sidebar-active: #e8f0fe;
            --text: #1a1a2e;
            --text-secondary: #5f6368;
            --text-muted: #9aa0a6;
            --border: #e8eaed;
            --border-light: #f1f3f4;
            --link: #1a73e8;
            --link-hover: #1558b0;
            --link-bg: rgba(26,115,232,0.08);
            --broken-link: #d93025;
            --code-bg: #f1f3f4;
            --accent: #1a73e8;
            --accent-soft: rgba(26,115,232,0.1);
            --shadow-sm: 0 1px 2px rgba(0,0,0,0.04);
            --shadow-md: 0 2px 8px rgba(0,0,0,0.08);
            --shadow-lg: 0 8px 24px rgba(0,0,0,0.12);
            --mark-bg: rgba(255,235,59,0.4);
        }

        /* ==================== Theme: Dark ==================== */
        [data-theme="dark"] {
            --bg: #0f0f0f;
            --bg-secondary: #161616;
            --sidebar-bg: #141414;
            --sidebar-hover: #1e1e1e;
            --sidebar-active: rgba(138,180,248,0.15);
            --text: #e8eaed;
            --text-secondary: #9aa0a6;
            --text-muted: #5f6368;
            --border: #2a2a2a;
            --border-light: #222222;
            --link: #8ab4f8;
            --link-hover: #aecbfa;
            --link-bg: rgba(138,180,248,0.1);
            --broken-link: #f28b82;
            --code-bg: #1a1a1a;
            --accent: #8ab4f8;
            --accent-soft: rgba(138,180,248,0.12);
            --shadow-sm: 0 1px 2px rgba(0,0,0,0.2);
            --shadow-md: 0 2px 8px rgba(0,0,0,0.3);
            --shadow-lg: 0 8px 24px rgba(0,0,0,0.4);
            --mark-bg: rgba(255,235,59,0.25);
        }

        /* Common */
        --radius-sm: 6px;
        --radius-md: 8px;
        --radius-lg: 12px;
        --font-sans: 'Inter', 'Noto Sans SC', -apple-system, BlinkMacSystemFont, sans-serif;
        --font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
        --header-height: 52px;
        --sidebar-width: 280px;
        --right-width: 260px;
        --transition: 150ms cubic-bezier(0.4, 0, 0.2, 1);

        /* ==================== Reset & Base ==================== */
        *, *::before, *::after { margin: 0; padding: 0; box-sizing: border-box; }
        html { font-size: 15px; -webkit-font-smoothing: antialiased; -moz-osx-font-smoothing: grayscale; }
        body {
            font-family: var(--font-sans);
            background: var(--bg);
            color: var(--text);
            display: flex;
            flex-direction: column;
            height: 100vh;
            overflow: hidden;
            transition: background 0.25s, color 0.25s;
        }

        /* ==================== Scrollbar ==================== */
        ::-webkit-scrollbar { width: 6px; height: 6px; }
        ::-webkit-scrollbar-track { background: transparent; }
        ::-webkit-scrollbar-thumb { background: var(--border); border-radius: 3px; }
        ::-webkit-scrollbar-thumb:hover { background: var(--text-muted); }

        /* ==================== Header ==================== */
        .header {
            display: flex;
            align-items: center;
            height: var(--header-height);
            padding: 0 20px;
            border-bottom: 1px solid var(--border);
            background: var(--bg);
            gap: 16px;
            flex-shrink: 0;
            z-index: 50;
            transition: background 0.25s, border-color 0.25s;
        }
        .logo {
            display: flex;
            align-items: center;
            gap: 10px;
            flex-shrink: 0;
        }
        .logo-icon {
            width: 28px; height: 28px;
            background: var(--accent);
            border-radius: 6px;
            display: flex; align-items: center; justify-content: center;
        }
        .logo-icon svg { width: 16px; height: 16px; fill: white; }
        .logo-text { font-size: 15px; font-weight: 600; letter-spacing: -0.3px; }
        .search-container { flex: 1; max-width: 520px; position: relative; }
        .search-box {
            width: 100%; padding: 8px 14px 8px 38px;
            border: 1px solid var(--border); border-radius: var(--radius-md);
            background: var(--bg-secondary); color: var(--text);
            font-size: 14px; font-family: var(--font-sans);
            transition: all var(--transition); outline: none;
        }
        .search-box:focus { border-color: var(--accent); background: var(--bg); box-shadow: 0 0 0 3px var(--accent-soft); }
        .search-box::placeholder { color: var(--text-muted); }
        .search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: var(--text-muted); pointer-events: none; }
        .search-icon svg { width: 16px; height: 16px; }
        .shortcut-hint { position: absolute; right: 10px; top: 50%; transform: translateY(-50%); padding: 2px 6px; border: 1px solid var(--border); border-radius: 4px; font-size: 11px; color: var(--text-muted); background: var(--bg); pointer-events: none; }

        /* Theme toggle */
        .theme-toggle {
            width: 36px; height: 36px; border: none; background: transparent;
            border-radius: var(--radius-sm); cursor: pointer; color: var(--text-muted);
            display: flex; align-items: center; justify-content: center;
            transition: all var(--transition); flex-shrink: 0;
        }
        .theme-toggle:hover { background: var(--sidebar-hover); color: var(--text); }
        .theme-toggle svg { width: 20px; height: 20px; }
        [data-theme="dark"] .icon-sun { display: none; }
        [data-theme="dark"] .icon-moon { display: block; }
        [data-theme="light"] .icon-sun { display: block; }
        [data-theme="light"] .icon-moon { display: none; }
        .icon-moon { display: none; }

        /* ==================== Main Layout ==================== */
        .main { display: flex; flex: 1; overflow: hidden; }

        /* ==================== Left Sidebar ==================== */
        .sidebar {
            width: var(--sidebar-width);
            border-right: 1px solid var(--border);
            overflow-y: auto; padding: 12px 0;
            background: var(--sidebar-bg); flex-shrink: 0;
            position: relative;
            transition: background 0.25s, border-color 0.25s;
        }
        .sidebar-resizer {
            position: absolute; top: 0; right: -3px; width: 6px; height: 100%;
            cursor: col-resize; z-index: 10;
        }
        .sidebar-resizer:hover, .sidebar-resizer.active { background: var(--accent); opacity: 0.3; }
        .sidebar-header { padding: 0 16px 10px; font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.8px; color: var(--text-muted); }

        /* Tree */
        .tree-item { list-style: none; }
        .tree-dir {
            display: flex; align-items: center; gap: 6px;
            padding: 5px 16px; cursor: pointer;
            font-size: 13.5px; font-weight: 500; color: var(--text);
            transition: background var(--transition); user-select: none;
        }
        .tree-dir:hover { background: var(--sidebar-hover); }
        .tree-dir .chevron { width: 16px; height: 16px; flex-shrink: 0; transition: transform var(--transition); color: var(--text-muted); }
        .tree-dir .chevron.collapsed { transform: rotate(-90deg); }
        .tree-file {
            display: flex; align-items: center; gap: 6px;
            padding: 4px 16px 4px 38px; cursor: pointer;
            font-size: 13.5px; color: var(--text-secondary);
            transition: all var(--transition); position: relative;
        }
        .tree-file:hover { background: var(--sidebar-hover); color: var(--text); }
        .tree-file.active { background: var(--sidebar-active); color: var(--accent); font-weight: 500; }
        .tree-file.active::before {
            content: ''; position: absolute; left: 0; top: 2px; bottom: 2px;
            width: 3px; background: var(--accent); border-radius: 0 2px 2px 0;
        }
        .tree-children { padding-left: 12px; }
        .tree-children.collapsed { display: none; }

        /* ==================== Content Area ==================== */
        .content-area { flex: 1; overflow-y: auto; padding: 0; position: relative; }
        .content-inner { max-width: 820px; margin: 0 auto; padding: 40px 48px 80px; }

        /* ==================== Note Content ==================== */
        .note-content { animation: fadeIn 0.2s ease; }
        @keyframes fadeIn { from { opacity: 0; transform: translateY(4px); } to { opacity: 1; transform: translateY(0); } }

        .note-content h1 { font-size: 2rem; font-weight: 700; line-height: 1.3; margin: 0 0 24px; letter-spacing: -0.5px; }
        .note-content h2 { font-size: 1.45rem; font-weight: 600; margin: 32px 0 16px; padding-bottom: 8px; border-bottom: 1px solid var(--border-light); letter-spacing: -0.3px; }
        .note-content h3 { font-size: 1.15rem; font-weight: 600; margin: 24px 0 12px; letter-spacing: -0.2px; }
        .note-content h4 { font-size: 1rem; font-weight: 600; margin: 20px 0 8px; }
        .note-content p { margin: 12px 0; line-height: 1.75; }
        .note-content ul, .note-content ol { margin: 12px 0; padding-left: 24px; }
        .note-content li { margin: 6px 0; line-height: 1.7; }
        .note-content li::marker { color: var(--text-muted); }
        .note-content code { font-family: var(--font-mono); font-size: 0.88em; background: var(--code-bg); padding: 2px 7px; border-radius: 4px; }
        .note-content pre { background: var(--code-bg); border-radius: var(--radius-md); overflow: hidden; margin: 16px 0; border: 1px solid var(--border-light); }
        .note-content pre code { display: block; padding: 16px 20px; background: none; overflow-x: auto; font-size: 13px; line-height: 1.6; }
        .note-content blockquote { border-left: 3px solid var(--accent); background: var(--accent-soft); padding: 12px 20px; margin: 16px 0; border-radius: 0 var(--radius-sm) var(--radius-sm) 0; color: var(--text-secondary); }
        .note-content blockquote p { margin: 4px 0; }
        .note-content table { border-collapse: collapse; width: 100%; margin: 16px 0; border-radius: var(--radius-md); overflow: hidden; border: 1px solid var(--border); }
        .note-content th { background: var(--bg-secondary); font-weight: 600; text-align: left; padding: 10px 16px; font-size: 13px; border-bottom: 2px solid var(--border); }
        .note-content td { padding: 10px 16px; border-bottom: 1px solid var(--border-light); font-size: 14px; }
        .note-content tr:last-child td { border-bottom: none; }
        .note-content tr:hover td { background: var(--bg-secondary); }
        .note-content img { max-width: 100%; height: auto; border-radius: var(--radius-md); margin: 8px 0; }
        .note-content a, .wikilink { color: var(--link); text-decoration: none; border-bottom: 1px solid transparent; transition: all var(--transition); }
        .note-content a:hover, .wikilink:hover { border-bottom-color: var(--link); }
        .note-content input[type="checkbox"] { margin-right: 8px; accent-color: var(--accent); }
        .note-content hr { border: none; height: 1px; background: var(--border); margin: 32px 0; }
        .broken-link { color: var(--broken-link); text-decoration: underline dotted; cursor: help; }
        .ambiguous-link { color: #e8a317; text-decoration: underline dotted; }
        .embed-image { margin: 16px 0; text-align: center; }
        .embed-image img { max-width: 100%; border-radius: var(--radius-md); box-shadow: var(--shadow-md); }
        .embed-pdf { margin: 16px 0; }
        .embed-pdf iframe { border: 1px solid var(--border); border-radius: var(--radius-md); }
        .embed-note { margin: 16px 0; padding: 14px 18px; background: var(--accent-soft); border-left: 3px solid var(--accent); border-radius: 0 var(--radius-md) var(--radius-md) 0; }
        .embed-broken { color: var(--broken-link); font-style: italic; padding: 8px 0; }

        /* ==================== Callouts ==================== */
        .callout {
            --callout-color: #448aff;
            margin: 16px 0;
            padding: 12px 16px;
            border-left: 4px solid var(--callout-color);
            background: color-mix(in srgb, var(--callout-color) 8%, var(--bg));
            border-radius: 0 var(--radius-md) var(--radius-md) 0;
            font-size: 14px;
        }
        .callout-title {
            display: flex;
            align-items: center;
            gap: 8px;
            font-weight: 600;
            font-size: 14px;
            color: var(--callout-color);
            margin-bottom: 6px;
            cursor: default;
        }
        .callout-icon { font-size: 16px; flex-shrink: 0; }
        .callout-title-text { flex: 1; }
        .callout-fold-icon { cursor: pointer; flex-shrink: 0; transition: transform var(--transition); opacity: 0.6; }
        .callout-foldable .callout-title { cursor: pointer; }
        .callout-foldable .callout-title:hover { opacity: 0.85; }
        .callout-foldable:not(.callout-collapsed) .callout-fold-icon svg { transform: rotate(180deg); }
        .callout-content { line-height: 1.65; }
        .callout-content p { margin: 4px 0; }
        .callout-content code { font-family: var(--font-mono); font-size: 0.88em; background: var(--code-bg); padding: 1px 5px; border-radius: 3px; }
        .callout-content pre { background: var(--code-bg); padding: 10px 14px; border-radius: var(--radius-sm); margin: 8px 0; overflow-x: auto; }
        .callout-content ul, .callout-content ol { margin: 4px 0; padding-left: 20px; }
        .callout-content li { margin: 2px 0; }
        .callout-collapsed .callout-content { display: none; }
        [data-theme="dark"] .callout { background: color-mix(in srgb, var(--callout-color) 12%, var(--bg)); }
        [id^="block-"] { scroll-margin-top: 20px; border-radius: var(--radius-sm); padding: 2px 4px; margin: -2px -4px; }

        /* ==================== Right Sidebar ==================== */
        .right-sidebar {
            width: var(--right-width); border-left: 1px solid var(--border);
            overflow-y: auto; padding: 20px 16px; background: var(--sidebar-bg); flex-shrink: 0;
            transition: background 0.25s, border-color 0.25s;
        }
        .right-sidebar h3 { font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.8px; color: var(--text-muted); margin: 20px 0 10px; padding-bottom: 8px; border-bottom: 1px solid var(--border-light); }
        .right-sidebar h3:first-child { margin-top: 0; }
        .toc-item { font-size: 13px; padding: 3px 0 3px 4px; border-left: 2px solid transparent; transition: all var(--transition); }
        .toc-item:hover { border-left-color: var(--accent); }
        .toc-item a { color: var(--text-secondary); text-decoration: none; display: block; padding: 2px 8px; border-radius: 4px; transition: all var(--transition); cursor: pointer; }
        .toc-item a:hover { color: var(--text); background: var(--sidebar-hover); }
        .tag-item { display: inline-flex; align-items: center; background: var(--accent-soft); color: var(--accent); padding: 3px 10px; border-radius: 20px; margin: 3px 2px; font-size: 12px; font-weight: 500; transition: all var(--transition); cursor: default; }
        .tag-item:hover { background: var(--accent); color: white; }
        .backlink-item { font-size: 13px; padding: 6px 8px; border-radius: var(--radius-sm); transition: background var(--transition); }
        .backlink-item:hover { background: var(--sidebar-hover); }
        .backlink-item a { color: var(--link); text-decoration: none; font-weight: 500; }
        .backlink-item a:hover { text-decoration: underline; }
        .backlink-item .bl-raw { color: var(--text-muted); font-size: 12px; margin-top: 2px; }
        .no-items { font-size: 12px; color: var(--text-muted); padding: 4px 0; font-style: italic; }

        /* Properties Panel */
        .prop-row { display: flex; align-items: baseline; padding: 5px 0; font-size: 13px; border-bottom: 1px solid var(--border-light); }
        .prop-row:last-child { border-bottom: none; }
        .prop-key { color: var(--text-muted); font-size: 12px; min-width: 70px; flex-shrink: 0; font-weight: 500; }
        .prop-value { color: var(--text); word-break: break-all; }
        .prop-value a { color: var(--link); text-decoration: none; }
        .prop-value a:hover { text-decoration: underline; }
        .prop-tag { display: inline-block; background: var(--accent-soft); color: var(--accent); padding: 1px 8px; border-radius: 12px; font-size: 11px; margin: 1px 2px; font-weight: 500; }
        .prop-alias { display: inline-block; background: var(--bg-secondary); border: 1px solid var(--border); color: var(--text-secondary); padding: 1px 8px; border-radius: 12px; font-size: 11px; margin: 1px 2px; }
        .prop-section { display: none; }
        .prop-section.has-props { display: block; }

        /* ==================== Search ==================== */
        .search-results {
            position: absolute; top: calc(var(--header-height) - 4px); left: 50%; transform: translateX(-50%);
            width: min(560px, 90vw); background: var(--bg); border: 1px solid var(--border);
            border-radius: var(--radius-lg); box-shadow: var(--shadow-lg); z-index: 100;
            max-height: 480px; overflow-y: auto;
        }
        .search-result-item { padding: 12px 18px; cursor: pointer; transition: background var(--transition); border-bottom: 1px solid var(--border-light); }
        .search-result-item:last-child { border-bottom: none; }
        .search-result-item:hover { background: var(--sidebar-hover); }
        .search-result-item:first-of-type { border-radius: var(--radius-lg) var(--radius-lg) 0 0; }
        .search-result-item:last-of-type { border-radius: 0 0 var(--radius-lg) var(--radius-lg); }
        .search-result-item:only-of-type { border-radius: var(--radius-lg); }
        .search-result-item .sr-title { font-weight: 600; font-size: 14px; }
        .search-result-item .sr-path { font-size: 12px; color: var(--text-muted); margin-top: 2px; }
        .search-result-item .sr-snippet { font-size: 13px; color: var(--text-secondary); margin-top: 4px; line-height: 1.5; }
        .search-result-item mark { background: var(--mark-bg); padding: 1px 3px; border-radius: 2px; }

        /* ==================== Welcome ==================== */
        .welcome { display: flex; flex-direction: column; align-items: center; justify-content: center; padding-top: 120px; text-align: center; }
        .welcome-icon { width: 72px; height: 72px; background: var(--accent-soft); border-radius: 18px; display: flex; align-items: center; justify-content: center; margin-bottom: 24px; }
        .welcome-icon svg { width: 36px; height: 36px; color: var(--accent); }
        .welcome h2 { font-size: 1.5rem; font-weight: 600; margin-bottom: 8px; }
        .welcome p { font-size: 15px; color: var(--text-muted); line-height: 1.6; }
        .error-page { text-align: center; padding-top: 120px; }
        .error-page h2 { font-size: 1.3rem; color: var(--broken-link); margin-bottom: 8px; }
        .error-page p { color: var(--text-muted); }
        .breadcrumb { font-size: 12px; color: var(--text-muted); padding: 0 0 20px; display: flex; align-items: center; gap: 4px; flex-wrap: wrap; }
        .breadcrumb span { cursor: pointer; transition: color var(--transition); }
        .breadcrumb span:hover { color: var(--link); }
        .breadcrumb .sep { cursor: default; }
        .breadcrumb .sep:hover { color: var(--text-muted); }

        /* ==================== Responsive ==================== */
        @media (max-width: 1100px) { .right-sidebar { display: none; } }
        @media (max-width: 768px) {
            :root { --sidebar-width: 220px; }
            .content-inner { padding: 24px 20px 60px; }
            .note-content h1 { font-size: 1.6rem; }
            .note-content h2 { font-size: 1.25rem; }
        }
        @media print { .header, .sidebar, .right-sidebar { display: none; } .content-area { overflow: visible; } .content-inner { max-width: 100%; padding: 0; } }
    </style>
</head>
<body>
    <div class="header">
        <div class="logo">
            <div class="logo-icon"><svg viewBox="0 0 24 24"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6zm-1 2l5 5h-5V4zM6 20V4h5v7h7v9H6z"/></svg></div>
            <span class="logo-text">Vault Reader</span>
        </div>
        <div class="search-container">
            <div class="search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
            <input type="text" class="search-box" placeholder="Search notes..." id="searchInput" autocomplete="off">
            <span class="shortcut-hint">/</span>
        </div>
        <button class="theme-toggle" id="themeToggle" title="Toggle theme">
            <svg class="icon-sun" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
            <svg class="icon-moon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
        </button>
        <div id="searchResults" class="search-results" style="display:none"></div>
    </div>
    <div class="main">
        <div class="sidebar" id="sidebar">
            <div class="sidebar-header">Explorer</div>
            <div class="sidebar-resizer" id="sidebarResizer"></div>
        </div>
        <div class="content-area" id="content">
            <div class="content-inner">
                <div class="welcome">
                    <div class="welcome-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg></div>
                    <h2>Vault Reader</h2>
                    <p>Select a note from the sidebar to begin reading</p>
                </div>
            </div>
        </div>
        <div class="right-sidebar" id="rightSidebar">
            <h3>Properties</h3>
            <div id="properties"></div>
            <h3>Outline</h3>
            <div id="toc"></div>
            <h3>Tags</h3>
            <div id="tags"></div>
            <h3>Backlinks</h3>
            <div id="backlinks"></div>
        </div>
    </div>
    <script>
    const $ = id => document.getElementById(id);
    let currentPath = null;

    // ==================== Theme Toggle ====================
    (function() {
        const saved = localStorage.getItem('vault-reader-theme');
        if (saved) {
            document.documentElement.setAttribute('data-theme', saved);
        } else {
            // Default: follow system
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            document.documentElement.setAttribute('data-theme', prefersDark ? 'dark' : 'light');
        }
        $('themeToggle').onclick = () => {
            const current = document.documentElement.getAttribute('data-theme');
            const next = current === 'dark' ? 'light' : 'dark';
            document.documentElement.setAttribute('data-theme', next);
            localStorage.setItem('vault-reader-theme', next);
        };
    })();

    // ==================== Icons ====================
    const svgChevron = '<svg class="chevron" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M4 6l4 4 4-4"/></svg>';

    // File icon: filled doc shape with soft accent color
    function svgFile(isActive) {
        const fill = isActive ? 'var(--accent)' : 'var(--text-muted)';
        const bg = isActive ? 'var(--accent-soft)' : 'var(--bg-secondary)';
        return '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">' +
            '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="' + fill + '" fill="' + bg + '"/>' +
            '<polyline points="14 2 14 8 20 8" stroke="' + fill + '"/>' +
            '<line x1="8" y1="13" x2="16" y2="13" stroke="' + fill + '" stroke-width="1"/>' +
            '<line x1="8" y1="17" x2="13" y2="17" stroke="' + fill + '" stroke-width="1"/>' +
            '</svg>';
    }

    function svgFolder(color) {
        return '<svg class="folder-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">' +
            '<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" stroke="' + color + '" fill="' + color + '18"/></svg>';
    }

    // Folder color palette
    const folderColors = [
        '#5b8def', '#50b87a', '#e8a838', '#e07850', '#8b6fc0',
        '#4db8c7', '#d66b8e', '#6d8e3f', '#c47a3f', '#7c8594',
    ];
    const folderColorMap = {};
    let colorIndex = 0;
    function getFolderColor(name) {
        if (!folderColorMap[name]) { folderColorMap[name] = folderColors[colorIndex % folderColors.length]; colorIndex++; }
        return folderColorMap[name];
    }

    // Strip .md / .markdown extension for display
    function stripMdExt(name) {
        return name.replace(/\.(md|markdown)$/i, '');
    }

    // ==================== Tree ====================
    async function loadTree() {
        const resp = await fetch('/api/tree');
        const tree = await resp.json();
        const sidebar = $('sidebar');
        sidebar.innerHTML = '<div class="sidebar-header">Explorer</div>';
        renderTree(tree.children || [], sidebar);
        // Re-add resizer
        const r = document.createElement('div');
        r.className = 'sidebar-resizer'; r.id = 'sidebarResizer';
        sidebar.appendChild(r);
    }

    function renderTree(items, container) {
        items.forEach(item => {
            const div = document.createElement('div');
            div.className = 'tree-item';
            if (item.type === 'dir') {
                const dirEl = document.createElement('div');
                dirEl.className = 'tree-dir';
                const color = getFolderColor(item.name);
                dirEl.innerHTML = svgChevron + svgFolder(color) + '<span style="color:' + color + '">' + escHtml(item.name) + '</span>';
                div.appendChild(dirEl);
                const children = document.createElement('div');
                children.className = 'tree-children';
                renderTree(item.children || [], children);
                div.appendChild(children);
                const chevron = dirEl.querySelector('.chevron');
                dirEl.onclick = () => { children.classList.toggle('collapsed'); chevron.classList.toggle('collapsed'); };
            } else {
                const isActive = item.path === currentPath;
                const fileEl = document.createElement('div');
                fileEl.className = 'tree-file' + (isActive ? ' active' : '');
                fileEl.innerHTML = svgFile(isActive) + '<span>' + escHtml(stripMdExt(item.name)) + '</span>';
                fileEl.onclick = () => loadNote(item.path);
                div.appendChild(fileEl);
            }
            container.appendChild(div);
        });
    }

    // ==================== Load Note ====================
    async function loadNote(path) {
        currentPath = path;
        // Extract hash for block/heading navigation
        let hashTarget = '';
        const hashIdx = path.indexOf('#');
        if (hashIdx !== -1) {
            hashTarget = path.substring(hashIdx + 1);
            path = path.substring(0, hashIdx);
        }
        try {
            const resp = await fetch('/api/note?path=' + encodeURIComponent(path));
            if (!resp.ok) throw new Error(resp.statusText);
            const note = await resp.json();

            const parts = path.split('/');
            const bcParts = parts.map(p => ' <span class="sep">/</span> <span>' + escHtml(stripMdExt(p)) + '</span>').join('');
            $('content').innerHTML = '<div class="content-inner">' +
                '<div class="breadcrumb"><span>Vault</span>' + bcParts + '</div>' +
                '<div class="note-content">' + (note.html || '') + '</div></div>';

            renderTOC(note.headings || []);
            renderTags(note.tags || []);
            renderBacklinks(note.backlinks || []);
            loadProperties(path);
            document.title = note.title + ' - Vault Reader';
            $('content').scrollTop = 0;
            loadTree();

            // Scroll to block or heading if hash target exists
            if (hashTarget) {
                scrollToElement(hashTarget);
            }
        } catch(e) {
            $('content').innerHTML = '<div class="content-inner"><div class="error-page"><h2>Error</h2><p>' + escHtml(e.message) + '</p></div></div>';
        }
    }

    // Scroll to a heading or block element and highlight it
    function scrollToElement(id) {
        const contentArea = $('content');
        let target = contentArea.querySelector('#' + CSS.escape(id));
        if (target) {
            target.scrollIntoView({ behavior: 'smooth', block: 'start' });
            target.style.transition = 'background 0.3s';
            target.style.background = 'var(--accent-soft)';
            target.style.borderRadius = 'var(--radius-sm)';
            setTimeout(() => { target.style.background = ''; target.style.borderRadius = ''; }, 2000);
        }
    }

    // ==================== Right Sidebar ====================
    function renderTOC(headings) {
        if (headings.length === 0) { $('toc').innerHTML = '<div class="no-items">No headings</div>'; return; }
        $('toc').innerHTML = headings.map(h =>
            '<div class="toc-item" style="padding-left:' + ((h.level - 1) * 12 + 4) + 'px">' +
            '<a data-slug="' + escAttr(h.slug) + '">' + escHtml(h.text) + '</a></div>'
        ).join('');
    }

    function renderTags(tags) {
        if (!tags || tags.length === 0) { $('tags').innerHTML = '<div class="no-items">No tags</div>'; return; }
        $('tags').innerHTML = tags.map(t => '<span class="tag-item">' + escHtml(t) + '</span>').join('');
    }

    function renderBacklinks(links) {
        if (!links || !Array.isArray(links) || links.length === 0) { $('backlinks').innerHTML = '<div class="no-items">No backlinks</div>'; return; }
        $('backlinks').innerHTML = links.map(l =>
            '<div class="backlink-item"><a href="#" onclick="loadNote(\'' + escAttr(l.fromPath) + '\');return false">' +
            escHtml(stripMdExt(l.title || l.fromPath.split('/').pop())) + '</a>' +
            '<div class="bl-raw">' + escHtml(l.raw) + '</div></div>'
        ).join('');
    }

    // ==================== Properties Panel ====================
    async function loadProperties(path) {
        try {
            const resp = await fetch('/api/properties?path=' + encodeURIComponent(path));
            if (!resp.ok) { $('properties').innerHTML = '<div class="no-items">No properties</div>'; return; }
            const data = await resp.json();
            renderProperties(data.items || []);
        } catch(e) {
            $('properties').innerHTML = '<div class="no-items">No properties</div>';
        }
    }

    function renderProperties(props) {
        if (!props || props.length === 0) {
            $('properties').innerHTML = '<div class="no-items">No properties</div>';
            return;
        }

        // Group array values by key
        const grouped = {};
        props.forEach(p => {
            if (!grouped[p.key]) grouped[p.key] = { values: [], type: p.valueType };
            grouped[p.key].values.push(p.value);
        });

        let html = '';
        for (const [key, data] of Object.entries(grouped)) {
            const label = escHtml(key);
            let valueHtml = '';

            if (key === 'aliases') {
                valueHtml = data.values.map(v => '<span class="prop-alias">' + escHtml(v) + '</span>').join('');
            } else if (key === 'tags') {
                // Tags already shown in Tags section, show as subtle tags
                valueHtml = data.values.map(v => '<span class="prop-tag">' + escHtml(v) + '</span>').join('');
            } else if (data.values.length === 1) {
                const v = data.values[0];
                if (v.startsWith('http://') || v.startsWith('https://')) {
                    valueHtml = '<a href="' + escHtml(v) + '" target="_blank" rel="noopener">' + escHtml(v) + '</a>';
                } else {
                    valueHtml = escHtml(v);
                }
            } else {
                valueHtml = data.values.map(v => escHtml(v)).join(', ');
            }

            html += '<div class="prop-row"><span class="prop-key">' + label + '</span><span class="prop-value">' + valueHtml + '</span></div>';
        }

        $('properties').innerHTML = html;
    }

    // ==================== TOC Click: scroll content to heading ====================
    document.addEventListener('click', e => {
        const tocLink = e.target.closest('.toc-item a');
        if (!tocLink) return;
        e.preventDefault();

        const slug = tocLink.getAttribute('data-slug');
        if (!slug) return;

        const contentArea = $('content');
        // Try by ID first
        let target = contentArea.querySelector('#' + CSS.escape(slug));

        // Fallback: iterate headings, match by comparing slug of text
        if (!target) {
            const headings = contentArea.querySelectorAll('.note-content h1, .note-content h2, .note-content h3, .note-content h4, .note-content h5, .note-content h6');
            for (const h of headings) {
                if (slugifyText(h.textContent) === slug) { target = h; break; }
            }
        }

        // Last fallback: match display text
        if (!target) {
            const headings = contentArea.querySelectorAll('.note-content h1, .note-content h2, .note-content h3, .note-content h4, .note-content h5, .note-content h6');
            const wanted = tocLink.textContent.trim();
            for (const h of headings) {
                if (h.textContent.trim() === wanted) { target = h; break; }
            }
        }

        if (target) {
            target.scrollIntoView({ behavior: 'smooth', block: 'start' });
            // Brief highlight
            target.style.transition = 'background 0.3s';
            target.style.background = 'var(--accent-soft)';
            target.style.borderRadius = 'var(--radius-sm)';
            setTimeout(() => { target.style.background = ''; target.style.borderRadius = ''; }, 1200);
        }
    });

    function slugifyText(text) {
        return text.trim().toLowerCase().replace(/\s+/g, '-').replace(/[^\w\u4e00-\u9fff-]/g, '');
    }

    // ==================== Search ====================
    let searchTimeout;
    $('searchInput').addEventListener('input', e => {
        clearTimeout(searchTimeout);
        const q = e.target.value.trim();
        if (!q) { $('searchResults').style.display = 'none'; return; }
        searchTimeout = setTimeout(async () => {
            try {
                const resp = await fetch('/api/search?q=' + encodeURIComponent(q));
                const data = await resp.json();
                const results = data.items || [];
                if (results.length === 0) {
                    $('searchResults').innerHTML = '<div class="search-result-item"><div class="sr-snippet" style="text-align:center;padding:8px">No results found</div></div>';
                } else {
                    $('searchResults').innerHTML = results.map(r =>
                        '<div class="search-result-item" onclick="loadNote(\'' + escAttr(r.path) + '\');$(\'searchResults\').style.display=\'none\'">' +
                        '<div class="sr-title">' + escHtml(stripMdExt(r.title)) + '</div>' +
                        '<div class="sr-path">' + escHtml(r.path) + '</div>' +
                        '<div class="sr-snippet">' + (r.snippet || '') + '</div></div>'
                    ).join('');
                }
                $('searchResults').style.display = 'block';
            } catch(e) {}
        }, 250);
    });

    // ==================== Keyboard Shortcuts ====================
    document.addEventListener('keydown', e => {
        if (e.key === '/' && document.activeElement.tagName !== 'INPUT') { e.preventDefault(); $('searchInput').focus(); }
        if (e.key === 'Escape') { $('searchInput').blur(); $('searchResults').style.display = 'none'; }
    });

    document.addEventListener('click', e => {
        if (!e.target.closest('.search-container') && !e.target.closest('.search-results'))
            $('searchResults').style.display = 'none';
    });

    // ==================== Wikilink Clicks ====================
    document.addEventListener('click', e => {
        const link = e.target.closest('a.wikilink');
        if (link) {
            e.preventDefault();
            const path = link.getAttribute('data-path');
            if (path) {
                // Check if link has a hash (block ref or heading)
                const href = link.getAttribute('href') || '';
                const hashIdx = href.indexOf('#');
                if (hashIdx !== -1) {
                    loadNote(path + '#' + href.substring(hashIdx + 1));
                } else {
                    loadNote(path);
                }
            }
        }
    });

    // ==================== Callout Fold Toggle ====================
    document.addEventListener('click', e => {
        const title = e.target.closest('.callout-foldable .callout-title');
        if (!title) return;
        const callout = title.closest('.callout');
        callout.classList.toggle('callout-collapsed');
        const isCollapsed = callout.classList.contains('callout-collapsed');
        callout.setAttribute('data-collapsed', isCollapsed);
    });

    // ==================== Sidebar Resize ====================
    (function() {
        let isResizing = false;
        document.addEventListener('mousedown', e => {
            const resizer = e.target.closest('.sidebar-resizer');
            if (!resizer) return;
            isResizing = true;
            resizer.classList.add('active');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
            e.preventDefault();
        });
        document.addEventListener('mousemove', e => {
            if (!isResizing) return;
            const sidebar = $('sidebar');
            const maxW = window.innerWidth * 0.3;
            sidebar.style.width = Math.min(Math.max(e.clientX, 160), maxW) + 'px';
        });
        document.addEventListener('mouseup', () => {
            if (!isResizing) return;
            isResizing = false;
            document.querySelectorAll('.sidebar-resizer').forEach(r => r.classList.remove('active'));
            document.body.style.cursor = '';
            document.body.style.userSelect = '';
        });
    })();

    // ==================== Helpers ====================
    function escHtml(s) { if (!s) return ''; return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }
    function escAttr(s) { return s.replace(/\\/g,'\\\\').replace(/'/g,"\\'"); }

    // ==================== Init ====================
    loadTree();
    </script>
</body>
</html>`
