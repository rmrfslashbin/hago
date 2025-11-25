package hago

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// wsConn manages the WebSocket connection to Home Assistant.
type wsConn struct {
	conn      *websocket.Conn
	mu        sync.Mutex
	msgID     atomic.Int64
	pending   map[int64]chan *wsResponse
	pendingMu sync.Mutex
	done      chan struct{}
	closeOnce sync.Once
}

// wsAuthMessage is sent to authenticate the WebSocket connection.
type wsAuthMessage struct {
	Type        string `json:"type"`
	AccessToken string `json:"access_token"`
}

// wsResponse represents a response from Home Assistant.
type wsResponse struct {
	ID      int64           `json:"id,omitempty"`
	Type    string          `json:"type"`
	Success bool            `json:"success,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *wsError        `json:"error,omitempty"`
	HAVersion string        `json:"ha_version,omitempty"`
	Message   string        `json:"message,omitempty"`
}

// wsError represents an error from Home Assistant.
type wsError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WebSocketError represents an error from the WebSocket API.
type WebSocketError struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *WebSocketError) Error() string {
	return fmt.Sprintf("websocket error [%s]: %s", e.Code, e.Message)
}

// connectWebSocket establishes and authenticates a WebSocket connection.
func (c *Client) connectWebSocket(ctx context.Context) error {
	c.wsMu.Lock()
	defer c.wsMu.Unlock()

	// Already connected
	if c.ws != nil {
		select {
		case <-c.ws.done:
			// Connection closed, need to reconnect
		default:
			return nil
		}
	}

	// Build WebSocket URL
	wsURL, err := c.buildWebSocketURL()
	if err != nil {
		return fmt.Errorf("build websocket URL: %w", err)
	}

	// Connect
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, wsURL, http.Header{})
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	ws := &wsConn{
		conn:    conn,
		pending: make(map[int64]chan *wsResponse),
		done:    make(chan struct{}),
	}

	// Authenticate
	if err := ws.authenticate(ctx, c.token); err != nil {
		conn.Close()
		return fmt.Errorf("websocket auth: %w", err)
	}

	// Start reader goroutine
	go ws.reader()

	c.ws = ws
	return nil
}

// buildWebSocketURL converts the REST API URL to a WebSocket URL.
func (c *Client) buildWebSocketURL() (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}

	// Convert http(s) to ws(s)
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		return "", fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	u.Path = strings.TrimSuffix(u.Path, "/") + "/api/websocket"
	return u.String(), nil
}

// authenticate performs the WebSocket authentication handshake.
func (ws *wsConn) authenticate(ctx context.Context, token string) error {
	// Read auth_required message
	var authReq wsResponse
	if err := ws.conn.ReadJSON(&authReq); err != nil {
		return fmt.Errorf("read auth_required: %w", err)
	}

	if authReq.Type != "auth_required" {
		return fmt.Errorf("expected auth_required, got %s", authReq.Type)
	}

	// Send auth message
	authMsg := wsAuthMessage{
		Type:        "auth",
		AccessToken: token,
	}
	if err := ws.conn.WriteJSON(authMsg); err != nil {
		return fmt.Errorf("write auth: %w", err)
	}

	// Read auth response
	var authResp wsResponse
	if err := ws.conn.ReadJSON(&authResp); err != nil {
		return fmt.Errorf("read auth response: %w", err)
	}

	switch authResp.Type {
	case "auth_ok":
		return nil
	case "auth_invalid":
		msg := authResp.Message
		if msg == "" {
			msg = "invalid authentication"
		}
		return fmt.Errorf("auth failed: %s", msg)
	default:
		return fmt.Errorf("unexpected auth response: %s", authResp.Type)
	}
}

// reader continuously reads messages from the WebSocket and dispatches them.
func (ws *wsConn) reader() {
	defer ws.close()

	for {
		var resp wsResponse
		if err := ws.conn.ReadJSON(&resp); err != nil {
			// Connection closed or error
			return
		}

		// Dispatch response to waiting caller
		if resp.ID != 0 {
			ws.pendingMu.Lock()
			if ch, ok := ws.pending[resp.ID]; ok {
				ch <- &resp
				delete(ws.pending, resp.ID)
			}
			ws.pendingMu.Unlock()
		}
	}
}

// sendCommand sends a command and waits for a response.
func (ws *wsConn) sendCommand(ctx context.Context, cmd any) (*wsResponse, error) {
	// Get next message ID
	id := ws.msgID.Add(1)

	// Create response channel
	respCh := make(chan *wsResponse, 1)
	ws.pendingMu.Lock()
	ws.pending[id] = respCh
	ws.pendingMu.Unlock()

	// Ensure cleanup
	defer func() {
		ws.pendingMu.Lock()
		delete(ws.pending, id)
		ws.pendingMu.Unlock()
	}()

	// Marshal command with ID
	cmdMap := make(map[string]any)
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("marshal command: %w", err)
	}
	if err := json.Unmarshal(data, &cmdMap); err != nil {
		return nil, fmt.Errorf("unmarshal command: %w", err)
	}
	cmdMap["id"] = id

	// Send command
	ws.mu.Lock()
	err = ws.conn.WriteJSON(cmdMap)
	ws.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("write command: %w", err)
	}

	// Wait for response
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ws.done:
		return nil, fmt.Errorf("websocket connection closed")
	case resp := <-respCh:
		if resp.Error != nil {
			return nil, &WebSocketError{
				Code:    resp.Error.Code,
				Message: resp.Error.Message,
			}
		}
		if !resp.Success && resp.Type == "result" {
			return nil, fmt.Errorf("command failed")
		}
		return resp, nil
	}
}

// close closes the WebSocket connection.
func (ws *wsConn) close() {
	ws.closeOnce.Do(func() {
		close(ws.done)
		ws.conn.Close()
	})
}

// CloseWebSocket closes the WebSocket connection if open.
func (c *Client) CloseWebSocket() {
	c.wsMu.Lock()
	defer c.wsMu.Unlock()

	if c.ws != nil {
		c.ws.close()
		c.ws = nil
	}
}

// wsCommand sends a WebSocket command and returns the result.
func (c *Client) wsCommand(ctx context.Context, cmd any, result any) error {
	// Ensure connected
	if err := c.connectWebSocket(ctx); err != nil {
		return err
	}

	// Send command
	resp, err := c.ws.sendCommand(ctx, cmd)
	if err != nil {
		return err
	}

	// Decode result if provided
	if result != nil && len(resp.Result) > 0 {
		if err := json.Unmarshal(resp.Result, result); err != nil {
			return fmt.Errorf("decode result: %w", err)
		}
	}

	return nil
}
