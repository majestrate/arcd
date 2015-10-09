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

  router := arc.NewBroadcastRouter(cfg.Local.Keys)

  var eth arc.Hub
  
  if len(cfg.Local.EtherBind) > 0 {
    eth = arc.CreateEthernetHub(cfg.Local.EtherBind, router)
  }

  irc := arc.NewIRCHub(router)
  go irc.Run()
  for _, remote := range cfg.IRC {
    go irc.Persist(remote)
  }
  
  hub := arc.CreateHub(cfg.Local.Bind, cfg.Local.Keys, router)
  for _, remote := range cfg.URC {
    go hub.Persist(remote)
  }
  go hub.Run()
  if eth == nil {
    router.Run(hub, irc)
  } else {
    go eth.Run()
    router.Run(hub, eth, irc)
  }
}