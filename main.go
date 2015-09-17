package main

import (
  "github.com/majestrate/arcd/arc"
  "os"
)

func main() {
  fname := "config.json"
  if len(os.Args) > 1 {
    fname = os.Args[1]
  }
  
  cfg := arc.LoadConfig(fname)

  hub := arc.CreateHub(cfg.Local.Bind, cfg.Local.Keys)
  for _, remote := range cfg.Remote {
    hub.PersistURC(remote.Addr, remote.Port, remote.ProxyType, remote.ProxyAddr, remote.ProxyPort)
  }
  hub.Run()
}
