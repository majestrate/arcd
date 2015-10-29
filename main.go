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

  var hubs []arc.Hub
  
  if len(cfg.Local.EtherBind) > 0 {
    eth := arc.CreateEthernetHub(cfg.Local.EtherBind, router)
    hubs = append(hubs, eth)
  }

  if cfg.IRC == nil {
    // no irc
  } else {
    irc := arc.NewIRCHub(router)
    go irc.Run()
    for _, remote := range cfg.IRC {
      go irc.Persist(remote)
    }
    hubs = append(hubs, irc)
  }
  
  hub := arc.CreateHub(cfg.Local.Bind, cfg.Local.Keys, router)
  for _, remote := range cfg.URC {
    go hub.Persist(remote)
  }
  hubs = append(hubs, hub)

  for idx, _ := range hubs {
    go hubs[idx].Run()
  }
  
  router.Run(hubs...)
}
