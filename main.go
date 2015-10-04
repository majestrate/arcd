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
    router = router.RegisterHub(eth)
    go eth.Run()
  }
  
  hub := arc.CreateHub(cfg.Local.Bind, cfg.Local.Keys, router)
  for _, remote := range cfg.Remote {
    hub.Persist(remote.Addr, remote.Port, remote.ProxyType, remote.ProxyAddr, remote.ProxyPort)
  }
  router = router.RegisterHub(hub)
  go hub.Run()
  router.Run()
}
