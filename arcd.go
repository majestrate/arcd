package main
import (
  "flag"
  "github.com/majestrate/arcd/arcd"
)

func main() {
  arcd.TestCrypto()
  peers := flag.String("peers", "peers.txt", "peers file")
  bind := flag.String("bind", "[::]:0", "bind hub to address")
  ircd_bind := flag.String("ircd", "[::]:0", "bind ircd to address")
  flag.Parse()
  var daemon arcd.Daemon
  var ircd arcd.IRCD
  ircd.BindAddr = *ircd_bind
  daemon.Init()
  ircd.Init(&daemon)
  go daemon.LoadPeers(*peers)
  daemon.Bind(*bind)
  go daemon.Run(&ircd)
  ircd.Run()
}