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

type coreConnectionManager struct {
  host string
  port int
  selfAddress string
  joinedNodeAddress string
  coreNodeSet mapset.Set
  edgeNodeSet mapset.Set
  messageManager *msg.MessageManager
  logCh *logch.LogCh

  addMsgBytes []byte
}

func NewCoreConnectionManager(host string, port int) (cm ConnectionManager, err error) {
  selfAddr := fmt.Sprintf("%s:%d", host, port)
  logCh := logch.NewLogCh()
  mm := msg.NewManager()
  addMsgBytes, err := mm.Build(msg.MSG_ADD, port, nil)
  if err != nil {
    err = fmt.Errorf("cm.messageManager.Build: %s", err)
    return
  }
  cm = &coreConnectionManager{
    host: host,
    port: port,
    selfAddress: selfAddr,
    coreNodeSet: mapset.NewSet(selfAddr),
    edgeNodeSet: mapset.NewSet(),
    messageManager: mm,
    logCh: logCh,

    addMsgBytes: addMsgBytes,
  }
  return
}

func (cm *coreConnectionManager) Start(ctx context.Context, wg *sync.WaitGroup) error {
  ctx, cancel := context.WithCancel(ctx)
  go cm.waitForAccess()
  cm.logCh.Infof("Connection opened as: %s", cm.selfAddress)

  wg.Add(1)
  go cm.checkCoreNodeConnection(ctx, wg)
  wg.Add(1)
  go cm.checkEdgeNodeConnection(ctx, wg)

  wg.Add(1)
  go cm.cancelHandler(ctx, cancel, wg)

  return nil
}

func (cm *coreConnectionManager) JoinNetwork(address string) error {
  cm.joinedNodeAddress = address
  if err := cm.sendMsg(address, cm.addMsgBytes); err != nil {
    return fmt.Errorf("cm.sendMsg: %s", err)
  }
  return nil
}

