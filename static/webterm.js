function WebTerm(id, options = {}) {
    // Create a new terminal instance
    const term = new Terminal(options);

    // Create fit addon instance
    const fitAddon = new window.FitAddon.FitAddon();
    const webglAddon = new window.WebglAddon.WebglAddon();

    term.loadAddon(fitAddon);
    term.loadAddon(webglAddon);

    // WebSocket connection management
    let ws = null;
    // Send terminal input to server via WebSocket
    send = (code, data) => {
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
            // Send initial terminal size
            send(0, JSON.stringify({ cols: term.cols, rows: term.rows }));
            // Notify user of successful connection
            term.writeln(`\x1b[32mConnected to webterm stream...\x1b[0m`);
            term.writeln('');
        };
        ws.onmessage = (event) => {
            term.write(event.data);
        };
        ws.onerror = (error) => {
            term.writeln('\x1b[31mConnection error.\x1b[0m', error);
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
    term.open(document.getElementById(id));

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
        send(1, data);
    });
    term.onResize((size) => {
        send(0, JSON.stringify({ cols: size.cols, rows: size.rows }));
    });

    // Fit terminal to container
    fitAddon.fit();
}