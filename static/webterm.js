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

    // Attach terminal to the DOM first
    let container = document.getElementById(id);
    term.open(container);
    container.style.backgroundColor = options.theme.background;

    // WebSocket connection management
    let ws = null;
    // Send terminal input to server via WebSocket
    term.send = (code, data) => {
        if (ws && ws.readyState === WebSocket.OPEN) {
            // Properly encode UTF-8 string to bytes
            const encoder = new TextEncoder();
            const dataBytes = encoder.encode(data);
            const buf = new Uint8Array(1 + dataBytes.length);
            buf[0] = code;
            buf.set(dataBytes, 1);
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
        ws.binaryType = 'arraybuffer';
        ws.onopen = () => {
            // Fit terminal to container
            fitAddon.fit();
            // Send initial terminal size
            term.send(0, JSON.stringify({ cols: term.cols, rows: term.rows }));
            // Auto focus on terminal
            term.focus();
        };
        ws.onmessage = (event) => {
            // Handle binary data from server
            if (event.data instanceof ArrayBuffer) {
                const uint8Array = new Uint8Array(event.data);
                if (uint8Array[0] === 1) { // 1 is data message, 0 is resize window, 2 is custom message
                    term.write(uint8Array.slice(1));
                } else {
                    term.write(uint8Array);
                }
            } else {
                // Handle text data
                term.write(event.data);
            }
        };
        ws.onerror = (error) => {
            console.log("WebSocket error:", error);
            term.writeln('\x1b[31mConnection error.\x1b[0m');
        };
        ws.onclose = () => {
            term.writeln('\x1b[33mConnection closed.\x1b[0m');
        };
    })();

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