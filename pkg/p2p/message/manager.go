package message

import (
  "fmt"
  "encoding/json"

  "github.com/hashicorp/go-version"
)

type MessageManager struct {}

func NewManager() *MessageManager {
  return &MessageManager{}
}

func (mm *MessageManager) Build(msgType MsgType, port int, payload interface{}) (jsonBytes []byte, errRet error) {
  message := Message{
    Protocol: PROTOCOL_NAME,
    Version: MY_VERSION,
    MsgType: msgType,
    Port: port,
  }

  if payload != nil {
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
      errRet = fmt.Errorf("json.Marshal: %s", err)
      return
    }
    message.PayloadStr = string(payloadBytes)
  }

  jsonBytes, err := json.Marshal(message)
  if err != nil {
    errRet = fmt.Errorf("json.Marshal: %s", err)
    return
  }
  return
}

func (mm *MessageManager) Parse(jsonBytes []byte) (m Message, errRet error) {
  if err := json.Unmarshal(jsonBytes, &m); err != nil {
    errRet = fmt.Errorf("json.Unmarshal: %s", err)
    return
  }

  ver1, err := version.NewVersion(m.Version)
  if err != nil {
    errRet = fmt.Errorf("version.NewVersion: '%s': %s", m.Version, err)
    return
  }

  ver2, err := version.NewVersion(MY_VERSION)
  if err != nil {
    errRet = fmt.Errorf("version.NewVersion: '%s': %s", m.Version, err)
    return
  }

  switch {
  case m.Protocol != PROTOCOL_NAME:
    errRet = fmt.Errorf("invalid protocol: %s", m.Protocol)
    return
  case ver1.GreaterThan(ver2):
    errRet = fmt.Errorf("invalid version: %s", m.Version)
    return
  default:
    return
  }
}
