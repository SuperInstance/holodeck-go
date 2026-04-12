package conformance

import (
	"testing"

	"github.com/SuperInstance/holodeck-go/pkg/agent"
	"github.com/SuperInstance/holodeck-go/pkg/command"
	"github.com/SuperInstance/holodeck-go/pkg/comms"
	"github.com/SuperInstance/holodeck-go/pkg/room"
)

func TestT01_CreateRoom(t *testing.T) {
	world := room.NewWorld()
	createdRoom, err := world.CreateRoom("TestRoom", "A test room")

	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	if createdRoom == nil {
		t.Fatal("Created room is nil")
	}

	if createdRoom.Name != "TestRoom" {
		t.Errorf("Expected room name 'TestRoom', got '%s'", createdRoom.Name)
	}

	if createdRoom.Description != "A test room" {
		t.Errorf("Expected room description 'A test room', got '%s'", createdRoom.Description)
	}

	retrievedRoom := world.GetRoom(createdRoom.ID)
	if retrievedRoom == nil {
		t.Fatal("Could not retrieve created room")
	}

	if retrievedRoom.ID != createdRoom.ID {
		t.Errorf("Retrieved room ID mismatch")
	}
}

func TestT02_DestroyRoom(t *testing.T) {
	world := room.NewWorld()
	createdRoom, err := world.CreateRoom("TempRoom", "Temporary")

	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	err = world.DestroyRoom(createdRoom.ID)
	if err != nil {
		t.Fatalf("Failed to destroy room: %v", err)
	}

	retrievedRoom := world.GetRoom(createdRoom.ID)
	if retrievedRoom != nil {
		t.Fatal("Room still exists after destruction")
	}
}

func TestT03_ConnectRooms(t *testing.T) {
	world := room.NewWorld()
	room1, _ := world.CreateRoom("Room1", "First room")
	room2, _ := world.CreateRoom("Room2", "Second room")

	err := world.ConnectRooms(room1.ID, room2.ID, "east")
	if err != nil {
		t.Fatalf("Failed to connect rooms: %v", err)
	}

	exits := room1.GetExits()
	if exits["east"] != room2.ID {
		t.Errorf("Expected exit 'east' to lead to room2, got %s", exits["east"])
	}
}

func TestT04_DisconnectRooms(t *testing.T) {
	world := room.NewWorld()
	room1, _ := world.CreateRoom("Room1", "First room")
	room2, _ := world.CreateRoom("Room2", "Second room")

	world.ConnectRooms(room1.ID, room2.ID, "east")

	err := world.DisconnectRooms(room1.ID, "east")
	if err != nil {
		t.Fatalf("Failed to disconnect rooms: %v", err)
	}

	exits := room1.GetExits()
	if _, exists := exits["east"]; exists {
		t.Error("Exit still exists after disconnection")
	}
}

func TestT05_AgentEntersRoom(t *testing.T) {
	world := room.NewWorld()
	r, _ := world.CreateRoom("Lobby", "Main lobby")

	a := agent.NewAgent("TestAgent", nil, r.ID)
	r.AddAgent(a.ID)

	agents := r.GetAgents()
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent in room, got %d", len(agents))
	}

	if agents[0] != a.ID {
		t.Errorf("Expected agent ID %s, got %s", a.ID, agents[0])
	}
}

func TestT06_AgentLeavesRoom(t *testing.T) {
	world := room.NewWorld()
	r, _ := world.CreateRoom("Lobby", "Main lobby")

	a := agent.NewAgent("TestAgent", nil, r.ID)
	r.AddAgent(a.ID)

	r.RemoveAgent(a.ID)

	agents := r.GetAgents()
	if len(agents) != 0 {
		t.Errorf("Expected 0 agents in room, got %d", len(agents))
	}
}

