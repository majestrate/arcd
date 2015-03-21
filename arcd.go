package main
import (
  "flag"
  "github.com/majestrate/arcd/arcd"
  "time"
)

func main() {
  
  // initial crypto test
  arcd.TestCrypto()
  
  // command line flags
  peers := flag.String("peers", "peers.txt", "peers file")
  bind := flag.String("bind", "127.0.0.1:11001", "bind hub to address")
  ircd_bind := flag.String("ircd", "[::1]:6667", "bind ircd to address")
  ping := flag.String("kad", "", "kad ping a peer")
  socksport := flag.Int("torsocks", 11005, "tor socks port")
  flag.Parse()
  
  
  // ircd
  var daemon arcd.Daemon
  var ircd arcd.IRCD
  ircd.BindAddr = *ircd_bind
  
  // initialize daemon
  daemon.Init()
  daemon.Bind(*bind, *socksport)
  
  // bind irc
  ircd.Init(&daemon)
  
  go daemon.LoadPeers(*peers)
  go daemon.Run(&ircd)
  go ircd.Run()
  
  
  // do pings as needed
  for {
    time.Sleep(2 * time.Second)
    if *ping != "" {
      peer := arcd.UnFormatHash(*ping)
      dmsg := arcd.NewDHTMessage("FIND")
      msg := arcd.NewArcKADMessage(peer, dmsg)
      daemon.SendKad(peer, msg)
    }
  }
  
}