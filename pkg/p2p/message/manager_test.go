package message

import (
  "testing"
)

func TestBuildParse(t *testing.T) {
  mm := NewManager()
  sl := []string{"a", "b", "c"}
  port := 50082
  jsonBytes, err := mm.Build(MSG_CORE_LIST, port, sl)
  if err != nil {
    t.Fatalf("error occured: %s", err)
  }

  message, err := mm.Parse(jsonBytes)
  if err != nil {
    t.Fatalf("error occured: %s", err)
  }
  if message.MsgType != MSG_CORE_LIST {
    t.Fatalf("MsgType should be %d, not %d", MSG_CORE_LIST, message.MsgType )
  }
  if message.Port != port {
    t.Fatalf("MsgType should be %d, not %d", port, message.Port )
  }

  s, err := message.Payload()
  if err != nil {
    t.Fatalf("error occured: %s", err)
  }
  if s.Cardinality() != 3 {
    t.Fatalf("invalid length: %d", s.Cardinality())
  }
}