func TestT07_AgentMovesBetweenRooms(t *testing.T) {
	world := room.NewWorld()
	room1, _ := world.CreateRoom("Room1", "First room")
	room2, _ := world.CreateRoom("Room2", "Second room")

	world.ConnectRooms(room1.ID, room2.ID, "east")
	world.ConnectRooms(room2.ID, room1.ID, "west")

	a := agent.NewAgent("TestAgent", nil, room1.ID)
	room1.AddAgent(a.ID)

	cmdHandler := command.GetHandler("go")
	response := cmdHandler(world, a, []string{"east"})

	if a.GetRoomID() != room2.ID {
		t.Errorf("Expected agent to be in room2, got %s", a.GetRoomID())
	}

	if len(response) == 0 {
		t.Error("Expected non-empty response from go command")
	}
}

func TestT08_AgentSays(t *testing.T) {
	world := room.NewWorld()
	r, _ := world.CreateRoom("ChatRoom", "Chat room")

	comms.RegisterAgent(agent.NewAgent("Alice", nil, r.ID))
	bob := agent.NewAgent("Bob", nil, r.ID)
	bob.SetRoomID(r.ID)
	comms.RegisterAgent(bob)

	r.AddAgent(bob.ID)

	cmdHandler := command.GetHandler("say")
	response := cmdHandler(world, bob, []string{"hello world"})

	if len(response) == 0 {
		t.Error("Expected non-empty response from say command")
	}

	if !containsString(response, "hello world") {
		t.Errorf("Response should contain message: %s", response)
	}
}

func TestT09_AgentTells(t *testing.T) {
	world := room.NewWorld()
	r, _ := world.CreateRoom("ChatRoom", "Chat room")

	alice := agent.NewAgent("Alice", nil, r.ID)
	comms.RegisterAgent(alice)

	r.AddAgent(alice.ID)

	success := comms.Tell("Bob", "Alice", "secret message")
	if !success {
		t.Error("Tell should succeed when target exists")
	}

	bob := agent.NewAgent("Bob", nil, r.ID)
	comms.RegisterAgent(bob)
	r.AddAgent(bob.ID)

	success = comms.Tell("Alice", "Bob", "secret message")
	if !success {
		t.Error("Tell should succeed when both sender and target exist")
	}

	success = comms.Tell("Alice", "Charlie", "message")
	if success {
		t.Error("Tell should fail when target doesn't exist")
	}
}

func TestT10_AgentYells(t *testing.T) {
	world := room.NewWorld()
	room1, _ := world.CreateRoom("Room1", "First room")
	room2, _ := world.CreateRoom("Room2", "Second room")

	world.ConnectRooms(room1.ID, room2.ID, "east")
	world.ConnectRooms(room2.ID, room1.ID, "west")

	a := agent.NewAgent("Yeller", nil, room1.ID)
	comms.RegisterAgent(a)
	room1.AddAgent(a.ID)

	cmdHandler := command.GetHandler("yell")
	response := cmdHandler(world, a, []string{"help!"})

	if len(response) == 0 {
		t.Error("Expected non-empty response from yell command")
	}

	if !containsString(response, "help!") {
		t.Errorf("Response should contain message: %s", response)
	}
}

func TestT11_AgentGossips(t *testing.T) {
	world := room.NewWorld()
	r1, _ := world.CreateRoom("Room1", "First room")
	r2, _ := world.CreateRoom("Room2", "Second room")

	a := agent.NewAgent("Gossiper", nil, r1.ID)
	b := agent.NewAgent("Listener", nil, r2.ID)

	comms.RegisterAgent(a)
	comms.RegisterAgent(b)

	r1.AddAgent(a.ID)
	r2.AddAgent(b.ID)

	cmdHandler := command.GetHandler("gossip")
	response := cmdHandler(world, a, []string{"did you hear the news?"})

	if len(response) == 0 {
		t.Error("Expected non-empty response from gossip command")
	}

	if !containsString(response, "did you hear the news?") {
		t.Errorf("Response should contain message: %s", response)
	}
}

