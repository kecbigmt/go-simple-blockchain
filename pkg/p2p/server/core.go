package server

import (
  "fmt"
  "sync"
  "context"

  "github.com/kecbigmt/go-simple-blockchain/pkg/p2p/connection"
  "github.com/kecbigmt/go-simple-blockchain/pkg/util"
  "github.com/kecbigmt/go-simple-blockchain/pkg/logch"
)

type ServerCore struct {
  ServerState ServerState
  Host string
  Port int
  ConnectionManager connection.ConnectionManager
  JoinedNodeAddress string
  LogCh *logch.LogCh
}

func NewServerCore(port int, nodeAddressToJoin string) (core *ServerCore, err error) {

  if port == 0 {
    err = fmt.Errorf("invalid port number: %d", port)
    return
  }
  ip, err := util.GetMyIP()
  if err != nil {
    err = fmt.Errorf("util.GetMyIP: %s", err)
    return
  }
  cm, err := connection.NewCoreConnectionManager(ip, port)
  if err != nil {
    err = fmt.Errorf("connection.NewCoreConnectionManager: %s", err)
    return
  }
  core = &ServerCore{
    ServerState: STATE_INIT,
    Host: ip,
    Port: port,
    JoinedNodeAddress: nodeAddressToJoin,
    ConnectionManager: cm,
    LogCh: cm.LogCh(),
  }
  return
}

func (core *ServerCore) Start(ctx context.Context, wg *sync.WaitGroup) error {
  core.ServerState = STATE_STANDBY
  ctx, cancel := context.WithCancel(ctx)
  if err := core.ConnectionManager.Start(ctx, wg); err != nil {
    return fmt.Errorf("core.ConnectionManager.Start: %s", err)
  }

  wg.Add(1)
  go core.cancelHandler(ctx, cancel, wg)

  return nil
}

func (core *ServerCore) JoinNetwork() error {
  core.ServerState = STATE_CONNECTED_TO_NETWORK
  if core.JoinedNodeAddress != "" {
    if err := core.ConnectionManager.JoinNetwork(core.JoinedNodeAddress); err != nil {
      return fmt.Errorf("core.ConnectionManager.JoinNetwork: %s", err)
    }
  }
  return nil
}

func (core *ServerCore) cancelHandler(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
  defer wg.Done()
  <- ctx.Done() // Startを呼び出した階層でcancelが実行されるまで待機
  core.ServerState = STATE_SHUTTING_DOWN
  cancel() // Start以下の階層をcancel
  core.LogCh.Debugf("Closing connection...")
  if err := core.ConnectionManager.ConnectionClose(); err != nil {
    core.LogCh.Warningf("core.ConnectionManager.ConnectionClose: %s", err)
  }
  core.LogCh.Debugf("Connection closed")
  return
}

func (core *ServerCore) getMyCurrentState() ServerState {
  return core.ServerState
}
