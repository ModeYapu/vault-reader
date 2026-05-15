package server

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN" id="htmlRoot">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Vault Reader</title>
    <style>
        @font-face { font-family: 'Inter'; font-style: normal; font-weight: 300; font-display: swap; src: url(__P__/vendor/inter-300.ttf) format('truetype'); }
        @font-face { font-family: 'Inter'; font-style: normal; font-weight: 400; font-display: swap; src: url(__P__/vendor/inter-400.ttf) format('truetype'); }
        @font-face { font-family: 'Inter'; font-style: normal; font-weight: 500; font-display: swap; src: url(__P__/vendor/inter-500.ttf) format('truetype'); }
        @font-face { font-family: 'Inter'; font-style: normal; font-weight: 600; font-display: swap; src: url(__P__/vendor/inter-600.ttf) format('truetype'); }
        @font-face { font-family: 'Inter'; font-style: normal; font-weight: 700; font-display: swap; src: url(__P__/vendor/inter-700.ttf) format('truetype'); }
        @font-face { font-family: 'JetBrains Mono'; font-style: normal; font-weight: 400; font-display: swap; src: url(__P__/vendor/jetbrains-400.ttf) format('truetype'); }
        @font-face { font-family: 'JetBrains Mono'; font-style: normal; font-weight: 500; font-display: swap; src: url(__P__/vendor/jetbrains-500.ttf) format('truetype'); }
    </style>
    <script src="__P__/vendor/mermaid.min.js"></script>
    <script>const BASE='__P__';</script>
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
            --code-header-bg: #e8eaed;
            --code-header-text: #5f6368;
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
            --code-header-bg: #222222;
            --code-header-text: #9aa0a6;
        }

        /* ==================== Chroma Syntax Highlighting (Light) ==================== */
        :root, [data-theme="light"] {
            .chroma { color: #1f2328; background-color: #f7f7f7; }
            .chroma .err { color: #f6f8fa; background-color: #82071e }
            .chroma .lnlinks { outline: none; text-decoration: none; color: inherit }
            .chroma .lntd { vertical-align: top; padding: 0; margin: 0; border: 0; }
            .chroma .lntable { border-spacing: 0; padding: 0; margin: 0; border: 0; }
            .chroma .hl { background-color: #dedede }
            .chroma .lnt { white-space: pre; -webkit-user-select: none; user-select: none; margin-right: 0.4em; padding: 0 0.4em; color: #7f7f7f }
            .chroma .ln { white-space: pre; -webkit-user-select: none; user-select: none; margin-right: 0.4em; padding: 0 0.4em; color: #7f7f7f }
            .chroma .line { display: flex; }
            .chroma .k { color: #cf222e } .chroma .kc { color: #cf222e } .chroma .kd { color: #cf222e }
            .chroma .kn { color: #cf222e } .chroma .kp { color: #cf222e } .chroma .kr { color: #cf222e } .chroma .kt { color: #cf222e }
            .chroma .na { color: #1f2328 } .chroma .nc { color: #1f2328 } .chroma .no { color: #0550ae }
            .chroma .nd { color: #0550ae } .chroma .ni { color: #6639ba } .chroma .nl { color: #990000; font-weight: bold }
            .chroma .nn { color: #24292e } .chroma .nx { color: #1f2328 } .chroma .nt { color: #0550ae }
            .chroma .nb { color: #6639ba } .chroma .bp { color: #6a737d }
            .chroma .nv { color: #953800 } .chroma .vc { color: #953800 } .chroma .vg { color: #953800 }
            .chroma .vi { color: #953800 } .chroma .vm { color: #953800 }
            .chroma .nf { color: #6639ba } .chroma .fm { color: #6639ba }
            .chroma .s { color: #0a3069 } .chroma .sa { color: #0a3069 } .chroma .sb { color: #0a3069 }
            .chroma .sc { color: #0a3069 } .chroma .dl { color: #0a3069 } .chroma .sd { color: #0a3069 }
            .chroma .s2 { color: #0a3069 } .chroma .se { color: #0a3069 } .chroma .sh { color: #0a3069 }
            .chroma .si { color: #0a3069 } .chroma .sx { color: #0a3069 } .chroma .sr { color: #0a3069 }
            .chroma .s1 { color: #0a3069 } .chroma .ss { color: #032f62 }
            .chroma .m { color: #0550ae } .chroma .mb { color: #0550ae } .chroma .mf { color: #0550ae }
            .chroma .mh { color: #0550ae } .chroma .mi { color: #0550ae } .chroma .il { color: #0550ae }
            .chroma .mo { color: #0550ae }
            .chroma .o { color: #0550ae } .chroma .ow { color: #0550ae } .chroma .or { color: #0550ae }
            .chroma .p { color: #1f2328 }
            .chroma .c { color: #57606a } .chroma .ch { color: #57606a } .chroma .cm { color: #57606a }
            .chroma .c1 { color: #57606a } .chroma .cs { color: #57606a } .chroma .cp { color: #57606a } .chroma .cpf { color: #57606a }
            .chroma .gd { color: #82071e; background-color: #ffebe9 }
            .chroma .ge { color: #1f2328 } .chroma .gi { color: #116329; background-color: #dafbe1 }
            .chroma .go { color: #1f2328 } .chroma .gl { text-decoration: underline }
            .chroma .w { color: #ffffff }
        }

        /* ==================== Chroma Syntax Highlighting (Dark) ==================== */
        [data-theme="dark"] {
            .chroma { color: #f8f8f2; background-color: #272822; }
            .chroma .err { color: #960050; background-color: #1e0010 }
            .chroma .lnlinks { outline: none; text-decoration: none; color: inherit }
            .chroma .lntd { vertical-align: top; padding: 0; margin: 0; border: 0; }
            .chroma .lntable { border-spacing: 0; padding: 0; margin: 0; border: 0; }
            .chroma .hl { background-color: #3c3d38 }
            .chroma .lnt { white-space: pre; -webkit-user-select: none; user-select: none; margin-right: 0.4em; padding: 0 0.4em; color: #7f7f7f }
            .chroma .ln { white-space: pre; -webkit-user-select: none; user-select: none; margin-right: 0.4em; padding: 0 0.4em; color: #7f7f7f }
            .chroma .line { display: flex; }
            .chroma .k { color: #66d9ef } .chroma .kc { color: #66d9ef } .chroma .kd { color: #66d9ef }
            .chroma .kn { color: #f92672 } .chroma .kp { color: #66d9ef } .chroma .kr { color: #66d9ef } .chroma .kt { color: #66d9ef }
            .chroma .na { color: #a6e22e } .chroma .nc { color: #a6e22e } .chroma .no { color: #66d9ef }
            .chroma .nd { color: #a6e22e } .chroma .ne { color: #a6e22e } .chroma .nx { color: #a6e22e }
            .chroma .nt { color: #f92672 }
            .chroma .nf { color: #a6e22e } .chroma .fm { color: #a6e22e }
            .chroma .l { color: #ae81ff } .chroma .ld { color: #e6db74 }
            .chroma .s { color: #e6db74 } .chroma .sa { color: #e6db74 } .chroma .sb { color: #e6db74 }
            .chroma .sc { color: #e6db74 } .chroma .dl { color: #e6db74 } .chroma .sd { color: #e6db74 }
            .chroma .s2 { color: #e6db74 } .chroma .se { color: #ae81ff } .chroma .sh { color: #e6db74 }
            .chroma .si { color: #e6db74 } .chroma .sx { color: #e6db74 } .chroma .sr { color: #e6db74 }
            .chroma .s1 { color: #e6db74 } .chroma .ss { color: #e6db74 }
            .chroma .m { color: #ae81ff } .chroma .mb { color: #ae81ff } .chroma .mf { color: #ae81ff }
            .chroma .mh { color: #ae81ff } .chroma .mi { color: #ae81ff } .chroma .il { color: #ae81ff }
            .chroma .mo { color: #ae81ff }
            .chroma .o { color: #f92672 } .chroma .ow { color: #f92672 } .chroma .or { color: #f92672 }
            .chroma .c { color: #75715e } .chroma .ch { color: #75715e } .chroma .cm { color: #75715e }
            .chroma .c1 { color: #75715e } .chroma .cs { color: #75715e } .chroma .cp { color: #75715e } .chroma .cpf { color: #75715e }
            .chroma .gd { color: #f92672 } .chroma .ge { font-style: italic }
            .chroma .gi { color: #a6e22e } .chroma .gs { font-weight: bold } .chroma .gu { color: #75715e }
        }

        /* Common */
        --radius-sm: 6px;
        --radius-md: 8px;
        --radius-lg: 12px;
        --font-sans: 'Inter', 'Microsoft YaHei', 'PingFang SC', 'Noto Sans SC', -apple-system, BlinkMacSystemFont, sans-serif;
        --font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
        --header-height: 52px;
        --sidebar-width: 280px;
        --right-width: 260px;
        --transition: 150ms cubic-bezier(0.4, 0, 0.2, 1);

        /* ==================== Reset & Base ==================== */
        *, *::before, *::after { margin: 0; padding: 0; box-sizing: border-box; }
        html { font-size: 15px; -webkit-font-smoothing: antialiased; -moz-osx-font-smoothing: grayscale; }
        body {
            margin: 0;
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
            position: relative;
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
        .note-content pre { position: relative; background: var(--code-bg); border-radius: var(--radius-md); overflow: hidden; margin: 16px 0; border: 1px solid var(--border-light); }
        .note-content pre code { display: block; padding: 16px 20px; background: none; overflow-x: auto; font-size: 13px; line-height: 1.6; }
        /* Chroma syntax highlighting — code blocks with language */
        .note-content pre.chroma { padding-top: 0; }
        .note-content pre.chroma code { padding-top: 40px; }
        .note-content pre .code-lang { position: absolute; top: 0; left: 0; right: 0; padding: 6px 44px 6px 14px; font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; color: var(--code-header-text); background: var(--code-header-bg); border-bottom: 1px solid var(--border-light); font-family: var(--font-sans); }
        .note-content pre .code-copy { position: absolute; top: 4px; right: 6px; background: none; border: 1px solid var(--border); border-radius: 4px; padding: 3px 8px; font-size: 11px; color: var(--text-muted); cursor: pointer; font-family: var(--font-sans); transition: all var(--transition); z-index: 2; }
        .note-content pre .code-copy:hover { color: var(--text); background: var(--bg-secondary); border-color: var(--text-muted); }
        .note-content pre .code-copy.copied { color: #1a9933; border-color: #1a9933; }
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
        .mermaid-diagram { margin: 16px 0; text-align: center; overflow-x: auto; overflow-y: visible; position: relative; padding: 8px 8px 8px 8px; border: 1px solid var(--border-light); border-radius: var(--radius-md); background: var(--bg-secondary); }
        .mermaid-diagram svg { max-width: 100%; height: auto; pointer-events: none; }
        .mermaid-diagram .diagram-enlarge { position: absolute; top: 12px; right: 12px; width: 32px; height: 32px; border-radius: 6px; border: 1px solid var(--border); background: var(--bg); color: var(--text-muted); cursor: pointer; display: flex; align-items: center; justify-content: center; font-size: 16px; box-shadow: var(--shadow-md); transition: all var(--transition); z-index: 10; pointer-events: auto; }
        .mermaid-diagram .diagram-enlarge:hover { color: var(--accent); border-color: var(--accent); background: var(--accent-soft); }
        .mermaid-error { border: 1px dashed var(--error, #e74c3c); padding: 8px 12px; border-radius: var(--radius-sm); background: color-mix(in srgb, var(--error, #e74c3c) 8%, var(--bg)); }
        .mermaid-error-msg { color: var(--error, #e74c3c); font-size: 0.85em; margin-top: 4px; }
        /* Mermaid fullscreen viewer */
        .mermaid-overlay { position: fixed; inset: 0; z-index: 1000; background: rgba(0,0,0,0.7); display: flex; align-items: center; justify-content: center; opacity: 0; pointer-events: none; transition: opacity 0.2s; }
        .mermaid-overlay.active { opacity: 1; pointer-events: auto; }
        .mermaid-overlay .mo-close { position: absolute; top: 16px; right: 20px; background: rgba(255,255,255,0.15); border: none; color: #fff; font-size: 28px; width: 40px; height: 40px; border-radius: 50%; cursor: pointer; display: flex; align-items: center; justify-content: center; z-index: 2; }
        .mermaid-overlay .mo-close:hover { background: rgba(255,255,255,0.3); }
        .mermaid-overlay .mo-toolbar { position: absolute; top: 16px; left: 50%; transform: translateX(-50%); display: flex; gap: 8px; z-index: 2; }
        .mermaid-overlay .mo-toolbar button { background: rgba(255,255,255,0.15); border: none; color: #fff; font-size: 13px; padding: 6px 14px; border-radius: 6px; cursor: pointer; }
        .mermaid-overlay .mo-toolbar button:hover { background: rgba(255,255,255,0.3); }
        .mermaid-overlay .mo-viewport { width: 90vw; height: 85vh; overflow: hidden; position: relative; background: var(--bg); border-radius: var(--radius-lg); box-shadow: var(--shadow-lg); }
        .mermaid-overlay .mo-inner { transform-origin: 0 0; transition: transform 0.15s ease; }
        .mermaid-overlay svg { max-width: none; height: auto; }

        /* ==================== Canvas Viewer ==================== */
        .canvas-container { position: relative; width: 100%; height: calc(100vh - var(--header-height) - 80px); overflow: hidden; background: var(--bg-secondary); border-radius: var(--radius-md); margin: 16px 0; cursor: grab; }
        .canvas-container:active { cursor: grabbing; }
        .canvas-viewport { position: absolute; top: 0; left: 0; transform-origin: 0 0; }
        .canvas-node { position: absolute; border: 2px solid var(--border); border-radius: var(--radius-md); background: var(--bg); overflow: hidden; font-size: 13px; box-shadow: var(--shadow-sm); transition: box-shadow var(--transition); }
        .canvas-node:hover { box-shadow: var(--shadow-md); }
        .canvas-node-header { padding: 8px 12px; font-weight: 600; font-size: 12px; border-bottom: 1px solid var(--border-light); display: flex; align-items: center; gap: 6px; }
        .canvas-node-body { padding: 10px 12px; max-height: 200px; overflow-y: auto; line-height: 1.5; }
        .canvas-node-body p { margin: 4px 0; }
        .canvas-node.file-node { cursor: pointer; }
        .canvas-node.file-node:hover { border-color: var(--accent); }
        .canvas-node.link-node { cursor: pointer; }
        .canvas-node.link-node:hover { border-color: var(--accent); }
        .canvas-node.group-node { background: transparent; border: 2px dashed var(--border); }
        .canvas-node.group-node .canvas-node-body { display: none; }
        .canvas-edges { position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; overflow: visible; }
        .canvas-toolbar { position: absolute; bottom: 12px; right: 12px; display: flex; gap: 6px; z-index: 10; }
        .canvas-toolbar button { width: 32px; height: 32px; border: 1px solid var(--border); background: var(--bg); border-radius: var(--radius-sm); cursor: pointer; font-size: 16px; color: var(--text-secondary); display: flex; align-items: center; justify-content: center; }
        .canvas-toolbar button:hover { background: var(--sidebar-hover); color: var(--text); }

        /* ==================== Graph View ==================== */
        .graph-container { position: relative; width: 100%; height: calc(100vh - var(--header-height) - 80px); overflow: hidden; background: var(--bg-secondary); border-radius: var(--radius-md); margin: 16px 0; }
        .graph-container svg { width: 100%; height: 100%; }
        .graph-node { cursor: pointer; }
        .graph-node:hover circle { r: 8; }
        .graph-node text { font-family: var(--font-sans); font-size: 11px; fill: var(--text-secondary); pointer-events: none; }
        .graph-edge { stroke: var(--border); stroke-width: 1.5; stroke-opacity: 0.4; }
        .graph-controls { position: absolute; top: 12px; right: 12px; display: flex; gap: 6px; z-index: 10; }
        .graph-controls button { padding: 6px 12px; border: 1px solid var(--border); background: var(--bg); border-radius: var(--radius-sm); cursor: pointer; font-size: 12px; color: var(--text-secondary); }
        .graph-controls button:hover { background: var(--sidebar-hover); color: var(--text); }
        .graph-controls button.active { background: var(--accent-soft); color: var(--accent); border-color: var(--accent); }
        .graph-info { position: absolute; bottom: 12px; left: 12px; font-size: 11px; color: var(--text-muted); }

        /* ==================== Dashboard ==================== */
        .dashboard { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 20px; padding: 20px 0; }
        .dash-card { background: var(--bg-secondary); border: 1px solid var(--border-light); border-radius: var(--radius-lg); padding: 20px; }
        .dash-card h3 { font-size: 13px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.6px; color: var(--text-muted); margin: 0 0 14px; }
        .dash-item { font-size: 14px; padding: 6px 0; display: flex; align-items: center; gap: 8px; }
        .dash-item a { color: var(--link); text-decoration: none; font-weight: 500; }
        .dash-item a:hover { text-decoration: underline; }
        .dash-item .dash-path { font-size: 12px; color: var(--text-muted); }
        .dash-tag { display: inline-block; background: var(--accent-soft); color: var(--accent); padding: 2px 8px; border-radius: 12px; font-size: 12px; margin: 2px; cursor: pointer; }
        .dash-tag:hover { background: var(--accent); color: white; }
        .dash-canvas-item { font-size: 14px; padding: 6px 0; }
        .dash-canvas-item a { color: #e8a838; text-decoration: none; font-weight: 500; }

        /* ==================== Vault Query ==================== */
        .vq-table { width: 100%; border-collapse: collapse; margin: 16px 0; font-size: 14px; border: 1px solid var(--border); border-radius: var(--radius-md); overflow: hidden; }
        .vq-table th { background: var(--bg-secondary); font-weight: 600; text-align: left; padding: 8px 12px; font-size: 12px; text-transform: uppercase; color: var(--text-muted); border-bottom: 1px solid var(--border); }
        .vq-table td { padding: 8px 12px; border-bottom: 1px solid var(--border-light); }
        .vq-table tr:last-child td { border-bottom: none; }
        .vq-table tr:hover td { background: var(--bg-secondary); }
        .vq-table a { color: var(--link); text-decoration: none; font-weight: 500; }
        .vq-list { list-style: none; padding: 0; margin: 16px 0; }
        .vq-list li { padding: 6px 12px; border-bottom: 1px solid var(--border-light); }
        .vq-list a { color: var(--link); text-decoration: none; }
        .vq-cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 12px; margin: 16px 0; }
        .vq-card { background: var(--bg-secondary); border: 1px solid var(--border-light); border-radius: var(--radius-md); padding: 14px; }
        .vq-card-title { font-weight: 600; font-size: 14px; margin-bottom: 6px; }
        .vq-card-title a { color: var(--link); text-decoration: none; }
        .vq-card-field { font-size: 12px; color: var(--text-secondary); margin: 2px 0; }
        .vq-error { border: 1px dashed var(--broken-link); padding: 8px 12px; border-radius: var(--radius-sm); color: var(--broken-link); font-size: 13px; margin: 8px 0; }

        /* ==================== Right Sidebar ==================== */
        .right-sidebar {
            width: var(--right-width); max-width: 24vw; border-left: 1px solid var(--border);
            overflow-y: auto; overflow-x: hidden; padding: 20px 16px; background: var(--sidebar-bg); flex-shrink: 0;
            transition: background 0.25s, border-color 0.25s;
        }
        .right-sidebar h3 { font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.8px; color: var(--text-muted); margin: 20px 0 10px; padding-bottom: 8px; border-bottom: 1px solid var(--border-light); }
        .right-sidebar h3:first-child { margin-top: 0; }
        .toc-item { font-size: 13px; padding: 3px 0 3px 4px; border-left: 2px solid transparent; transition: all var(--transition); }
        .toc-item:hover { border-left-color: var(--accent); }
        .toc-item a { color: var(--text-secondary); text-decoration: none; display: block; padding: 2px 8px; border-radius: 4px; transition: all var(--transition); cursor: pointer; }
        .toc-item a:hover { color: var(--text); background: var(--sidebar-hover); }
        .tag-item { display: inline-flex; align-items: center; background: var(--accent-soft); color: var(--accent); padding: 3px 10px; border-radius: 20px; margin: 3px 2px; font-size: 12px; font-weight: 500; transition: all var(--transition); cursor: pointer; text-decoration: none; max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .tag-item:hover { background: var(--accent); color: white; }
        .tag-tree-node { font-size: 13px; padding: 3px 0; }
        .tag-tree-label { display: flex; align-items: center; gap: 4px; padding: 2px 6px; border-radius: 4px; cursor: pointer; transition: background var(--transition); }
        .tag-tree-label:hover { background: var(--sidebar-hover); }
        .tag-tree-label .tag-tree-count { color: var(--text-muted); font-size: 11px; margin-left: auto; }
        .tag-tree-children { padding-left: 16px; }
        .tag-tree-toggle { width: 14px; height: 14px; flex-shrink: 0; color: var(--text-muted); transition: transform var(--transition); }
        .tag-tree-toggle.collapsed { transform: rotate(-90deg); }
        .tag-files { padding: 8px 0; }
        .tag-file-item { font-size: 13px; padding: 4px 8px; border-radius: var(--radius-sm); transition: background var(--transition); }
        .tag-file-item:hover { background: var(--sidebar-hover); }
        .tag-file-item a { color: var(--link); text-decoration: none; font-weight: 500; }
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
            position: absolute; top: 50%; left: 50%; transform: translateX(-50%);
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

        /* ==================== Mobile Hamburger ==================== */
        .hamburger {
            display: none; width: 36px; height: 36px; border: none; background: transparent;
            cursor: pointer; color: var(--text-muted); align-items: center; justify-content: center;
            flex-shrink: 0; border-radius: var(--radius-sm); transition: all var(--transition);
        }
        .hamburger:hover { color: var(--text); background: var(--bg-secondary); }
        .hamburger svg { width: 20px; height: 20px; }
        .sidebar-backdrop {
            display: none; position: fixed; inset: 0; background: rgba(0,0,0,0.4);
            z-index: 90; opacity: 0; transition: opacity 0.2s;
        }
        .sidebar-backdrop.active { opacity: 1; }

        /* ==================== Responsive ==================== */
        @media (max-width: 1100px) { .right-sidebar { display: none; } }
        @media (max-width: 768px) {
            :root { --sidebar-width: 280px; --header-height: 48px; }
            .hamburger { display: flex; }
            .logo-text { display: none; }
            .shortcut-hint { display: none; }
            .sidebar {
                position: fixed; top: 0; left: 0; bottom: 0;
                z-index: 100; transform: translateX(-100%);
                transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
                box-shadow: none;
            }
            .sidebar.open { transform: translateX(0); box-shadow: var(--shadow-lg); }
            .sidebar-backdrop { display: block; }
            .sidebar-resizer { display: none; }
            .content-inner { padding: 20px 16px 80px; }
            .note-content h1 { font-size: 1.5rem; }
            .note-content h2 { font-size: 1.2rem; }
            .note-content h3 { font-size: 1.05rem; }
            .note-content table { font-size: 13px; display: block; overflow-x: auto; }
            .note-content th, .note-content td { padding: 8px 10px; }
            .note-content pre code { font-size: 12px; padding: 12px 14px; }
            .note-content blockquote { padding: 10px 14px; }
            .mermaid-diagram { margin: 12px -8px; border-radius: 0; }
            .mermaid-overlay .mo-viewport { width: 98vw; height: 90vh; border-radius: 0; }
            .search-container { max-width: none; }
            .header { padding: 0 12px; gap: 10px; }
            .breadcrumb { display: none; }
            .tree-dir, .tree-file { padding: 8px 16px; font-size: 14px; }
        }
        @media print { .header, .sidebar, .right-sidebar { display: none; } .content-area { overflow: visible; } .content-inner { max-width: 100%; padding: 0; } }
    </style>
</head>
<body>
    <div class="header">
        <button class="hamburger" id="hamburgerBtn" onclick="toggleSidebar()">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
        </button>
        <div class="logo">
            <div class="logo-icon"><svg viewBox="0 0 24 24"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6zm-1 2l5 5h-5V4zM6 20V4h5v7h7v9H6z"/></svg></div>
            <span class="logo-text">Vault Reader</span>
        </div>
        <div class="search-container">
            <div class="search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
            <input type="text" class="search-box" placeholder="Search notes..." id="searchInput" autocomplete="off">
            <span class="shortcut-hint">/</span>
        </div>
        <button class="theme-toggle" id="graphBtn" title="Graph View" onclick="toggleGraph()">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="6" cy="6" r="3"/><circle cx="18" cy="6" r="3"/><circle cx="12" cy="18" r="3"/><line x1="8" y1="7" x2="11" y2="16"/><line x1="16" y1="7" x2="13" y2="16"/><line x1="8.5" y1="5.5" x2="15.5" y2="5.5"/></svg>
        </button>
        <button class="theme-toggle" id="themeToggle" title="Toggle theme">
            <svg class="icon-sun" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
            <svg class="icon-moon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
        </button>
        <div id="searchResults" class="search-results" style="display:none"></div>
    </div>
    <div class="sidebar-backdrop" id="sidebarBackdrop" onclick="toggleSidebar()"></div>
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
            <h3>Tag Tree</h3>
            <div id="tagTree"></div>
            <h3>Backlinks</h3>
            <div id="backlinks"></div>
        </div>
    </div>
    <script>
    const $ = id => document.getElementById(id);
    let currentPath = null;

    // ==================== Mobile Sidebar Toggle ====================
    function toggleSidebar() {
        const sidebar = $('sidebar');
        const backdrop = $('sidebarBackdrop');
        const isOpen = sidebar.classList.toggle('open');
        backdrop.classList.toggle('active', isOpen);
        backdrop.style.pointerEvents = isOpen ? 'auto' : 'none';
    }
    // Close sidebar when a note is selected on mobile
    function closeSidebarOnMobile() {
        if (window.innerWidth <= 768) {
            const sidebar = $('sidebar');
            const backdrop = $('sidebarBackdrop');
            sidebar.classList.remove('open');
            backdrop.classList.remove('active');
            backdrop.style.pointerEvents = 'none';
        }
    }

    // ==================== Mermaid Init ====================
    (function() {
        if (typeof mermaid !== 'undefined') {
            const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
            mermaid.initialize({
                startOnLoad: false,
                theme: isDark ? 'dark' : 'default',
                securityLevel: 'strict',
                fontFamily: 'var(--font-mono)'
            });
        }
    })();

    async function renderVaultQueries() {
        const content = document.querySelector('.note-content');
        if (!content) return;
        const blocks = content.querySelectorAll('code.language-vault-query');
        for (const block of blocks) {
            const pre = block.parentElement;
            if (!pre || pre.tagName !== 'PRE') continue;
            if (pre.classList.contains('vq-rendered')) continue;
            pre.classList.add('vq-rendered');
            const yaml = block.textContent;
            try {
                const resp = await fetch(BASE+'/api/vault-query', {
                    method: 'POST',
                    headers: {'Content-Type': 'text/plain'},
                    body: yaml
                });
                if (!resp.ok) throw new Error(resp.statusText);
                const data = await resp.json();
                pre.replaceWith(renderQueryResult(data));
            } catch(err) {
                const errDiv = document.createElement('div');
                errDiv.className = 'vq-error';
                errDiv.textContent = 'Vault Query error: ' + (err.message || err);
                pre.replaceWith(errDiv);
            }
        }
    }

    function renderQueryResult(data) {
        const results = data.results || [];
        const fields = data.fields || [];
        const type = data.type || 'table';

        if (results.length === 0) {
            const div = document.createElement('div');
            div.className = 'vq-error';
            div.textContent = 'Vault Query: no results';
            return div;
        }

        const container = document.createElement('div');

        if (type === 'table') {
            let html = '<table class="vq-table"><thead><tr><th>Title</th>';
            fields.forEach(f => { html += '<th>' + escHtml(f) + '</th>'; });
            html += '</tr></thead><tbody>';
            results.forEach(r => {
                html += '<tr><td><a href="#" onclick="loadNote(\'' + escAttr(r.path) + '\');return false">' +
                    escHtml(stripMdExt(r.title || r.path.split('/').pop())) + '</a></td>';
                fields.forEach(f => {
                    const val = (r.fields && r.fields[f]) || '-';
                    html += '<td>' + escHtml(val) + '</td>';
                });
                html += '</tr>';
            });
            html += '</tbody></table>';
            container.innerHTML = html;
        } else if (type === 'list') {
            let html = '<ul class="vq-list">';
            results.forEach(r => {
                html += '<li><a href="#" onclick="loadNote(\'' + escAttr(r.path) + '\');return false">' +
                    escHtml(stripMdExt(r.title || r.path.split('/').pop())) + '</a></li>';
            });
            html += '</ul>';
            container.innerHTML = html;
        } else if (type === 'cards') {
            let html = '<div class="vq-cards">';
            results.forEach(r => {
                html += '<div class="vq-card"><div class="vq-card-title"><a href="#" onclick="loadNote(\'' +
                    escAttr(r.path) + '\');return false">' + escHtml(stripMdExt(r.title || r.path.split('/').pop())) + '</a></div>';
                if (r.fields) {
                    for (const [k, v] of Object.entries(r.fields)) {
                        if (fields.length === 0 || fields.includes(k)) {
                            html += '<div class="vq-card-field">' + escHtml(k) + ': ' + escHtml(v) + '</div>';
                        }
                    }
                }
                html += '</div>';
            });
            html += '</div>';
            container.innerHTML = html;
        }

        return container;
    }

    const ENLARGE_BTN = '<button class="diagram-enlarge" title="放大查看"><svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="11" y1="8" x2="11" y2="14"/><line x1="8" y1="11" x2="14" y2="11"/></svg></button>';

    async function renderMermaidDiagrams() {
        if (typeof mermaid === 'undefined') return;
        const content = document.querySelector('.note-content');
        if (!content) return;
        const blocks = content.querySelectorAll('pre.language-mermaid, code.language-mermaid');
        console.log('[mermaid] found', blocks.length, 'mermaid blocks');
        for (const block of blocks) {
            const pre = block.tagName === 'PRE' ? block : block.parentElement;
            if (!pre || pre.tagName !== 'PRE') continue;
            if (pre.classList.contains('mermaid-rendered')) continue;
            pre.classList.add('mermaid-rendered');
            const id = 'mermaid-' + Math.random().toString(36).substr(2, 9);
            const codeEl = pre.querySelector('code');
            let src = codeEl ? codeEl.textContent : pre.textContent;
            // Normalize XHTML self-closing tags for Mermaid compatibility
            src = src.replace(/<br\s*\/>/gi, '<br>');
            try {
                const { svg } = await mermaid.render(id, src);
                const wrapper = document.createElement('div');
                wrapper.className = 'mermaid-diagram';
                wrapper.innerHTML = svg + ENLARGE_BTN;
                console.log('[mermaid] rendered', id, 'children:', wrapper.children.length);
                pre.replaceWith(wrapper);
            } catch (err) {
                console.error('[mermaid] render error:', err);
                pre.classList.add('mermaid-error');
                const errorDiv = document.createElement('div');
                errorDiv.className = 'mermaid-error-msg';
                errorDiv.textContent = 'Mermaid render error: ' + (err.message || err);
                pre.appendChild(errorDiv);
            }
        }
    }

    // Safety net: add enlarge buttons to any .mermaid-diagram that doesn't have one
    function addEnlargeButtons() {
        document.querySelectorAll('.mermaid-diagram').forEach(d => {
            if (!d.querySelector('.diagram-enlarge')) {
                d.insertAdjacentHTML('beforeend', ENLARGE_BTN);
                console.log('[mermaid] added missing enlarge button');
            }
        });
    }

    function openMermaidViewer(svgHtml) {
        const overlay = document.createElement('div');
        overlay.className = 'mermaid-overlay';
        overlay.innerHTML =
            '<button class="mo-close">&times;</button>' +
            '<div class="mo-toolbar">' +
            '<button class="mo-zout">-</button>' +
            '<button class="mo-reset">Reset</button>' +
            '<button class="mo-zin">+</button>' +
            '</div>' +
            '<div class="mo-viewport"><div class="mo-inner">' + svgHtml + '</div></div>';
        document.body.appendChild(overlay);
        requestAnimationFrame(() => overlay.classList.add('active'));

        const viewport = overlay.querySelector('.mo-viewport');
        const inner = overlay.querySelector('.mo-inner');
        let scale = 1, panX = 0, panY = 0, dragging = false, startX, startY;

        function apply() { inner.style.transform = 'translate(' + panX + 'px,' + panY + 'px) scale(' + scale + ')'; }

        // Fit to viewport
        const svg = inner.querySelector('svg');
        if (svg) {
            const vb = svg.getAttribute('viewBox');
            if (vb) {
                const parts = vb.split(/[\s,]+/).map(Number);
                const svgW = parts[2] || 800, svgH = parts[3] || 600;
                scale = Math.min(viewport.clientWidth / svgW, viewport.clientHeight / svgH, 1);
                panX = (viewport.clientWidth - svgW * scale) / 2;
                panY = (viewport.clientHeight - svgH * scale) / 2;
            }
        }
        apply();

        function doZoom(delta) {
            const old = scale;
            scale = Math.max(0.05, Math.min(scale + delta, 8));
            const cx = viewport.clientWidth / 2, cy = viewport.clientHeight / 2;
            panX = cx - (cx - panX) * (scale / old);
            panY = cy - (cy - panY) * (scale / old);
            apply();
        }
        function doReset() {
            if (svg) {
                const vb = svg.getAttribute('viewBox');
                if (vb) {
                    const parts = vb.split(/[\s,]+/).map(Number);
                    const svgW = parts[2] || 800, svgH = parts[3] || 600;
                    scale = Math.min(viewport.clientWidth / svgW, viewport.clientHeight / svgH, 1);
                    panX = (viewport.clientWidth - svgW * scale) / 2;
                    panY = (viewport.clientHeight - svgH * scale) / 2;
                }
            }
            apply();
        }

        overlay.querySelector('.mo-close').onclick = () => { overlay.remove(); };
        overlay.querySelector('.mo-zin').onclick = (e) => { e.stopPropagation(); doZoom(0.2); };
        overlay.querySelector('.mo-zout').onclick = (e) => { e.stopPropagation(); doZoom(-0.2); };
        overlay.querySelector('.mo-reset').onclick = (e) => { e.stopPropagation(); doReset(); };
        overlay.onclick = (e) => { if (e.target === overlay) overlay.remove(); };

        viewport.onmousedown = (e) => { if (e.target.closest('.mo-toolbar')) return; dragging = true; startX = e.clientX - panX; startY = e.clientY - panY; e.preventDefault(); };
        const onMove = (e) => { if (!dragging) return; panX = e.clientX - startX; panY = e.clientY - startY; apply(); };
        const onUp = () => { dragging = false; };
        window.addEventListener('mousemove', onMove);
        window.addEventListener('mouseup', onUp);

        viewport.onwheel = (e) => {
            e.preventDefault();
            const rect = viewport.getBoundingClientRect();
            const mx = e.clientX - rect.left, my = e.clientY - rect.top;
            const old = scale;
            scale += e.deltaY < 0 ? 0.15 : -0.15;
            scale = Math.max(0.05, Math.min(scale, 8));
            panX = mx - (mx - panX) * (scale / old);
            panY = my - (my - panY) * (scale / old);
            apply();
        };

        // Cleanup on close
        const origRemove = overlay.remove.bind(overlay);
        overlay.remove = () => { window.removeEventListener('mousemove', onMove); window.removeEventListener('mouseup', onUp); origRemove(); };
    }

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

    function svgCanvasIcon(isActive) {
        const fill = isActive ? 'var(--accent)' : '#e8a838';
        return '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="' + fill + '" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">' +
            '<rect x="3" y="3" width="18" height="18" rx="3"/>' +
            '<circle cx="8" cy="8" r="2" fill="' + fill + '"/>' +
            '<circle cx="16" cy="16" r="2" fill="' + fill + '"/>' +
            '<line x1="10" y1="8" x2="14" y2="16"/></svg>';
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
        const resp = await fetch(BASE+'/api/tree');
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
                const icon = item.isCanvas ? svgCanvasIcon(isActive) : svgFile(isActive);
                fileEl.innerHTML = icon + '<span>' + escHtml(item.isCanvas ? item.name : stripMdExt(item.name)) + '</span>';
                fileEl.onclick = () => loadNote(item.path);
                div.appendChild(fileEl);
            }
            container.appendChild(div);
        });
    }

    // ==================== Load Note ====================
    async function loadNote(path) {
        closeSidebarOnMobile();
        currentPath = path;
        // Extract hash for block/heading navigation
        let hashTarget = '';
        const hashIdx = path.indexOf('#');
        if (hashIdx !== -1) {
            hashTarget = path.substring(hashIdx + 1);
            path = path.substring(0, hashIdx);
        }

        // Canvas file: show canvas viewer
        if (path.endsWith('.canvas')) {
            loadCanvas(path);
            return;
        }

        try {
            const resp = await fetch(BASE+'/api/note?path=' + encodeURIComponent(path));
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
            loadTagTree();
            document.title = note.title + ' - Vault Reader';
            $('content').scrollTop = 0;
            loadTree();

            // Render mermaid diagrams
            await renderMermaidDiagrams();
            // Safety net: ensure enlarge buttons exist after async rendering
            addEnlargeButtons();

            // Enhance code blocks: add language label + copy button
            enhanceCodeBlocks();

            // Render vault-query blocks
            renderVaultQueries();

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

    // ==================== Mermaid Click (delegated) ====================
    document.addEventListener('click', e => {
        const btn = e.target.closest('.diagram-enlarge');
        if (btn) {
            e.preventDefault();
            e.stopPropagation();
            const diagram = btn.closest('.mermaid-diagram');
            if (diagram) {
                const svg = diagram.querySelector('svg');
                if (svg) openMermaidViewer(svg.outerHTML);
            }
            return;
        }
        const diagram = e.target.closest('.mermaid-diagram');
        if (diagram) {
            e.preventDefault();
            e.stopPropagation();
            const svg = diagram.querySelector('svg');
            if (svg) openMermaidViewer(svg.outerHTML);
        }
    });

    // MutationObserver: auto-add enlarge buttons to any new mermaid diagrams
    new MutationObserver(mutations => {
        for (const m of mutations) {
            for (const node of m.addedNodes) {
                if (node.nodeType === 1) {
                    if (node.classList && node.classList.contains('mermaid-diagram')) {
                        if (!node.querySelector('.diagram-enlarge')) {
                            node.insertAdjacentHTML('beforeend', ENLARGE_BTN);
                        }
                    } else if (node.querySelector) {
                        node.querySelectorAll('.mermaid-diagram').forEach(d => {
                            if (!d.querySelector('.diagram-enlarge')) {
                                d.insertAdjacentHTML('beforeend', ENLARGE_BTN);
                            }
                        });
                    }
                }
            }
        }
    }).observe(document.body, { childList: true, subtree: true });

    // ==================== Code Block Enhancement ====================
    function enhanceCodeBlocks() {
        const blocks = document.querySelectorAll('.note-content pre');
        blocks.forEach(pre => {
            const code = pre.querySelector('code');
            if (!code) return;

            // Detect language from pre class (chroma wrapper adds language-xxx on <pre>)
            let lang = '';
            const preClass = pre.className || '';
            const m = preClass.match(/language-(\S+)/);
            if (m) lang = m[1];

            // Also check code class for fallback
            if (!lang) {
                const codeClass = code.className || '';
                const cm = codeClass.match(/language-(\S+)/);
                if (cm) lang = cm[1];
            }

            if (!lang) return;

            // Skip mermaid blocks — they get rendered as diagrams
            if (lang === 'mermaid') return;

            // Add language label
            const label = document.createElement('span');
            label.className = 'code-lang';
            label.textContent = lang;
            pre.insertBefore(label, pre.firstChild);

            // Add copy button
            const btn = document.createElement('button');
            btn.className = 'code-copy';
            btn.textContent = 'Copy';
            btn.onclick = function() {
                const text = code.textContent || code.innerText;
                navigator.clipboard.writeText(text).then(() => {
                    btn.textContent = 'Copied!';
                    btn.classList.add('copied');
                    setTimeout(() => { btn.textContent = 'Copy'; btn.classList.remove('copied'); }, 1500);
                });
            };
            pre.insertBefore(btn, label.nextSibling);
        });
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
        $('tags').innerHTML = tags.map(t => '<span class="tag-item" onclick="showTagFiles(\'' + escAttr(t) + '\')">' + escHtml(t) + '</span>').join('');
    }

    // ==================== Tag Tree ====================
    let tagTreeData = null;
    async function loadTagTree() {
        try {
            const resp = await fetch(BASE+'/api/tag-tree');
            if (!resp.ok) { $('tagTree').innerHTML = '<div class="no-items">Tag tree unavailable</div>'; return; }
            const data = await resp.json();
            tagTreeData = data.items || [];
            renderTagTree(tagTreeData);
        } catch(e) {
            $('tagTree').innerHTML = '<div class="no-items">Tag tree unavailable</div>';
        }
    }

    function renderTagTree(nodes) {
        if (!nodes || nodes.length === 0) { $('tagTree').innerHTML = '<div class="no-items">No tags</div>'; return; }
        $('tagTree').innerHTML = nodes.map(n => renderTagTreeNode(n)).join('');
    }

    function renderTagTreeNode(node) {
        const hasChildren = node.children && node.children.length > 0;
        const toggleSvg = hasChildren
            ? '<svg class="tag-tree-toggle" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M4 6l4 4 4-4"/></svg>'
            : '<span style="width:14px;display:inline-block"></span>';
        let html = '<div class="tag-tree-node">' +
            '<div class="tag-tree-label" onclick="handleTagTreeClick(event, this, \'' + escAttr(node.fullName) + '\')">' +
            toggleSvg +
            '<span>' + escHtml(node.name) + '</span>' +
            (node.count > 0 ? '<span class="tag-tree-count">' + node.count + '</span>' : '') +
            '</div>';
        if (hasChildren) {
            html += '<div class="tag-tree-children">' + node.children.map(c => renderTagTreeNode(c)).join('') + '</div>';
        }
        html += '</div>';
        return html;
    }

    function handleTagTreeClick(event, labelEl, fullName) {
        // Toggle children visibility
        const node = labelEl.parentElement;
        const children = node.querySelector(':scope > .tag-tree-children');
        const toggle = labelEl.querySelector('.tag-tree-toggle');
        if (children) {
            children.style.display = children.style.display === 'none' ? '' : 'none';
            if (toggle) toggle.classList.toggle('collapsed', children.style.display === 'none');
        }
        // If this tag has a count, show files
        showTagFiles(fullName);
    }

    async function showTagFiles(tag) {
        try {
            const resp = await fetch(BASE+'/api/tag?name=' + encodeURIComponent(tag));
            if (!resp.ok) return;
            const data = await resp.json();
            const files = data.items || [];
            if (files.length === 0) return;
            // Show files in a popup-like overlay below the tag tree
            let html = '<div class="tag-files"><div style="font-size:11px;color:var(--text-muted);margin-bottom:4px;">Files with #' + escHtml(tag) + '</div>';
            html += files.map(f =>
                '<div class="tag-file-item"><a href="#" onclick="loadNote(\'' + escAttr(f.path) + '\');return false">' +
                escHtml(stripMdExt(f.title || f.path.split('/').pop())) + '</a></div>'
            ).join('');
            html += '</div>';
            $('tagTree').innerHTML = html + '<div style="margin-top:8px"><a href="#" onclick="loadTagTree();return false" style="font-size:12px;color:var(--text-muted)">&larr; Back to tag tree</a></div>';
        } catch(e) {}
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
            const resp = await fetch(BASE+'/api/properties?path=' + encodeURIComponent(path));
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
                const resp = await fetch(BASE+'/api/search?q=' + encodeURIComponent(q));
                const data = await resp.json();
                const results = data.items || [];
                if (results.length === 0) {
                    $('searchResults').innerHTML = '<div class="search-result-item"><div class="sr-snippet" style="text-align:center;padding:8px">No results found</div></div>';
                } else {
                    $('searchResults').innerHTML = results.map(r =>
                        '<div class="search-result-item" onclick="loadNote(\'' + escAttr(r.path) + '\');$(\'searchResults\').style.display=\'none\'">' +
                        '<div class="sr-title">' + escHtml(stripMdExt(r.title)) + '</div>' +
                        '<div class="sr-path">' + escHtml(r.path) + '</div>' +
                        '<div class="sr-snippet">' + escHtml(r.snippet || '') + '</div></div>'
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

    // ==================== Graph View ====================
    let graphVisible = false;

    async function toggleGraph() {
        graphVisible = !graphVisible;
        $('graphBtn').classList.toggle('active', graphVisible);
        if (graphVisible) {
            await loadGraphView();
        } else {
            // Restore previous note or welcome
            if (currentPath && !currentPath.endsWith('.canvas')) {
                loadNote(currentPath);
            } else {
                $('content').innerHTML = '<div class="content-inner"><div class="welcome"><div class="welcome-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg></div><h2>Vault Reader</h2><p>Select a note from the sidebar to begin reading</p></div></div>';
            }
        }
    }

    async function loadGraphView(folder, tag) {
        let url = BASE+'/api/graph?max=200';
        if (folder) url += '&folder=' + encodeURIComponent(folder);
        if (tag) url += '&tag=' + encodeURIComponent(tag);

        try {
            const resp = await fetch(url);
            if (!resp.ok) throw new Error(resp.statusText);
            const data = await resp.json();

            $('content').innerHTML = '<div class="content-inner">' +
                '<div style="display:flex;align-items:center;gap:12px;margin-bottom:16px">' +
                '<h1 style="margin:0;font-size:1.5rem">Graph View</h1>' +
                (folder ? '<span style="font-size:13px;color:var(--text-muted)">Folder: ' + escHtml(folder) + '</span>' : '') +
                (tag ? '<span style="font-size:13px;color:var(--text-muted)">Tag: #' + escHtml(tag) + '</span>' : '') +
                '</div>' +
                '<div class="graph-container" id="graphContainer"></div></div>';

            renderForceGraph(data.nodes || [], data.edges || []);
        } catch(e) {
            $('content').innerHTML = '<div class="content-inner"><div class="error-page"><h2>Error</h2><p>' + escHtml(e.message) + '</p></div></div>';
        }
    }

    function renderForceGraph(nodes, edges) {
        const container = $('graphContainer');
        if (!container) return;
        const w = container.clientWidth, h = container.clientHeight;

        // Simple force simulation
        const sim = { nodes: [], alpha: 1 };
        nodes.forEach(n => {
            sim.nodes.push({ ...n, x: w/2 + (Math.random()-0.5)*w*0.5, y: h/2 + (Math.random()-0.5)*h*0.5, vx: 0, vy: 0 });
        });

        const nodeMap = {};
        sim.nodes.forEach((n,i) => { nodeMap[n.id || n.path] = i; });

        const edgeIdxs = [];
        edges.forEach(e => {
            const si = nodeMap[e.source], ti = nodeMap[e.target];
            if (si !== undefined && ti !== undefined) edgeIdxs.push({source: si, target: ti});
        });

        // Color map by group
        const groups = {};
        let gc = 0;
        sim.nodes.forEach(n => {
            const g = n.group || 'default';
            if (!groups[g]) { groups[g] = folderColors[gc % folderColors.length]; gc++; }
            n.color = groups[g];
        });

        // Run simulation iterations
        for (let iter = 0; iter < 200; iter++) {
            sim.alpha *= 0.98;
            // Repulsion
            for (let i = 0; i < sim.nodes.length; i++) {
                for (let j = i+1; j < sim.nodes.length; j++) {
                    const dx = sim.nodes[j].x - sim.nodes[i].x;
                    const dy = sim.nodes[j].y - sim.nodes[i].y;
                    const dist = Math.sqrt(dx*dx + dy*dy) || 1;
                    const force = 3000 / (dist * dist);
                    const fx = dx/dist * force * sim.alpha;
                    const fy = dy/dist * force * sim.alpha;
                    sim.nodes[i].vx -= fx; sim.nodes[i].vy -= fy;
                    sim.nodes[j].vx += fx; sim.nodes[j].vy += fy;
                }
            }
            // Attraction along edges
            edgeIdxs.forEach(e => {
                const s = sim.nodes[e.source], t = sim.nodes[e.target];
                const dx = t.x - s.x, dy = t.y - s.y;
                const dist = Math.sqrt(dx*dx + dy*dy) || 1;
                const force = (dist - 120) * 0.02 * sim.alpha;
                const fx = dx/dist * force, fy = dy/dist * force;
                s.vx += fx; s.vy += fy;
                t.vx -= fx; t.vy -= fy;
            });
            // Center gravity
            sim.nodes.forEach(n => {
                n.vx += (w/2 - n.x) * 0.001 * sim.alpha;
                n.vy += (h/2 - n.y) * 0.001 * sim.alpha;
            });
            // Apply velocity with damping
            sim.nodes.forEach(n => {
                n.vx *= 0.6; n.vy *= 0.6;
                n.x += n.vx; n.y += n.vy;
                // Keep in bounds
                n.x = Math.max(40, Math.min(w-40, n.x));
                n.y = Math.max(20, Math.min(h-20, n.y));
            });
        }

        // Render SVG
        let svg = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 '+w+' '+h+'">';
        // Edges
        edgeIdxs.forEach(e => {
            const s = sim.nodes[e.source], t = sim.nodes[e.target];
            svg += '<line class="graph-edge" x1="'+s.x+'" y1="'+s.y+'" x2="'+t.x+'" y2="'+t.y+'"/>';
        });
        // Nodes
        sim.nodes.forEach(n => {
            svg += '<g class="graph-node" onclick="loadNote(\''+escAttr(n.path)+'\')">' +
                '<circle cx="'+n.x+'" cy="'+n.y+'" r="6" fill="'+n.color+'" stroke="var(--bg)" stroke-width="2"/>' +
                '<text x="'+n.x+'" y="'+(n.y-12)+'" text-anchor="middle">'+escHtml(stripMdExt(n.title||n.path.split('/').pop()))+'</text>' +
                '</g>';
        });
        svg += '</svg>';

        container.innerHTML = svg +
            '<div class="graph-info">' + nodes.length + ' nodes, ' + edges.length + ' links</div>' +
            '<div class="graph-controls">' +
            '<button onclick="loadGraphView()">All</button>' +
            '</div>';
    }

    // ==================== Canvas Viewer ====================
    async function loadCanvas(path) {
        try {
            const resp = await fetch(BASE+'/api/canvas?path=' + encodeURIComponent(path));
            if (!resp.ok) throw new Error(resp.statusText);
            const doc = await resp.json();

            const parts = path.split('/');
            const bcParts = parts.map(p => ' <span class="sep">/</span> <span>' + escHtml(stripMdExt(p)) + '</span>').join('');
            $('content').innerHTML = '<div class="content-inner">' +
                '<div class="breadcrumb"><span>Vault</span>' + bcParts + '</div>' +
                '<h1>' + escHtml(stripMdExt(path.split('/').pop())) + '</h1>' +
                '<div class="canvas-container" id="canvasContainer">' +
                '<div class="canvas-viewport" id="canvasViewport">' +
                '<svg class="canvas-edges" id="canvasEdges"></svg>' +
                '</div>' +
                '<div class="canvas-toolbar">' +
                '<button onclick="canvasZoom(1.2)" title="Zoom in">+</button>' +
                '<button onclick="canvasZoom(1/1.2)" title="Zoom out">&minus;</button>' +
                '<button onclick="canvasReset()" title="Reset view">&#x21bb;</button>' +
                '</div>' +
                '</div></div>';

            renderCanvas(doc);
            loadTree();
            document.title = stripMdExt(path.split('/').pop()) + ' - Vault Reader';

            // Clear right sidebar
            $('properties').innerHTML = '<div class="no-items">Canvas file</div>';
            $('toc').innerHTML = '<div class="no-items">No headings</div>';
            $('tags').innerHTML = '<div class="no-items">No tags</div>';
            $('backlinks').innerHTML = '<div class="no-items">No backlinks</div>';
        } catch(e) {
            $('content').innerHTML = '<div class="content-inner"><div class="error-page"><h2>Error</h2><p>' + escHtml(e.message) + '</p></div></div>';
        }
    }

    let canvasState = { panX: 0, panY: 0, scale: 1, dragging: false, startX: 0, startY: 0 };

    function renderCanvas(doc) {
        const viewport = $('canvasViewport');
        const svg = $('canvasEdges');
        if (!viewport || !svg) return;

        // Render nodes
        const nodes = doc.nodes || [];
        const nodeEls = {};

        nodes.forEach(node => {
            const el = document.createElement('div');
            el.className = 'canvas-node ' + node.type + '-node';
            el.style.left = node.x + 'px';
            el.style.top = node.y + 'px';
            el.style.width = node.width + 'px';
            el.style.height = node.height + 'px';
            if (node.color) { el.style.borderColor = node.color; }

            let header = '';
            let body = '';
            if (node.type === 'text') {
                header = '<div class="canvas-node-header">&#128196; Text</div>';
                body = '<div class="canvas-node-body">' + escHtml(node.text || '') + '</div>';
            } else if (node.type === 'file') {
                const fileName = stripMdExt(node.file ? node.file.split('/').pop() : '');
                header = '<div class="canvas-node-header">&#128209; ' + escHtml(fileName) + '</div>';
                body = '<div class="canvas-node-body"><em>' + escHtml(node.file || '') + '</em></div>';
                el.onclick = () => { if (node.file) loadNote(node.file); };
            } else if (node.type === 'link') {
                header = '<div class="canvas-node-header">&#128279; Link</div>';
                body = '<div class="canvas-node-body"><a href="' + escHtml(node.url || '') + '" target="_blank" rel="noopener">' + escHtml(node.url || '') + '</a></div>';
            } else if (node.type === 'group') {
                header = '<div class="canvas-node-header">' + escHtml(node.label || 'Group') + '</div>';
                body = '<div class="canvas-node-body"></div>';
            }

            el.innerHTML = header + body;
            viewport.appendChild(el);
            nodeEls[node.id] = el;
        });

        // Render edges
        const edges = doc.edges || [];
        let svgContent = '';
        edges.forEach(edge => {
            const from = nodes.find(n => n.id === edge.fromNode);
            const to = nodes.find(n => n.id === edge.toNode);
            if (!from || !to) return;

            const x1 = from.x + from.width / 2;
            const y1 = from.y + from.height / 2;
            const x2 = to.x + to.width / 2;
            const y2 = to.y + to.height / 2;

            const color = edge.color || 'var(--text-muted)';
            svgContent += '<line x1="' + x1 + '" y1="' + y1 + '" x2="' + x2 + '" y2="' + y2 + '" stroke="' + color + '" stroke-width="2" stroke-opacity="0.5"/>';
            if (edge.label) {
                const mx = (x1 + x2) / 2, my = (y1 + y2) / 2;
                svgContent += '<text x="' + mx + '" y="' + my + '" fill="var(--text-secondary)" font-size="11" text-anchor="middle">' + escHtml(edge.label) + '</text>';
            }
        });
        svg.innerHTML = svgContent;

        // Setup pan and zoom
        setupCanvasInteraction();
    }

    function setupCanvasInteraction() {
        const container = $('canvasContainer');
        const viewport = $('canvasViewport');
        if (!container || !viewport) return;

        canvasState = { panX: 0, panY: 0, scale: 1, dragging: false, startX: 0, startY: 0 };

        container.onmousedown = e => {
            if (e.target.closest('.canvas-node') && !e.target.closest('.group-node')) return;
            canvasState.dragging = true;
            canvasState.startX = e.clientX - canvasState.panX;
            canvasState.startY = e.clientY - canvasState.panY;
        };

        container.onmousemove = e => {
            if (!canvasState.dragging) return;
            canvasState.panX = e.clientX - canvasState.startX;
            canvasState.panY = e.clientY - canvasState.startY;
            updateCanvasTransform();
        };

        container.onmouseup = () => { canvasState.dragging = false; };
        container.onmouseleave = () => { canvasState.dragging = false; };

        container.onwheel = e => {
            e.preventDefault();
            const delta = e.deltaY > 0 ? 0.9 : 1.1;
            canvasState.scale = Math.max(0.1, Math.min(5, canvasState.scale * delta));
            updateCanvasTransform();
        };
    }

    function updateCanvasTransform() {
        const viewport = $('canvasViewport');
        if (viewport) {
            viewport.style.transform = 'translate(' + canvasState.panX + 'px,' + canvasState.panY + 'px) scale(' + canvasState.scale + ')';
        }
    }

    function canvasZoom(factor) {
        canvasState.scale = Math.max(0.1, Math.min(5, canvasState.scale * factor));
        updateCanvasTransform();
    }

    function canvasReset() {
        canvasState.panX = 0;
        canvasState.panY = 0;
        canvasState.scale = 1;
        updateCanvasTransform();
    }

    // ==================== Helpers ====================
    function escHtml(s) { if (!s) return ''; return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }
    function escAttr(s) { if (!s) return ''; return s.replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/'/g,'&#39;').replace(/</g,'&lt;').replace(/>/g,'&gt;'); }

    // ==================== Dashboard ====================
    async function loadDashboard() {
        try {
            const resp = await fetch(BASE+'/api/dashboard');
            if (!resp.ok) throw new Error(resp.statusText);
            const data = await resp.json();

            let html = '<div class="dashboard">';

            // Recent
            if (data.recent && data.recent.length > 0) {
                html += '<div class="dash-card"><h3>Recent</h3>';
                data.recent.forEach(f => {
                    html += '<div class="dash-item"><a href="#" onclick="loadNote(\'' + escAttr(f.path) + '\');return false">' +
                        escHtml(stripMdExt(f.title || f.path.split('/').pop())) + '</a>' +
                        '<span class="dash-path">' + escHtml(f.path) + '</span></div>';
                });
                html += '</div>';
            }

            // Inbox
            if (data.inbox && data.inbox.length > 0) {
                html += '<div class="dash-card"><h3>Inbox</h3>';
                data.inbox.forEach(f => {
                    html += '<div class="dash-item"><a href="#" onclick="loadNote(\'' + escAttr(f.path) + '\');return false">' +
                        escHtml(stripMdExt(f.title || f.path.split('/').pop())) + '</a></div>';
                });
                html += '</div>';
            }

            // Active
            if (data.active && data.active.length > 0) {
                html += '<div class="dash-card"><h3>Active</h3>';
                data.active.forEach(f => {
                    html += '<div class="dash-item"><a href="#" onclick="loadNote(\'' + escAttr(f.path) + '\');return false">' +
                        escHtml(stripMdExt(f.title || f.path.split('/').pop())) + '</a></div>';
                });
                html += '</div>';
            }

            // Debug
            if (data.debug && data.debug.length > 0) {
                html += '<div class="dash-card"><h3>Debug Notes</h3>';
                data.debug.forEach(f => {
                    html += '<div class="dash-item"><a href="#" onclick="loadNote(\'' + escAttr(f.path) + '\');return false">' +
                        escHtml(stripMdExt(f.title || f.path.split('/').pop())) + '</a></div>';
                });
                html += '</div>';
            }

            // Tags
            if (data.tags && data.tags.length > 0) {
                html += '<div class="dash-card"><h3>Top Tags</h3>';
                data.tags.forEach(t => {
                    html += '<span class="dash-tag" onclick="showTagFiles(\'' + escAttr(t.tag) + '\')">' + escHtml(t.tag) + ' (' + t.count + ')</span>';
                });
                html += '</div>';
            }

            // Canvas
            if (data.canvas && data.canvas.length > 0) {
                html += '<div class="dash-card"><h3>Canvas</h3>';
                data.canvas.forEach(f => {
                    html += '<div class="dash-canvas-item"><a href="#" onclick="loadNote(\'' + escAttr(f.path) + '\');return false">' +
                        escHtml(f.title || f.path.split('/').pop()) + '</a></div>';
                });
                html += '</div>';
            }

            html += '</div>';
            return html;
        } catch(e) {
            return '<div class="error-page"><h2>Error</h2><p>' + escHtml(e.message) + '</p></div>';
        }
    }

    // ==================== Init ====================
    loadTree();
    loadDashboard().then(html => {
        if (!currentPath) {
            $('content').innerHTML = '<div class="content-inner">' +
                '<div style="display:flex;align-items:center;gap:12px;margin-bottom:24px">' +
                '<div class="welcome-icon" style="margin-bottom:0"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg></div>' +
                '<div><h2 style="margin:0;font-size:1.5rem">Dashboard</h2><p style="color:var(--text-muted);margin:0;font-size:14px">Vault overview</p></div></div>' +
                html + '</div>';
        }
    });
    </script>
</body>
</html>`
