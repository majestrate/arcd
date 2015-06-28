//
// router_testnet.go
// arc router implementation for test network
//

package arcd

import (
  "log"
  "net"
)

type TestnetRouterMain struct {
  ourInfo PeerInfo
  // the key of our node
  ourKey NodeHash
  // dht
  r5n R5NKademlia 
  // our node database
  nodes NodeDatabase
  // our udp link level transport
  udp testnetLink
}

// configure our router
func (self TestnetRouterMain) Configure(args map[string]string) RouterMain {
  self.nodes = nodeTextFile("nodes.txt")
  // make chan for link
  self.udp.ibMsgChan = make(chan Message)
  // bind the udp link
  bindaddr := args["testnet_bindaddr"]
  uaddr, err := net.ResolveUDPAddr("udp6", bindaddr)
  if err != nil {
    log.Fatalf("failed to resolve %s: %s", bindaddr, err)
  }
  log.Println("bind to udp address", uaddr)
  self.udp.udp_socket, err = net.ListenUDP("udp6", uaddr)
  if err != nil {
    log.Fatalf("failed to bind to %s: %s", bindaddr, err)
  }

  log.Println("bound to", self.udp.udp_socket.LocalAddr())
  
  self.ourInfo = testnetPeerInfo(bindaddr)
  
  // make a random key
  randkey := randBytes(len(self.ourKey))
  copy(self.ourKey[:], randkey)

  self.r5n = newR5NKad(self.ourInfo, self.ourKey, 32)
  // start link level mainloop
  go self.udp.Mainloop()
  
  return self
}

// runit :^3
func (self TestnetRouterMain) Run() {
  chnl := self.udp.MessageChan()
  for {
    select {
    case msg := <- chnl:
      self.got_message(msg)
    }
  }
}

func (self *TestnetRouterMain) got_message(msg Message) {
  
}