func (cm *coreConnectionManager) ConnectionClose() error {
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

func (cm *coreConnectionManager) LogCh() *logch.LogCh {
  return cm.logCh
}

// 指定されたノードに対してメッセージを送信するメソッド。
func (cm *coreConnectionManager) sendMsg(address string, msgBytes []byte) error {
  conn, err := net.Dial("tcp", address)
  if err != nil {
    cm.removeCoreNode(address)
    return fmt.Errorf("Connection failed for node: %s\n", address)
  }
  defer conn.Close()
  conn.Write(msgBytes)
  return nil
}

// Coreノードリストに登録されている全てのノードに対して同じメッセージをブロードキャストするメソッド。
func (cm *coreConnectionManager) sendMsgToAllPeer(msgBytes []byte) {
  it := cm.coreNodeSet.Iterator()
  for elem := range it.C {
    address, ok := elem.(string)
    if !ok {
      cm.logCh.Warningf("invalid address: %v\n", elem)
      continue
    }
    if address != cm.selfAddress {
      if err := cm.sendMsg(address, msgBytes); err != nil {
        cm.logCh.Warningf("faild to send message: cm.sendMsg: %s\n", err)
      }
    }
  }
}

func (cm *coreConnectionManager) cancelHandler(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
  defer wg.Done()
  <- ctx.Done()
  cancel()
}

func (cm *coreConnectionManager) addCoreNode(node string) {
  if cm.coreNodeSet.Add(node) {
    cm.logCh.Infof("Core node added: %s\n", node)
  }
}

func (cm *coreConnectionManager) removeCoreNode(node string) {
  if cm.coreNodeSet.Contains(node) {
    cm.coreNodeSet.Remove(node)
    cm.logCh.Infof("Core node removed: %s\n", node)
  }
}

func (cm *coreConnectionManager) addEdgeNode(node string) {
  if cm.edgeNodeSet.Add(node) {
    cm.logCh.Infof("Edge node added: %s\n", node)
  }
}

func (cm *coreConnectionManager) removeEdgeNode(node string) {
  if cm.edgeNodeSet.Contains(node) {
    cm.edgeNodeSet.Remove(node)
    cm.logCh.Infof("Edge node removed: %s\n", node)
  }
}

// nodeからのメッセージを待ち受けるメソッド。メッセージが来たら別ゴルーチン（cm.handleMessage）に処理を依頼。
func (cm *coreConnectionManager) waitForAccess() {
  listen, err := net.Listen("tcp", cm.selfAddress)
  if err != nil {
    cm.logCh.Errorf("net.Listen: %s", err)
    return
  }
  for {
    cm.logCh.Infof("Waiting for the connection ...")
    conn, err := listen.Accept() //接続があるまで待機
    if err != nil {
      continue
    }
    go cm.handleMessage(conn)
  }
}

// 受信したメッセージを確認して、内容に応じた処理を行うメソッド。
func (cm *coreConnectionManager) handleMessage(conn net.Conn) {
  defer conn.Close()

  buf := make([]byte, 1024)
  length, err := conn.Read(buf)
  if err != nil {
    log.Fatal(err)
  }

  tcpAddr := conn.RemoteAddr().(*net.TCPAddr)
  nodeIP := tcpAddr.IP.String()
  cm.logCh.Infof("Message received from %s: %s\n", nodeIP, string(buf[:length]))
  message, err := cm.messageManager.Parse(buf[:length])
  if err != nil {
    cm.logCh.Warningf("cm.messageManager.Parse: %s", err)
    return
  }

  node := fmt.Sprintf("%s:%d", nodeIP, message.Port)

  switch message.MsgType {
  case msg.MSG_ADD:
    cm.logCh.Infof("MSG_ADD from: %s\n", node)
    // リスト更新（Peer追加）
    cm.addCoreNode(node)
    // 最新リストのブロードキャスト
    if err := cm.broadcastCoreNodeList(); err != nil {
      cm.logCh.Errorf("failed to broadcast: cm.broadcastCoreNodeList: %s", err)
      return
    }

  case msg.MSG_ADD_AS_EDGE:
    cm.logCh.Infof("MSG_ADD_AS_EDGE from: %s\n", node)
    // リスト更新（Peer追加）
    cm.addEdgeNode(node)
    // 最新のCoreNodeListを共有
    if err := cm.sendCoreNodeList(node); err != nil {
      cm.logCh.Warningf("failed to send core node list: m.sendCoreNodeList: %s", err)
    }

  case msg.MSG_REMOVE_AS_EDGE:
    cm.logCh.Infof("MSG_REMOVE_AS_EDGE from: %s\n", node)
    // リスト更新（Peer削除）
    cm.removeEdgeNode(node)

  case msg.MSG_REMOVE:
    cm.logCh.Infof("MSG_REMOVE from: %s\n", node)
    // リスト更新（Peer削除）
    cm.removeCoreNode(node)
    // 最新リストのブロードキャスト
    if err := cm.broadcastCoreNodeList(); err != nil {
      cm.logCh.Errorf("failed to broadcast: cm.broadcastCoreNodeList: %s", err)
      return
    }

  case msg.MSG_PING:
    cm.logCh.Infof("MSG_PING from %s", node)
    return

  case msg.MSG_REQUEST_CORE_LIST:
    cm.logCh.Infof("MSG_REQUEST_CORE_LIST from: %s\n", node)
    if err := cm.sendCoreNodeList(node); err != nil {
      cm.logCh.Warningf("failed to send core node list: m.sendCoreNodeList: %s", err)
    }

  case msg.MSG_CORE_LIST:
    log.Printf("[INFO]MSG_CORE_LIST from %s\n", node)
    newSet, err := message.Payload()
    if err != nil {
      cm.logCh.Warningf("message.Payload: %s\n", err)
      return
    }

    toAdd := newSet.Difference(cm.coreNodeSet)
    toRemove := cm.coreNodeSet.Difference(newSet)
    toAdd.Each(func(elem interface{}) bool {
      cm.coreNodeSet.Add(elem)
      return false
    })
    toRemove.Each(func(elem interface{}) bool {
      cm.coreNodeSet.Remove(elem)
      return false
    })

    cm.logCh.Infof("Core node list refreshed. Current: %s\n", cm.coreNodeSet.String())

  default:
    cm.logCh.Warningf("Unexpected message type: %d\n", message.MsgType)

  }
}

// 有効coreノードを定期的にチェックするメソッド。接続されている全ノードにメッセージを送信して、応答がないノードがあればそのノードを自らの接続リストから削除する。削除後、最新のリストをネットワークにブロードキャストする。contextで定期チェックをcancelできるように実装。
func (cm *coreConnectionManager) checkCoreNodeConnection(ctx context.Context, wg *sync.WaitGroup) {
  defer wg.Done()
  msgBytes, err := cm.messageManager.Build(msg.MSG_PING, DEFAULT_PORT_TO_SEND, nil)
  if err != nil {
    cm.logCh.Errorf("cm.messageManager.Build: %s", err)
    return
  }

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
        if err := cm.sendMsg(address, msgBytes); err != nil {
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

      if changed {
        if err := cm.broadcastCoreNodeList(); err != nil {
          cm.logCh.Errorf("cm.broadcastCoreNodeList: %s", err)
          return
        }
      }
    }
  }
}

