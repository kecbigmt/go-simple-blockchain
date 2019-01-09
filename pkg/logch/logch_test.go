package logch_test

import (
  "fmt"
  "testing"

  "github.com/kecbigmt/go-simple-bitcoin/pkg/logch"
)

const (
  str1 = "foo"
  str2 = "bar"
)

func TestInfo(t *testing.T) {
  ch := logch.NewLogCh()
  go func(){
    ch.Infof("%s %s", str1, str2)
    ch.InfoClose()
  }()
  var cnt int
  for v := range ch.InfoOut() {
    cnt++
    if v != fmt.Sprintf("%s %s", str1, str2) {
      t.Fatalf("invalid log: %s", v)
    }
  }
  if cnt != 1 {
    t.Fatalf("invalid log count: %d", cnt)
  }
}

func TestDebug(t *testing.T) {
  ch := logch.NewLogCh()
  go func(){
    ch.Debugf("%s %s", str1, str2)
    ch.DebugClose()
  }()
  var cnt int
  for v := range ch.DebugOut() {
    cnt++
    if v != fmt.Sprintf("%s %s", str1, str2) {
      t.Fatalf("invalid log: %s", v)
    }
  }
  if cnt != 1 {
    t.Fatalf("invalid log count: %d", cnt)
  }
}

func TestWarning(t *testing.T) {
  ch := logch.NewLogCh()
  go func(){
    ch.Warningf("%s %s", str1, str2)
    ch.WarningClose()
  }()
  var cnt int
  for v := range ch.WarningOut() {
    cnt++
    if v != fmt.Sprintf("%s %s", str1, str2) {
      t.Fatalf("invalid log: %s", v)
    }
  }
  if cnt != 1 {
    t.Fatalf("invalid log count: %d", cnt)
  }
}

func TestError(t *testing.T) {
  ch := logch.NewLogCh()
  go func(){
    ch.Errorf("%s %s", str1, str2)
    ch.ErrorClose()
  }()
  var cnt int
  for v := range ch.ErrorOut() {
    cnt++
    if v.Error() != fmt.Sprintf("%s %s", str1, str2) {
      t.Fatalf("invalid log: %s", v)
    }
  }
  if cnt != 1 {
    t.Fatalf("invalid log count: %d", cnt)
  }
}
