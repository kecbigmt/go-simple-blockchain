package util

import (
  "net"
)

func MakeTCPAddr(ip string, port int) net.TCPAddr {
  return net.TCPAddr{
    IP: net.ParseIP(ip),
    Port: port,
  }
}
