package protocol

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNDJSONRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	want := Message{Type: "invoke", ID: "inv-1", Action: "echo", Input: json.RawMessage(`{"value":1}`)}
	if err := Encode(&buf, want); err != nil {
		t.Fatal(err)
	}
	dec := NewDecoder(&buf, 0)
	var got Message
	if err := dec.Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Type != want.Type || got.ID != want.ID || got.Action != want.Action || string(got.Input) != string(want.Input) {
		t.Fatalf("round trip mismatch: %#v", got)
	}
}
