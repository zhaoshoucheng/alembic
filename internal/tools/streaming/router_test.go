package streaming

import (
	"testing"
)

type testPusher struct {
	events []ToolDeltaEvent
}

func (p *testPusher) PushDeltaEvent(e ToolDeltaEvent) error {
	p.events = append(p.events, e)
	return nil
}

func TestRouter_SingleToolLifecycle(t *testing.T) {
	router := NewStreamingRouter()
	router.Register(NewTerminalStreamHandler("shell_exec"))
	pusher := &testPusher{}

	// tool start with first chunk
	err := router.ProcessChunk(DeltaToolCall{
		ID:       "call_1",
		Name:     "shell_exec",
		ArgChunk: `{"comma`,
	}, pusher)
	if err != nil {
		t.Fatal(err)
	}

	// argument delta
	err = router.ProcessChunk(DeltaToolCall{
		ID:       "call_1",
		ArgChunk: `nd": "ls -la"`,
	}, pusher)
	if err != nil {
		t.Fatal(err)
	}

	// closing brace
	err = router.ProcessChunk(DeltaToolCall{
		ID:       "call_1",
		ArgChunk: `}`,
	}, pusher)
	if err != nil {
		t.Fatal(err)
	}

	// finish
	err = router.FinishCurrentTool(pusher)
	if err != nil {
		t.Fatal(err)
	}

	// check events: init, delta(s), argumentsFinished
	if len(pusher.events) < 3 {
		t.Fatalf("expected at least 3 events, got %d", len(pusher.events))
	}
	if pusher.events[0].DeltaProgress != DeltaInit {
		t.Fatalf("first event should be init, got %s", pusher.events[0].DeltaProgress)
	}
	last := pusher.events[len(pusher.events)-1]
	if last.DeltaProgress != DeltaArgumentsFinished {
		t.Fatalf("last event should be argumentsFinished, got %s", last.DeltaProgress)
	}

	// check finished tools
	if len(router.FinishedTools) != 1 {
		t.Fatalf("expected 1 finished tool, got %d", len(router.FinishedTools))
	}
	ft := router.FinishedTools[0]
	if ft.ToolName != "shell_exec" {
		t.Fatalf("expected shell_exec, got %s", ft.ToolName)
	}
	if ft.Params["command"] != "ls -la" {
		t.Fatalf("expected command=ls -la, got %v", ft.Params["command"])
	}
}

func TestRouter_UnknownToolIgnored(t *testing.T) {
	router := NewStreamingRouter()
	router.Register(NewTerminalStreamHandler("shell_exec"))
	pusher := &testPusher{}

	err := router.ProcessChunk(DeltaToolCall{
		ID:       "call_1",
		Name:     "unknown_tool",
		ArgChunk: `{"x": 1}`,
	}, pusher)
	if err != nil {
		t.Fatal(err)
	}
	if len(pusher.events) != 0 {
		t.Fatalf("expected no events for unknown tool, got %d", len(pusher.events))
	}
}

func TestRouter_Abort(t *testing.T) {
	router := NewStreamingRouter()
	router.Register(NewTerminalStreamHandler("shell_exec"))
	pusher := &testPusher{}

	router.ProcessChunk(DeltaToolCall{ID: "call_1", Name: "shell_exec", ArgChunk: `{"command": "ls"`}, pusher)
	err := router.Abort("user cancelled", pusher)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, e := range pusher.events {
		if e.DeltaProgress == DeltaRollback {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected rollback event")
	}
}

func TestRouter_ToolSwitch(t *testing.T) {
	router := NewStreamingRouter()
	router.Register(NewTerminalStreamHandler("shell_exec"))
	router.Register(NewSearchStreamHandler("search"))
	pusher := &testPusher{}

	// first tool
	router.ProcessChunk(DeltaToolCall{ID: "call_1", Name: "shell_exec", ArgChunk: `{"command": "ls"}`}, pusher)

	// switch to second tool — should auto-finish first
	router.ProcessChunk(DeltaToolCall{ID: "call_2", Name: "search", ArgChunk: `{"query": "hello"}`}, pusher)
	router.FinishCurrentTool(pusher)

	if len(router.FinishedTools) != 2 {
		t.Fatalf("expected 2 finished tools, got %d", len(router.FinishedTools))
	}
	if router.FinishedTools[0].ToolName != "shell_exec" {
		t.Fatalf("first finished should be shell_exec, got %s", router.FinishedTools[0].ToolName)
	}
	if router.FinishedTools[1].ToolName != "search" {
		t.Fatalf("second finished should be search, got %s", router.FinishedTools[1].ToolName)
	}
}
