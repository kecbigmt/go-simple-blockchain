package util

import(
  "fmt"
  "net"
)

func GetMyIP() (ip string, err error) {
  conn, err := net.Dial("udp", "8.8.8.8:80")
  if err != nil {
    err = fmt.Errorf("net.Dial: %s", err)
    return
  }
  defer conn.Close()

  addr := conn.LocalAddr().(*net.UDPAddr)
  ip = addr.IP.String()
  return
}
