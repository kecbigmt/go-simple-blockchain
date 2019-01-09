package util

import(
  "testing"
)

func TestGetMyIP(t *testing.T) {
  ip, err := GetMyIP()
  if err != nil {
    t.Fatalf("error occured: %s", err)
  }
  if ip == "" {
    t.Fatal("ip is empty")
  }
}
