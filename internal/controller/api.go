package controller

type Event interface {
}

type InitEvent struct {
	Event
	NodeID  string
	NodeKey string
}

type AuthenticationEvent struct {
	Event
	Success bool
}

type EventChannel <-chan Event

type API interface {
	Init() error
	Authenticate(nodeID string, nodeKey string) error
	Events() EventChannel
	Run() error
}
