package message

const (
  PROTOCOL_NAME = "simple_bitcoin_protocol"
  MY_VERSION = "0.1.0"
)

type MsgType int

const (
  MSG_ADD MsgType = iota
  MSG_REMOVE
  MSG_CORE_LIST
  MSG_REQUEST_CORE_LIST
  MSG_PING
  MSG_ADD_AS_EDGE
  MSG_REMOVE_AS_EDGE
)

type Reason int

const (
  ERR_PROTOCOL_UNMATCH Reason = iota
  ERR_VERSION_UNMATCH
  OK_WITH_PAYLOAD
  OK_WITHOUT_PAYLOAD
)
