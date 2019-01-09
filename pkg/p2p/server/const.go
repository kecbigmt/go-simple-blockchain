package server

type ServerState int

const (
  STATE_INIT ServerState = iota
  STATE_STANDBY
  STATE_CONNECTED_TO_NETWORK
  STATE_SHUTTING_DOWN
)
