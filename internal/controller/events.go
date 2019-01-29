package controller

import (
	"encoding/json"
)

type EventMap map[string]chan json.RawMessage

func (c *Controller) OnMessage(msgType string) <-chan json.RawMessage {
	ch := make(chan json.RawMessage)
	c.events[msgType] = ch
	return ch
}

func (c *Controller) HandleMessage(msg []byte) {
	var data json.RawMessage
	packet := Packet{
		Data: &data,
	}

	// Decode json
	err := json.Unmarshal(msg, &packet)
	if err != nil {
		//log.Print("Error parsing JSONPacket: ", err)
		return
	}

	// Find channel and dispatch data
	ch, ok := c.events[packet.Type]
	if ok {
		ch <- data
	}
}
