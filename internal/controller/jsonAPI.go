package controller

import (
	"encoding/json"
	"fmt"
)

const (
	MaxPacketSize = 512
)

func NewJsonAPI(t Transport) API {
	return &JsonAPI{
		transport: t,
	}
}

type Packet struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type JsonAPI struct {
	transport Transport
	eventChan chan Event
}

func (a *JsonAPI) SendMessage(kind string, data interface{}) error {
	msg := Packet{
		Type: kind,
		Data: data,
	}

	buf, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %s", err)
	}

	n, err := a.transport.Write(buf)
	if err != nil {
		return fmt.Errorf("transport write failed: %s", err)
	}
	if n != len(buf) {
		return fmt.Errorf("transport writen %d out of %d bytes (%s)", n, len(buf), err)
	}

	return nil
}

func (a *JsonAPI) Init() error {
	return a.SendMessage("init", struct{}{})
}

func (a *JsonAPI) Authenticate(nodeID string, nodeKey string) error {
	var AuthRequest struct {
		NodeID  string `json:"nodeId"`
		NodeKey string `json:"nodeKey"`
	}

	AuthRequest.NodeID = nodeID
	AuthRequest.NodeKey = nodeKey

	return a.SendMessage("auth", AuthRequest)
}

func (a *JsonAPI) ParseMessage(p []byte) {

}

func (a *JsonAPI) Events() EventChannel {
	return a.eventChan
}

func (a *JsonAPI) Run() error {
	p := make([]byte, MaxPacketSize)
	for {
		_, err := a.transport.Read(p)
		if err != nil {
			return err
		}

		a.ParseMessage(p)
	}
}

/*
func ctrlRun(wg *sync.WaitGroup, stop signalCh) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case msg := <-ctrl.MessageChan("init"):
				ctrlOnInitResponse(msg)
			case <-stop:
				return
			}
		}
	}()
}

func ctrlOnInitResponse(msg json.RawMessage) {
	var InitResponse struct {
		NodeID  string `json:"nodeId"`
		NodeKey string `json:"nodeKey"`
	}

	err := json.Unmarshal(msg, &InitResponse)
	if err != nil {
		log.Error().Str("msg", string(msg)).Msg("Error unmarshaling InitResponse")
		return
	}

	log.Info().
		Str("NodeID", InitResponse.NodeID).
		Str("NodeKey", InitResponse.NodeKey).
		Msg("Received new node credentials")

	// Set newly acquired node id
	config.Set("NodeID", InitResponse.NodeID)
	config.Set("NodeKey", InitResponse.NodeKey)
	WriteConfig()
}

func ctrlOnAuthResponse(msg json.RawMessage) {
	var AuthResponse struct {
		Success bool `json:"success"`
	}

	err := json.Unmarshal(msg, &AuthResponse)
	if err != nil {
		log.Error().Str("msg", string(msg)).Msg("Error unmarshaling AuthResponse")
		return
	}

	if AuthResponse.Success {
		log.Info().Msg("Controller authentication successful")
	} else {
		log.Fatal().Msg("Controller authentication failed")
	}
}

func ctrlAuthRequest(nodeID string, nodeKey string) {
	var AuthRequest struct {
		NodeID  string `json:"nodeId"`
		NodeKey string `json:"nodeKey"`
	}

	AuthRequest.NodeID = nodeID
	AuthRequest.NodeKey = nodeKey

	ctrl.SendMessage("auth", AuthRequest)
}
*/
