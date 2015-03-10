package main
import (
  "flag"
  "github.com/majestrate/arcd/arcd"
  "time"
)

func main() {
  arcd.TestCrypto()
  peers := flag.String("peers", "peers.txt", "peers file")
  bind := flag.String("bind", "[::]:0", "bind hub to address")
  ircd_bind := flag.String("ircd", "[::]:0", "bind ircd to address")
  ping := flag.String("kad", "", "kad ping a peer")
  flag.Parse()
  var daemon arcd.Daemon
  var ircd arcd.IRCD
  ircd.BindAddr = *ircd_bind
  daemon.Init()
  ircd.Init(&daemon)
  go daemon.LoadPeers(*peers)
  daemon.Bind(*bind)
  go daemon.Run(&ircd)
  go ircd.Run()
  time.Sleep(3 * time.Second)
  for {
    time.Sleep(time.Second)
    if *ping != "" {
      peer := arcd.UnFormatHash(*ping)
      msg := arcd.NewArcKADMessage(peer)
      daemon.SendKad(peer, msg)
    }
  }
}