package message

import (
  "fmt"
  "encoding/json"

  "github.com/deckarep/golang-set"
)

type Message struct{
  Protocol string `json:"protocol"`
  Version string `json:"version"`
  PayloadStr string `json:"payload"`

  MsgType MsgType `json:"msg_type"`
  Port int `json:"port"`
}

func (m Message) Payload() (s mapset.Set, errRet error) {
  switch m.MsgType {
  case MSG_CORE_LIST:
    var sl []interface{}
    if err := json.Unmarshal([]byte(m.PayloadStr), &sl); err != nil {
      errRet = fmt.Errorf("json.Unmarshal: %s", err)
      return
    }
    s = mapset.NewSetFromSlice(sl)
    return
  default:
    errRet = fmt.Errorf("invalid message type: %d", m.MsgType)
    return
  }
}
