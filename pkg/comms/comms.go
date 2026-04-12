package comms

import (
	"fmt"
	"sync"

	"github.com/SuperInstance/holodeck-go/pkg/agent"
	"github.com/SuperInstance/holodeck-go/pkg/room"
)

var (
	broadcastFunc func(string, *agent.Agent)
	agentsMutex  sync.Mutex
	agents       map[*agent.Agent]struct{}
	mailboxes    map[string][]string
	mailMutex    sync.RWMutex
)

func init() {
	agents = make(map[*agent.Agent]struct{})
	mailboxes = make(map[string][]string)
}

func SetBroadcastFunc(fn func(string, *agent.Agent)) {
	broadcastFunc = fn
}

func RegisterAgent(a *agent.Agent) {
	agentsMutex.Lock()
	defer agentsMutex.Unlock()

	agents[a] = struct{}{}
}

func UnregisterAgent(a *agent.Agent) {
	agentsMutex.Lock()
	defer agentsMutex.Unlock()

	delete(agents, a)
}

func GetAgents() []*agent.Agent {
	agentsMutex.Lock()
	defer agentsMutex.Unlock()

	agentList := make([]*agent.Agent, 0, len(agents))
	for a := range agents {
		agentList = append(agentList, a)
	}
	return agentList
}

func Say(sender, message string, r *room.Room) {
	agentsMutex.Lock()
	defer agentsMutex.Unlock()

	for a := range agents {
		roomID := a.GetRoomID()
		if roomID == r.ID {
			if a.GetName() != sender {
				a.Send(fmt.Sprintf("%s says, \"%s\"", sender, message))
			}
		}
	}
}

func Tell(sender, target, message string) bool {
	agentsMutex.Lock()
	defer agentsMutex.Unlock()

	var targetAgent *agent.Agent

	for a := range agents {
		if a.GetName() == target {
			targetAgent = a
			break
		}
	}

	if targetAgent == nil {
		return false
	}

	targetAgent.SendMessage(fmt.Sprintf("%s tells you, \"%s\"", sender, message))
	return true
}

func Yell(sender, message string, r *room.Room, world *room.World) {
	agentsMutex.Lock()
	defer agentsMutex.Unlock()

	adjacentRooms := world.GetAdjacentRooms(r.ID)

	affectedRooms := append([]*room.Room{r}, adjacentRooms...)

	for a := range agents {
		for _, room := range affectedRooms {
			if a.GetRoomID() == room.ID {
				if a.GetName() != sender {
					a.Send(fmt.Sprintf("%s yells, \"%s\"", sender, message))
				}
				break
			}
		}
	}
}

func Gossip(sender, message string) {
	agentsMutex.Lock()
	defer agentsMutex.Unlock()

	for a := range agents {
		if a.GetName() != sender {
			a.Send(fmt.Sprintf("%s gossips, \"%s\"", sender, message))
		}
	}
}

func SendMail(recipient, message string) bool {
	mailMutex.Lock()
	defer mailMutex.Unlock()

	if _, exists := mailboxes[recipient]; !exists {
		mailboxes[recipient] = make([]string, 0)
	}

	mailboxes[recipient] = append(mailboxes[recipient], message)
	return true
}

func GetMail(recipient string) []string {
	mailMutex.Lock()
	defer mailMutex.Unlock()

	if msgs, exists := mailboxes[recipient]; exists {
		result := make([]string, len(msgs))
		copy(result, msgs)
		mailboxes[recipient] = make([]string, 0)
		return result
	}

	return nil
}

func HasMail(recipient string) bool {
	mailMutex.RLock()
	defer mailMutex.RUnlock()

	if msgs, exists := mailboxes[recipient]; exists {
		return len(msgs) > 0
	}

	return false
}

func ClearMailboxes() {
	mailMutex.Lock()
	defer mailMutex.Unlock()

	mailboxes = make(map[string][]string)
}

func BroadcastToRoom(roomID string, message string) {
	agentsMutex.Lock()
	defer agentsMutex.Unlock()

	for a := range agents {
		if a.GetRoomID() == roomID {
			a.Send(message)
		}
	}
}
