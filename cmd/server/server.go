package main

import (
  "os"
  "log"
  "flag"
  "sync"
  "context"
  "syscall"
  "os/signal"

  "github.com/kecbigmt/go-simple-blockchain/pkg/p2p/server"
)

var (
  port = flag.Int("p", 50081, "give an available port number to run the server.")
  node = flag.String("n", "", "give a core node address to join(e.g. 111.111.111.111:8080). if you don't give this flag, the server will start as a genesis node.")
)

func main() {
  flag.Parse()

  signalChan := make(chan os.Signal, 1)
  signal.Notify(signalChan, syscall.SIGINT)

  core, err := server.NewServerCore(*port, *node)
  if err != nil {
    log.Printf("server.NewServerCore: %s\n", err)
    return
  }

  ctx, cancel := context.WithCancel(context.Background())
  wg := &sync.WaitGroup{}

  log.Println("Starting node server...")
  core.Start(ctx, wg)

  if err := core.JoinNetwork(); err != nil {
    log.Printf("failed to join the network: core.JoinNetwork: %s", err)
    return
  }
  if *node != "" {
    log.Printf("Server running: host=%s, port=%d, joined_node=%s\n", core.Host, core.Port, core.JoinedNodeAddress)
  } else {
    log.Printf("Server running as a genesis node: host=%s, port=%d\n", core.Host, core.Port)
  }

  for {
    select {
    case s, ok := <- signalChan:
      if !ok {
        log.Println("Signal channel closed. Terminating main process...")
        return
      }
      switch s {
      case syscall.SIGINT:
        log.Println("SINGINT signal catched")
        go terminate(core, signalChan, cancel, wg)
      default:
        core.LogCh.Warningf("Unknown singal: %s\n", s)
      }
    case v, ok := <- core.LogCh.InfoOut():
      if ok {
        log.Printf("[INFO]%s", v)
      }
    case v, ok := <- core.LogCh.DebugOut():
      if ok {
        log.Printf("[DEBUG]%s", v)
      }
    case v, ok := <- core.LogCh.WarningOut():
      if ok {
        log.Printf("[WARNING]%s", v)
      }
    case v, ok := <- core.LogCh.ErrorOut():
      if ok {
        log.Printf("[ERROR]%s", v)
        go terminate(core, signalChan, cancel, wg)
      }
    }
  }
}

func terminate(core *server.ServerCore, signalChan chan os.Signal, cancel context.CancelFunc, wg *sync.WaitGroup){
  defer close(signalChan)
  log.Println("Terminating the server ...")
  cancel()
  wg.Wait()
  log.Println("Server terminated")
  core.LogCh.AllClose()
}
