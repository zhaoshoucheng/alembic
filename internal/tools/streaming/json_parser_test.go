package streaming

import (
	"testing"
)

func TestParsePartialJSON_Complete(t *testing.T) {
	val, ok := parsePartialJSON(`{"command": "ls -la", "session_id": "s1"}`)
	if !ok {
		t.Fatal("expected ok")
	}
	if val["command"] != "ls -la" {
		t.Fatalf("got command=%v", val["command"])
	}
}

func TestParsePartialJSON_TruncatedString(t *testing.T) {
	val, ok := parsePartialJSON(`{"command": "ls -l`)
	if !ok {
		t.Fatal("expected ok for truncated string")
	}
	if val["command"] != "ls -l" {
		t.Fatalf("got command=%v", val["command"])
	}
}

func TestParsePartialJSON_TruncatedObject(t *testing.T) {
	val, ok := parsePartialJSON(`{"command": "ls", "path": "/tmp"`)
	if !ok {
		t.Fatal("expected ok for truncated object")
	}
	if val["command"] != "ls" {
		t.Fatalf("got command=%v", val["command"])
	}
	if val["path"] != "/tmp" {
		t.Fatalf("got path=%v", val["path"])
	}
}

func TestParsePartialJSON_TrailingComma(t *testing.T) {
	val, ok := parsePartialJSON(`{"command": "ls",`)
	if !ok {
		t.Fatal("expected ok for trailing comma")
	}
	if val["command"] != "ls" {
		t.Fatalf("got command=%v", val["command"])
	}
}

func TestParsePartialJSON_Empty(t *testing.T) {
	_, ok := parsePartialJSON("")
	if ok {
		t.Fatal("expected not ok for empty")
	}
}

func TestParsePartialJSON_NestedObject(t *testing.T) {
	val, ok := parsePartialJSON(`{"options": {"verbose": true, "count": 5`)
	if !ok {
		t.Fatal("expected ok for nested truncation")
	}
	opts, _ := val["options"].(map[string]any)
	if opts == nil {
		t.Fatal("expected nested object")
	}
	if opts["verbose"] != true {
		t.Fatalf("got verbose=%v", opts["verbose"])
	}
}

func TestJsonStreamParser_Incremental(t *testing.T) {
	p := NewJsonStreamParser()

	r1 := p.Parse(`{"comm`)
	if !r1.Changed {
		// 可能解析不出完整 key-value，changed 可以是 false
	}

	r2 := p.Parse(`and": "ls`)
	if r2.Value != nil && r2.Value["command"] != nil {
		cmd, _ := r2.Value["command"].(string)
		if cmd != "ls" {
			t.Fatalf("got command=%v", cmd)
		}
	}

	r3 := p.Parse(` -la"}`)
	if !r3.Changed {
		t.Fatal("expected changed on complete")
	}
	if r3.Value["command"] != "ls -la" {
		t.Fatalf("got command=%v", r3.Value["command"])
	}
}

func TestJsonStreamParser_Reset(t *testing.T) {
	p := NewJsonStreamParser()
	p.Parse(`{"a": 1}`)
	p.Reset()
	if p.CurrentValue() != nil {
		t.Fatal("expected nil after reset")
	}
}

func TestRepairJSON_EscapedQuote(t *testing.T) {
	val, ok := parsePartialJSON(`{"text": "he said \"hello`)
	if !ok {
		t.Fatal("expected ok for escaped quote")
	}
	if val["text"] == nil {
		t.Fatal("expected text field")
	}
}
