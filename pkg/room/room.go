package room

import (
	"fmt"
	"sync"
)

type Room struct {
	ID          string
	Name        string
	Description string
	Exits       map[string]string
	Agents      map[string]struct{}
	Notes       []string
	mu          sync.RWMutex
}

type World struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewWorld() *World {
	return &World{
		rooms: make(map[string]*Room),
	}
}

func (w *World) CreateRoom(name, description string) (*Room, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	id := generateID(name)

	if _, exists := w.rooms[id]; exists {
		return nil, fmt.Errorf("room %s already exists", name)
	}

	room := &Room{
		ID:          id,
		Name:        name,
		Description: description,
		Exits:       make(map[string]string),
		Agents:      make(map[string]struct{}),
		Notes:       make([]string, 0),
	}

	w.rooms[id] = room
	return room, nil
}

func (w *World) DestroyRoom(id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	room, exists := w.rooms[id]
	if !exists {
		return fmt.Errorf("room %s not found", id)
	}

	room.mu.RLock()
	if len(room.Agents) > 0 {
		room.mu.RUnlock()
		return fmt.Errorf("cannot destroy room with agents present")
	}
	room.mu.RUnlock()

	for _, otherRoom := range w.rooms {
		otherRoom.mu.Lock()
		for dir, exitID := range otherRoom.Exits {
			if exitID == id {
				delete(otherRoom.Exits, dir)
			}
		}
		otherRoom.mu.Unlock()
	}

	delete(w.rooms, id)
	return nil
}

func (w *World) ConnectRooms(fromID, toID, direction string) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	from, exists := w.rooms[fromID]
	if !exists {
		return fmt.Errorf("source room %s not found", fromID)
	}

	_, exists = w.rooms[toID]
	if !exists {
		return fmt.Errorf("destination room %s not found", toID)
	}

	from.mu.Lock()
	from.Exits[direction] = toID
	from.mu.Unlock()

	return nil
}

func (w *World) DisconnectRooms(roomID, direction string) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	room, exists := w.rooms[roomID]
	if !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	_, exists = room.Exits[direction]
	if !exists {
		return fmt.Errorf("exit %s not found in room %s", direction, roomID)
	}

	delete(room.Exits, direction)
	return nil
}

func (w *World) GetRoom(id string) *Room {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.rooms[id]
}

func (w *World) GetRoomByName(name string) *Room {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, room := range w.rooms {
		room.mu.RLock()
		if room.Name == name {
			room.mu.RUnlock()
			return room
		}
		room.mu.RUnlock()
	}

	return nil
}

func (r *Room) GetDescription() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	desc := fmt.Sprintf("%s\n%s\n", r.Name, r.Description)

	if len(r.Exits) > 0 {
		desc += "Exits: "
		first := true
		for dir := range r.Exits {
			if !first {
				desc += ", "
			}
			desc += dir
			first = false
		}
		desc += "\n"
	}

	return desc
}

func (r *Room) AddAgent(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Agents[agentID] = struct{}{}
}

func (r *Room) RemoveAgent(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Agents, agentID)
}

func (r *Room) GetAgents() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]string, 0, len(r.Agents))
	for agentID := range r.Agents {
		agents = append(agents, agentID)
	}
	return agents
}

func (r *Room) AddNote(note string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Notes = append(r.Notes, note)
}

func (r *Room) GetNotes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	notes := make([]string, len(r.Notes))
	copy(notes, r.Notes)
	return notes
}

func (r *Room) GetExits() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	exits := make(map[string]string, len(r.Exits))
	for dir, roomID := range r.Exits {
		exits[dir] = roomID
	}
	return exits
}

func (w *World) GetAllRooms() []*Room {
	w.mu.RLock()
	defer w.mu.RUnlock()

	rooms := make([]*Room, 0, len(w.rooms))
	for _, room := range w.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

func (w *World) GetAdjacentRooms(roomID string) []*Room {
	w.mu.RLock()
	defer w.mu.RUnlock()

	room := w.rooms[roomID]
	if room == nil {
		return nil
	}

	room.mu.RLock()
	adjacent := make([]*Room, 0, len(room.Exits))
	for _, exitID := range room.Exits {
		if adjRoom, exists := w.rooms[exitID]; exists {
			adjacent = append(adjacent, adjRoom)
		}
	}
	room.mu.RUnlock()

	return adjacent
}

func generateID(name string) string {
	return fmt.Sprintf("%x", name)
}
