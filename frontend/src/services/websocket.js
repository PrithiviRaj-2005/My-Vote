// WebSocket connection helper for PulseVote
export function createPollWebSocket(shareCode, onMessage, onError, onClose) {
  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
  
  // Convert http/https URL to ws/wss URL
  let wsUrl = apiUrl.replace(/^http/, 'ws');
  wsUrl = `${wsUrl}/ws/polls/${shareCode}`;

  let socket = null;
  let shouldReconnect = true;
  let reconnectTimer = null;

  function connect() {
    try {
      socket = new WebSocket(wsUrl);

      socket.onopen = () => {
        console.log(`[WebSocket] Connected to live poll stream [${shareCode}]`);
      };

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (onMessage) onMessage(data);
        } catch (e) {
          console.error('[WebSocket] Error parsing message payload:', e, event.data);
        }
      };

      socket.onerror = (err) => {
        console.warn(`[WebSocket] Error on poll [${shareCode}]:`, err);
        if (onError) onError(err);
      };

      socket.onclose = (event) => {
        console.log(`[WebSocket] Connection closed for poll [${shareCode}]`);
        if (onClose) onClose(event);

        if (shouldReconnect) {
          reconnectTimer = setTimeout(() => {
            console.log(`[WebSocket] Attempting to reconnect to [${shareCode}]...`);
            connect();
          }, 3000);
        }
      };
    } catch (e) {
      console.error('[WebSocket] Setup exception:', e);
      if (shouldReconnect) {
        reconnectTimer = setTimeout(connect, 3000);
      }
    }
  }

  connect();

  return {
    disconnect: () => {
      shouldReconnect = false;
      if (reconnectTimer) clearTimeout(reconnectTimer);
      if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
        socket.close();
      }
    },
    getSocket: () => socket,
  };
}
