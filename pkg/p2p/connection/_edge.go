package connection

import (
  "fmt"
  "log"
  "net"
  "time"
  "sync"
  "strings"
  "context"

  "github.com/deckarep/golang-set"

  msg "github.com/kecbigmt/go-simple-blockchain/pkg/p2p/message"
  "github.com/kecbigmt/go-simple-blockchain/pkg/logch"
)

type edgeConnectionManager struct {
  host string
  port int
  selfAddress string
  joinedNodeAddress string
  coreNodeSet mapset.Set
  edgeNodeSet mapset.Set
  messageManager *msg.MessageManager
  logCh *logch.LogCh

  addMsgBytes []byte
  pingMsgBytes []byte
}

func NewEdgeConnectionManager(host string, port int) (cm ConnectionManager, err error) {
  selfAddr := fmt.Sprintf("%s:%d", host, port)
  logCh := logch.NewLogCh()
  mm := msg.NewManager()
  addMsgBytes, err := mm.Build(msg.MSG_ADD_AS_EDGE, port, nil)
  if err != nil {
    err = fmt.Errorf("cm.messageManager.Build: %s", err)
    return
  }
  pingMsgBytes, err := cm.messageManager.Build(msg.MSG_PING, DEFAULT_PORT_TO_SEND, nil)
  if err != nil {
    cm.logCh.Errorf("cm.messageManager.Build: %s", err)
    return
  }

  cm = &edgeConnectionManager{
    host: host,
    port: port,
    selfAddress: selfAddr,
    coreNodeSet: mapset.NewSet(selfAddr),
    messageManager: mm,
    logCh: logCh,
    addMsgBytes: addMsgBytes,
    pingMsgBytes: pingMsgBytes,
  }
  return
}

func (cm *edgeConnectionManager) Start(ctx context.Context, wg *sync.WaitGroup) error {
  ctx, cancel := context.WithCancel(ctx)
  go cm.waitForAccess()
  cm.logCh.Infof("Connection opened as: %s", cm.selfAddress)

  wg.Add(1)
  go cm.checkCoreNodeConnection(ctx, wg)

  wg.Add(1)
  go cm.cancelHandler(ctx, cancel, wg)

  return nil
}

func (cm *edgeConnectionManager) JoinNetwork(address string) error {
  cm.joinedNodeAddress = address
  if err := cm.sendMsg(address, cm.addMsgBytes); err != nil {
    return fmt.Errorf("cm.sendMsg: %s", err)
  }
  return nil
}

func (cm *edgeConnectionManager) ConnectionClose() error {
  if cm.joinedNodeAddress != "" {
    msgBytes, err := cm.messageManager.Build(msg.MSG_REMOVE, cm.port, nil)
    if err != nil {
      return fmt.Errorf("cm.messageManager.Build: %s", err)
    }
    if err := cm.sendMsg(cm.joinedNodeAddress, msgBytes); err != nil {
      return fmt.Errorf("cm.sendMsg: %s", err)
    }
  }
  return nil
}

func (cm *edgeConnectionManager) LogCh() *logch.LogCh {
  return cm.logCh
}

// 指定されたノードに対してメッセージを送信するメソッド。
func (cm *edgeConnectionManager) sendMsg(address string, msgBytes []byte) error {
  conn, err := net.Dial("tcp", address)
  // 接続エラー時には他のCoreノードに移住する
  if err != nil {
    cm.removeCoreNode(address)
    return fmt.Errorf("Connection failed for node: %s\n", address)
  }
  defer conn.Close()
  conn.Write(msgBytes)
  return nil
}

func (cm *edgeConnectionManager) reconnect() error {
  if cm.coreNodeSet.Cardinality() > 0 {
    it := cm.coreNodeSet.Iterator()
    for elem := range it {
      cm.logCh.Infof("Reconnecting...")
      coreNode, ok := elem.(string)
      if !ok {
        cm.logCh.Warningf("invalid address: %v", elem)
        cm.removeCoreNode(elem)
        continue
      }
      if err := cm.JoinNetwork(coreNode); err != nil {
        cm.logCh.Warningf("Failed to connect with core node: %s", coreNode)
        cm.removeCoreNode(elem)
        continue
      }
      return nil
    }
    return fmt.Errorf("Failed to reconnect with all core nodes in list")
  } else {
    return fmt.Errorf("No core node to reconnect")
  }
}

// 有効coreノードを定期的にチェックするメソッド。接続されている全ノードにメッセージを送信して、応答がないノードがあればそのノードを自らの接続リストから削除する。削除後、最新のリストをネットワークにブロードキャストする。contextで定期チェックをcancelできるように実装。
func (cm *edgeConnectionManager) checkCoreNodeConnection(ctx context.Context, wg *sync.WaitGroup) {
  defer wg.Done()

  for {
    select {
    case <- ctx.Done():
      cm.logCh.Debugf("Terminating checkCoreNodeConnection goroutine...")
      return
    case <- time.Tick(PING_INTERVAL * time.Second):
      var changed bool

      var deadCoreNodes []string
      cm.coreNodeSet.Each(func(elem interface{}) bool {
        address, ok := elem.(string)
        if !ok {
          cm.logCh.Warningf("invalid address: %v", elem)
          return false
        }
        if address == cm.selfAddress {
          return false
        }
        if err := cm.sendMsg(address, cm.pingMsgBytes); err != nil {
          deadCoreNodes = append(deadCoreNodes, address)
          cm.logCh.Infof("Dead core node found: %s: %s", address, err)
        }
        return false
      })

      if len(deadCoreNodes) > 0 {
        changed = true
        cm.logCh.Infof("Removing dead core nodes ... : %s", strings.Join(deadCoreNodes, ", "))
        for _, address := range deadCoreNodes{
          cm.removeCoreNode(address)
        }
      }

      cm.logCh.Infof("Current Core List: %s", cm.coreNodeSet.String())
    }
  }
}
