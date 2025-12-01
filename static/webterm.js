function WebTerm(id, options = {}) {
    // Create a new terminal instance
    const term = new Terminal(options);

    // Create fit addon instance
    const fitAddon = new window.FitAddon.FitAddon();
    const webLinksAddon = new window.WebLinksAddon.WebLinksAddon();
    const webglAddon = new window.WebglAddon.WebglAddon();

    // Load addons into terminal
    term.loadAddon(fitAddon);
    term.loadAddon(webLinksAddon);
    term.loadAddon(webglAddon);

    // WebSocket connection management
    let ws = null;
    // Send terminal input to server via WebSocket
    term.send = (code, data) => {
        if (ws && ws.readyState === WebSocket.OPEN) {
            var buf = new Uint8Array(1 + data.length);
            buf[0] = code;
            buf.set(new Uint8Array(data.split('').map(c => c.charCodeAt(0))), 1);
            ws.send(buf);
        } else {
            console.log("WebSocket is not open. Unable to send data.");
        }
    }

    (() => {
        // Build WebSocket URL with filter and selected parameters
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        let url = `${protocol}//${window.location.host}${window.location.pathname}data`;

        // Connect to WebSocket endpoint
        ws = new WebSocket(url);
        ws.onopen = () => {
            // Fit terminal to container
            fitAddon.fit();
            // Send initial terminal size
            term.send(0, JSON.stringify({ cols: term.cols, rows: term.rows }));
        };
        //ws.onmessage = (event) => {
            //term.write(event.data);
        //};
        ws.onerror = (error) => {
            console.log("WebSocket error:", error);
            term.writeln('\x1b[31mConnection error.\x1b[0m');
        };
        ws.onclose = () => {
            term.writeln('\x1b[33mConnection closed.\x1b[0m');
        };
    })();

    // Create fit addon instance
    const attachAddon = new window.AttachAddon.AttachAddon(ws);

    // Load addon into terminal
    term.loadAddon(attachAddon);

    // Attach terminal to the DOM
    let container = document.getElementById(id);
    term.open(container);
    container.style.backgroundColor = options.theme.background;

    // Refit on window resize with debounce
    let resizeTimeout;
    window.addEventListener('resize', () => {
        clearTimeout(resizeTimeout);
        resizeTimeout = setTimeout(() => {
            fitAddon.fit();
        }, 100);
    });
    // Cleanup on page unload
    window.addEventListener('beforeunload', () => {
        if (ws) {
            ws.close();
        }
    });

    term.onData((data) => {
        term.send(1, data);
    });
    term.onResize((size) => {
        term.send(0, JSON.stringify({ cols: size.cols, rows: size.rows }));
    });

    return term;
}