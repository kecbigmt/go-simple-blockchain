package message

import (
  "fmt"
  "testing"
)

func TestMessagePayload(t *testing.T) {
  str1 := "192.168.3.11:80"
  str2 := "192.168.3.12:8080"

  m := Message{
    MsgType: MSG_CORE_LIST,
    PayloadStr: fmt.Sprintf(`["%s", "%s"]`, str1, str2),
  }
  s, err := m.Payload()
  if err != nil {
    t.Fatalf("m.Payload: %s", err)
  }
  if s.Cardinality() != 2 {
    t.Fatalf("payload length is invalid: %d", s.Cardinality())
  }
  s.Each(func(elem interface{}) bool {
    str, ok := elem.(string)
    if !ok {
      t.Fatalf("invalid element: %v", elem)
      return true
    }
    if str != str1 && str != str2 {
      t.Fatalf("invalid element: %s", str)
      return true
    }
    return false
  })
}
