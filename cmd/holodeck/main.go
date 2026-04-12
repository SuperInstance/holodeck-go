package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/SuperInstance/holodeck-go/pkg/agent"
	"github.com/SuperInstance/holodeck-go/pkg/command"
	"github.com/SuperInstance/holodeck-go/pkg/comms"
	"github.com/SuperInstance/holodeck-go/pkg/room"
)

const (
	defaultPort = 7777
)

var (
	world      *room.World
	agentMutex sync.Mutex
	agents     map[*agent.Agent]struct{}
)

func main() {
	agents = make(map[*agent.Agent]struct{})

	world = room.NewWorld()

	err := loadDefaultRooms()
	if err != nil {
		log.Fatalf("Failed to load rooms: %v", err)
	}

	port := defaultPort
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &port)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Holodeck listening on :%d", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection error: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func loadDefaultRooms() error {
	limbo, err := world.CreateRoom("Limbo", "A vast, misty void where agents spawn.")
	if err != nil {
		return err
	}

	atrium, err := world.CreateRoom("Atrium", "A grand hall with marble floors and high ceilings.")
	if err != nil {
		return err
	}

	err = world.ConnectRooms(limbo.ID, atrium.ID, "east")
	if err != nil {
		return err
	}

	err = world.ConnectRooms(atrium.ID, limbo.ID, "west")
	if err != nil {
		return err
	}

	corridor, err := world.CreateRoom("Corridor", "A long corridor stretching into darkness.")
	if err != nil {
		return err
	}

	err = world.ConnectRooms(atrium.ID, corridor.ID, "north")
	if err != nil {
		return err
	}

	err = world.ConnectRooms(corridor.ID, atrium.ID, "south")
	if err != nil {
		return err
	}

	return nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	fmt.Fprintf(writer, "Welcome to Holodeck.\nWhat is your name? ")
	writer.Flush()

	name, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Read error: %v", err)
		return
	}

	name = strings.TrimSpace(name)
	if name == "" {
		fmt.Fprintln(writer, "Name cannot be empty. Goodbye.")
		writer.Flush()
		return
	}

	limbo := world.GetRoomByName("Limbo")
	if limbo == nil {
		log.Printf("Failed to find Limbo room")
		return
	}

	a := agent.NewAgent(name, conn, limbo.ID)

	agentMutex.Lock()
	agents[a] = struct{}{}
	agentMutex.Unlock()

	a.EnterRoom(world)

	fmt.Fprintf(writer, "Welcome, %s.\n", name)
	fmt.Fprintln(writer, "Commands: look, go <direction>, say <message>, tell <agent> <message>, who, quit")
	writer.Flush()

	for {
		a.SetState(agent.StatePlaying)

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Connection closed: %v", err)
			break
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		cmd := strings.ToLower(parts[0])

		if cmd == "quit" {
			fmt.Fprintln(writer, "Goodbye.")
			writer.Flush()
			break
		}

		cmdHandler := command.GetHandler(cmd)
		if cmdHandler == nil {
			fmt.Fprintf(writer, "Unknown command: %s\n", cmd)
			writer.Flush()
			continue
		}

		response := cmdHandler(world, a, parts[1:])

		fmt.Fprint(writer, response)
		writer.Flush()

		if a.GetState() == agent.StateQuit {
			break
		}
	}

	a.LeaveRoom(world)

	agentMutex.Lock()
	delete(agents, a)
	agentMutex.Unlock()

	log.Printf("Agent %s disconnected", name)
}

func Broadcast(message string, sender *agent.Agent) {
	agentMutex.Lock()
	defer agentMutex.Unlock()

	for a := range agents {
		if a != sender {
			a.Send(message)
		}
	}
}

func init() {
	comms.SetBroadcastFunc(Broadcast)
}
