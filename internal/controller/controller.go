package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var defaultLogger = zerolog.New(ioutil.Discard)
var log = &defaultLogger

// SetLogger : sets library logger
func SetLogger(l *zerolog.Logger) {
	log = l
}

type Packet struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Controller : connects to a remote network controller
type Controller struct {
	wsClient *websocket.Conn
	events   EventMap
}

// Connect : connects to a controller server
func (c *Controller) Connect(urlString string) error {
	url, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("URL parse failed: %s", err)
	}

	log.Debug().Str("url", url.String()).Msg("initiating websocket connection")

	c.wsClient, _, err = websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %s", err)
	}
	defer c.wsClient.Close()

	log.Info().Str("url", url.String()).Msg("websocket connected")

	return nil
}

func (c *Controller) Disconnect() error {
	err := c.wsClient.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return fmt.Errorf("websocket close failed: %s", err)
	}

	return nil
}

func (c *Controller) Start(wg sync.WaitGroup, stop chan os.Signal) {
	wg.Add(1)

	go func() {
		<-stop
		c.Disconnect()
	}()

	go func() {
		defer wg.Done()

		for {
			t, message, err := c.wsClient.ReadMessage()
			if err != nil {
				log.Fatal().Err(err).Msg("websocket read error")
				return
			}

			if t == websocket.TextMessage {
				c.HandleMessage(message)
			}
		}
	}()
}

func (c *Controller) SendMessage(msgType string, data interface{}) error {
	msg := Packet{
		Type: msgType,
		Data: data,
	}

	buf, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %s", err)
	}

	err = c.wsClient.WriteMessage(websocket.TextMessage, buf)
	if err != nil {
		return fmt.Errorf("websocket write failed: %s", err)
	}

	return nil
}