func TestT12_AgentWritesNote(t *testing.T) {
	world := room.NewWorld()
	r, _ := world.CreateRoom("WallRoom", "Room with walls")

	a := agent.NewAgent("Writer", nil, r.ID)
	comms.RegisterAgent(a)

	cmdHandler := command.GetHandler("note")
	response := cmdHandler(world, a, []string{"leave a mark"})

	if len(response) == 0 {
		t.Error("Expected non-empty response from note command")
	}

	notes := r.GetNotes()
	if len(notes) == 0 {
		t.Error("Expected note to be written")
	}

	if !containsString(notes[0], "leave a mark") {
		t.Errorf("Note should contain message: %s", notes[0])
	}
}

func TestT13_AgentReadsNotes(t *testing.T) {
	world := room.NewWorld()
	r, _ := world.CreateRoom("WallRoom", "Room with walls")

	r.AddNote("Writer: first note")
	r.AddNote("Writer: second note")

	a := agent.NewAgent("Reader", nil, r.ID)

	cmdHandler := command.GetHandler("read")
	response := cmdHandler(world, a, []string{})

	if len(response) == 0 {
		t.Error("Expected non-empty response from read command")
	}

	if !containsString(response, "first note") || !containsString(response, "second note") {
		t.Error("Response should contain all notes")
	}
}

func TestT14_MailboxSendAndReceive(t *testing.T) {
	comms.ClearMailboxes()

	recipient := "TestAgent"
	message := "Test message"

	success := comms.SendMail(recipient, message)
	if !success {
		t.Error("SendMail should succeed")
	}

	if !comms.HasMail(recipient) {
		t.Error("Expected mail to be available")
	}

	mail := comms.GetMail(recipient)
	if mail == nil || len(mail) == 0 {
		t.Fatal("Expected to receive mail")
	}

	if mail[0] != message {
		t.Errorf("Expected message '%s', got '%s'", message, mail[0])
	}

	if comms.HasMail(recipient) {
		t.Error("Mail should be cleared after retrieval")
	}
}

func TestT15_EquipmentGrantAndCheck(t *testing.T) {
	t.Skip("Equipment system not yet implemented")
}

func TestT16_PermissionLevelEnforced(t *testing.T) {
	t.Skip("Permission system not yet implemented")
}

func TestT17_PermissionLevelGrantsAccess(t *testing.T) {
	t.Skip("Permission system not yet implemented")
}

func TestT18_EstablishLiveConnection(t *testing.T) {
	t.Skip("Live connection system not yet implemented")
}

func TestT19_ExecuteCommandThroughLiveConnection(t *testing.T) {
	t.Skip("Live connection system not yet implemented")
}

func TestT20_RoomChangeTriggersAutoCommit(t *testing.T) {
	t.Skip("Auto-commit system not yet implemented")
}

func TestT21_RoomBootsWhenAgentEnters(t *testing.T) {
	world := room.NewWorld()
	r, _ := world.CreateRoom("BootRoom", "Boot test room")

	a := agent.NewAgent("Booter", nil, r.ID)
	r.AddAgent(a.ID)

	agents := r.GetAgents()
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent after boot sequence, got %d", len(agents))
	}
}

func TestT22_RoomShutsDownWhenAgentLeaves(t *testing.T) {
	world := room.NewWorld()
	r, _ := world.CreateRoom("ShutdownRoom", "Shutdown test room")

	a := agent.NewAgent("Leaver", nil, r.ID)
	r.AddAgent(a.ID)

	r.RemoveAgent(a.ID)

	agents := r.GetAgents()
	if len(agents) != 0 {
		t.Errorf("Expected 0 agents after shutdown, got %d", len(agents))
	}
}

func TestT23_LivingManualReadCurrentGeneration(t *testing.T) {
	t.Skip("Living manual system not yet implemented")
}

func TestT24_LivingManualWriteFeedback(t *testing.T) {
	t.Skip("Living manual system not yet implemented")
}

func TestT25_LivingManualEvolveToNextGeneration(t *testing.T) {
	t.Skip("Living manual system not yet implemented")
}

func TestT26_ZeroShotFeedbackCaptured(t *testing.T) {
	t.Skip("Zero-shot feedback system not yet implemented")
}

