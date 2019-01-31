package controller

import (
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

// Controller : connects to a remote network controller
type WebsocketTransport struct {
	websocketClient *websocket.Conn
}

// Connect : connects to a controller server
func (c *WebsocketTransport) Connect(urlString string) error {
	url, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("URL parse failed: %s", err)
	}

	c.websocketClient, _, err = websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return fmt.Errorf("WebsocketTransport connection failed: %s", err)
	}
	defer c.websocketClient.Close()

	return nil
}

func (c *WebsocketTransport) Disconnect() error {
	err := c.websocketClient.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return fmt.Errorf("WebsocketTransport close failed: %s", err)
	}

	return nil
}

func (c *WebsocketTransport) Read(p []byte) (n int, err error) {
	// Loop and skip non-text messages
	for {
		t, message, err := c.websocketClient.ReadMessage()
		if err != nil {
			return 0, err
		}

		if t == websocket.TextMessage {
			p = message
			return len(message), nil
		}
	}
}

func (c *WebsocketTransport) Write(p []byte) (n int, err error) {
	err = c.websocketClient.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}
