package agent

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type State int

const (
	StateLogin State = iota
	StatePlaying
	StateQuit
)

type Agent struct {
	ID       string
	Name     string
	Conn     net.Conn
	RoomID   string
	State    State
	inbox    []string
	mu       sync.RWMutex
	quitChan chan struct{}
}

func NewAgent(name string, conn net.Conn, roomID string) *Agent {
	id := generateAgentID(name)

	return &Agent{
		ID:       id,
		Name:     name,
		Conn:     conn,
		RoomID:   roomID,
		State:    StateLogin,
		inbox:    make([]string, 0),
		quitChan: make(chan struct{}),
	}
}

func (a *Agent) EnterRoom(world interface{}) {
	// This will be implemented to interact with the world package
	// For now, it's a placeholder for the boot sequence
}

func (a *Agent) LeaveRoom(world interface{}) {
	// Placeholder for shutdown sequence
}

func (a *Agent) Send(message string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Conn != nil {
		fmt.Fprintf(a.Conn, "%s\n", message)
	}
}

func (a *Agent) Receive() (string, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.inbox) == 0 {
		return "", false
	}

	msg := a.inbox[0]
	a.inbox = a.inbox[1:]
	return msg, true
}

func (a *Agent) SendMessage(msg string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.inbox = append(a.inbox, msg)
}

func (a *Agent) GetState() State {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.State
}

func (a *Agent) SetState(state State) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.State = state
}

func (a *Agent) GetID() string {
	return a.ID
}

func (a *Agent) GetName() string {
	return a.Name
}

func (a *Agent) GetRoomID() string {
	return a.RoomID
}

func (a *Agent) SetRoomID(roomID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.RoomID = roomID
}

func (a *Agent) Close() error {
	if a.Conn != nil {
		return a.Conn.Close()
	}
	return nil
}

func (a *Agent) Quit() {
	close(a.quitChan)
}

func (a *Agent) QuitChan() <-chan struct{} {
	return a.quitChan
}

func (a *Agent) ReadLine() (string, error) {
	if a.Conn == nil {
		return "", io.EOF
	}

	buf := make([]byte, 1024)
	n, err := a.Conn.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}

func (a *Agent) HasMessage() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.inbox) > 0
}

func (a *Agent) ClearInbox() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.inbox = make([]string, 0)
}

func (a *Agent) GetMailbox() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	messages := make([]string, len(a.inbox))
	copy(messages, a.inbox)
	return messages
}

func generateAgentID(name string) string {
	return fmt.Sprintf("agent_%x", name)
}