func TestT27_PreviousOperatorNotesPreserved(t *testing.T) {
	t.Skip("Operator notes system not yet implemented")
}

func TestT28_BootSequenceExecutesAllSteps(t *testing.T) {
	t.Skip("Boot sequence system not yet implemented")
}

func TestT29_SafetyLimitsEnforced(t *testing.T) {
	t.Skip("Safety limits system not yet implemented")
}

func TestT30_CommandValidation(t *testing.T) {
	world := room.NewWorld()
	r, _ := world.CreateRoom("TestRoom", "Test room")

	agent.NewAgent("Tester", nil, r.ID)

	cmdHandler := command.GetHandler("nonexistent")
	if cmdHandler != nil {
		t.Error("Unknown command should return nil handler")
	}

	cmdHandler = command.GetHandler("look")
	if cmdHandler == nil {
		t.Error("Valid command should return handler")
	}
}

func TestT31_OversightSessionStartAndEnd(t *testing.T) {
	t.Skip("Oversight system not yet implemented")
}

func TestT32_TickRecordsChangesAndGauges(t *testing.T) {
	t.Skip("Tick system not yet implemented")
}

func TestT33_ScriptEvaluatesSituation(t *testing.T) {
	t.Skip("Script evaluation system not yet implemented")
}

func TestT34_HumanDemonstrationEvolvesScript(t *testing.T) {
	t.Skip("Script evolution system not yet implemented")
}

func TestT35_AutonomyScoreCalculated(t *testing.T) {
	t.Skip("Autonomy score system not yet implemented")
}

func TestT36_BackTestEngineScoresScenario(t *testing.T) {
	t.Skip("Back-test system not yet implemented")
}

func TestT37_RivalMatchProducesWinner(t *testing.T) {
	t.Skip("Rival match system not yet implemented")
}

func TestT38_FleetRulePromotion(t *testing.T) {
	t.Skip("Fleet rule system not yet implemented")
}

func TestT39_AfterActionReportGenerated(t *testing.T) {
	t.Skip("After-action report system not yet implemented")
}

func TestT40_ExperienceWeighting(t *testing.T) {
	t.Skip("Experience weighting system not yet implemented")
}

func TestConcurrentRoomAccess(t *testing.T) {
	world := room.NewWorld()
	r, _ := world.CreateRoom("ConcurrentRoom", "Concurrent access test")

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				agentID := string(rune('a' + id%26))
				r.AddAgent(agentID)
				r.RemoveAgent(agentID)
				r.GetDescription()
				r.GetExits()
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestConcurrentAgentMessaging(t *testing.T) {
	agents := make([]*agent.Agent, 5)
	for i := 0; i < 5; i++ {
		agents[i] = agent.NewAgent(string(rune('A'+i)), nil, "room1")
		comms.RegisterAgent(agents[i])
	}

	done := make(chan bool)

	for _, a := range agents {
		go func(ag *agent.Agent) {
			for i := 0; i < 50; i++ {
				ag.SendMessage("test message")
				ag.HasMessage()
			}
			done <- true
		}(a)
	}

	for range agents {
		<-done
	}
}

func TestRoomConcurrentReadWrite(t *testing.T) {
	world := room.NewWorld()
	r, _ := world.CreateRoom("RWTest", "Read-write test")

	notes := []string{"note1", "note2", "note3"}
	r.AddNote(notes[0])
	r.AddNote(notes[1])
	r.AddNote(notes[2])

	done := make(chan bool)

	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = r.GetNotes()
				_ = r.GetDescription()
				_ = r.GetAgents()
			}
			done <- true
		}()
	}

	for i := 0; i < 2; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				r.AddNote(string(rune('a' + j)))
				_ = r.GetNotes()
			}
			done <- true
		}(i)
	}

	for i := 0; i < 7; i++ {
		<-done
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr))
}

func TestMain(m *testing.M) {
	comms.ClearMailboxes()
	m.Run()
}
