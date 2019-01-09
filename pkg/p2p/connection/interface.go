package connection

import (
  "sync"
  "context"

  "github.com/kecbigmt/go-simple-blockchain/pkg/logch"
)

type ConnectionManager interface{
  // P2Pネットワークとの接続を開始するメソッド。メッセージの待受と、接続確認のためのping送信を開始する。
  Start(ctx context.Context, wg *sync.WaitGroup) error
  // 既存ネットワークへの接続を行うメソッド。指定されたアドレスのノードにネットワークへの追加依頼を送信する。
  JoinNetwork(address string) error
  // ネットワークからの切断を行うメソッド。接続されているノード（Peer）にネットワークからの削除依頼を送信する。
  ConnectionClose() error
  // ログ送信のためのチャンネルを取得するためのメソッド。
  LogCh() *logch.LogCh
}
