package command

import (
	"fmt"
	"strings"

	"github.com/SuperInstance/holodeck-go/pkg/agent"
	"github.com/SuperInstance/holodeck-go/pkg/comms"
	"github.com/SuperInstance/holodeck-go/pkg/room"
)

type Handler func(world *room.World, a *agent.Agent, args []string) string

var handlers map[string]Handler

func init() {
	handlers = make(map[string]Handler)

	handlers["look"] = handleLook
	handlers["go"] = handleGo
	handlers["say"] = handleSay
	handlers["tell"] = handleTell
	handlers["who"] = handleWho
	handlers["yell"] = handleYell
	handlers["gossip"] = handleGossip
	handlers["note"] = handleNote
	handlers["read"] = handleRead
	handlers["help"] = handleHelp
}

func GetHandler(cmd string) Handler {
	return handlers[cmd]
}

func handleLook(world *room.World, a *agent.Agent, args []string) string {
	r := world.GetRoom(a.GetRoomID())
	if r == nil {
		return "You are in the void.\n"
	}

	desc := r.GetDescription()

	agents := r.GetAgents()
	if len(agents) > 0 {
		desc += "\nAlso here:"
		for _, agentID := range agents {
			if agentID != a.GetID() {
				desc += " " + agentID
			}
		}
		desc += "\n"
	}

	return desc + "\n"
}

func handleGo(world *room.World, a *agent.Agent, args []string) string {
	if len(args) == 0 {
		return "Go where?\n"
	}

	direction := strings.ToLower(args[0])

	r := world.GetRoom(a.GetRoomID())
	if r == nil {
		return "You are lost in the void.\n"
	}

	exits := r.GetExits()
	roomID, exists := exits[direction]
	if !exists {
		return fmt.Sprintf("You cannot go %s from here.\n", direction)
	}

	newRoom := world.GetRoom(roomID)
	if newRoom == nil {
		return "That exit leads nowhere.\n"
	}

	oldRoom := world.GetRoom(a.GetRoomID())
	oldRoom.RemoveAgent(a.GetID())

	a.SetRoomID(roomID)
	newRoom.AddAgent(a.GetID())

	result := fmt.Sprintf("You go %s.\n", direction)
	result += newRoom.GetDescription() + "\n"

	return result
}

func handleSay(world *room.World, a *agent.Agent, args []string) string {
	if len(args) == 0 {
		return "Say what?\n"
	}

	message := strings.Join(args, " ")

	r := world.GetRoom(a.GetRoomID())
	if r == nil {
		return "You are in the void.\n"
	}

	comms.Say(a.GetName(), message, r)

	return fmt.Sprintf("You say, \"%s\"\n", message)
}

func handleTell(world *room.World, a *agent.Agent, args []string) string {
	if len(args) < 2 {
		return "Tell whom what? Usage: tell <agent> <message>\n"
	}

	targetName := args[0]
	message := strings.Join(args[1:], " ")

	success := comms.Tell(a.GetName(), targetName, message)
	if !success {
		return fmt.Sprintf("%s is not here.\n", targetName)
	}

	return fmt.Sprintf("You tell %s, \"%s\"\n", targetName, message)
}

func handleWho(world *room.World, a *agent.Agent, args []string) string {
	rooms := world.GetAllRooms()

	result := "Agents in the world:\n"

	for _, r := range rooms {
		agents := r.GetAgents()
		if len(agents) > 0 {
			result += fmt.Sprintf("  %s:", r.Name)
			for _, agentID := range agents {
				result += " " + agentID
			}
			result += "\n"
		}
	}

	return result + "\n"
}

func handleYell(world *room.World, a *agent.Agent, args []string) string {
	if len(args) == 0 {
		return "Yell what?\n"
	}

	message := strings.Join(args, " ")

	r := world.GetRoom(a.GetRoomID())
	if r == nil {
		return "You are in the void.\n"
	}

	comms.Yell(a.GetName(), message, r, world)

	return fmt.Sprintf("You yell, \"%s\"\n", message)
}

func handleGossip(world *room.World, a *agent.Agent, args []string) string {
	if len(args) == 0 {
		return "Gossip what?\n"
	}

	message := strings.Join(args, " ")

	comms.Gossip(a.GetName(), message)

	return fmt.Sprintf("You gossip, \"%s\"\n", message)
}

func handleNote(world *room.World, a *agent.Agent, args []string) string {
	if len(args) == 0 {
		return "Write what on the wall? Usage: note <message>\n"
	}

	message := strings.Join(args, " ")

	r := world.GetRoom(a.GetRoomID())
	if r == nil {
		return "You are in the void.\n"
	}

	r.AddNote(fmt.Sprintf("%s: %s", a.GetName(), message))

	return "You write a note on the wall.\n"
}

func handleRead(world *room.World, a *agent.Agent, args []string) string {
	r := world.GetRoom(a.GetRoomID())
	if r == nil {
		return "You are in the void.\n"
	}

	notes := r.GetNotes()
	if len(notes) == 0 {
		return "There are no notes on the wall.\n"
	}

	result := "Notes on the wall:\n"
	for _, note := range notes {
		result += "  " + note + "\n"
	}

	return result + "\n"
}

func handleHelp(world *room.World, a *agent.Agent, args []string) string {
	return `Available commands:
  look           - Look around the current room
  go <direction> - Move in a direction (north, south, east, west)
  say <message>  - Say something to everyone in the room
  tell <agent> <message> - Send a private message to another agent
  who            - See all agents in the world
  yell <message> - Yell to adjacent rooms
  gossip <msg>   - Broadcast a message to the entire world
  note <message> - Write a note on the wall
  read           - Read notes on the wall
  quit           - Disconnect from the server

`
}