// 有効edgeノードを定期的にチェックするメソッド。応答がないノードはリストから削除する。
func (cm *coreConnectionManager) checkEdgeNodeConnection(ctx context.Context, wg *sync.WaitGroup) {
  defer wg.Done()
  msgBytes, err := cm.messageManager.Build(msg.MSG_PING, DEFAULT_PORT_TO_SEND, nil)
  if err != nil {
    cm.logCh.Errorf("cm.messageManager.Build: %s", err)
    return
  }
  for {
    select {
    case <-ctx.Done():
      cm.logCh.Debugf("Terminating checkEdgeNodeConnection goroutine...")
      return
    case <-time.Tick(PING_INTERVAL * time.Second):
      var deadEdgeNodes []string
      cm.edgeNodeSet.Each(func(elem interface{}) bool {
        address, ok := elem.(string)
        if !ok {
          cm.logCh.Warningf("invalid address: %v", elem)
          return false
        }
        if address == cm.selfAddress {
          return false
        }
        if err := cm.sendMsg(address, msgBytes); err != nil {
          deadEdgeNodes = append(deadEdgeNodes, address)
          cm.logCh.Infof("Dead edge node found: %s: %s", address, err)
        }
        return false
      })

      if len(deadEdgeNodes) > 0 {
        cm.logCh.Infof("Removing dead edge nodes ... : %s", strings.Join(deadEdgeNodes, ", "))
        for _, address := range deadEdgeNodes{
          cm.removeEdgeNode(address)
        }
      }

      cm.logCh.Infof("Current Edge List: %s", cm.edgeNodeSet.String())
    }
  }
}

// 最新のリストを特定のノードに送るメソッド。
func (cm *coreConnectionManager) sendCoreNodeList(address string) error {
  mBytes, err := cm.messageManager.Build(msg.MSG_CORE_LIST, DEFAULT_PORT_TO_SEND, cm.coreNodeSet.ToSlice())
  if err != nil {
    return fmt.Errorf("msg.messageManager.Build: %s\n", err)
  }
  if err := cm.sendMsg(address, mBytes); err != nil {
    return fmt.Errorf("cm.sendMsg: %s\n", err)
  }
  return nil
}

// 最新のリストをブロードキャストするメソッド。
func (cm *coreConnectionManager) broadcastCoreNodeList() error {
  mBytes, err := cm.messageManager.Build(msg.MSG_CORE_LIST, DEFAULT_PORT_TO_SEND, cm.coreNodeSet.ToSlice())
  if err != nil {
    return fmt.Errorf("msg.messageManager.Build: %s\n", err)
  }
  cm.sendMsgToAllPeer(mBytes)
  return nil
}
