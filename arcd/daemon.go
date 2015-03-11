package arcd

import (
  "bytes"
  "crypto/ecdsa"
  "log"
  "net"
  "time"
)

type Daemon struct {
  Listener *net.TCPListener
  Broadacst chan *ARCMessage
  KadMessage chan *ARCMessage
  Filter DecayingBloomFilter
  Kad RoutingTable
  Us Peer
  PrivKey *ecdsa.PrivateKey
  PeerLoader PeerFileLoader
  hubs []*HubHandler
}

type HubHandler struct {
  conn net.Conn
  daemon *Daemon
  Broadacst chan *ARCMessage
  TheirHash []byte
  KadTries map[string][][]byte // key = target, val = keys tried
}

func (self *HubHandler) Init(daemon *Daemon, conn net.Conn) {
  self.daemon = daemon
  self.conn = conn
  self.Broadacst = make(chan *ARCMessage, 24)
  daemon.hubAdd(self)
}

func (self *Daemon) Init() {
  fname := "privkey.pem"
  var err error
  if FileExists(fname) {
    self.PrivKey, err = LoadECC_256(fname)
    if err != nil {
      log.Fatal(err)
    }
  } else {
    log.Println("generate signing key")
    self.PrivKey, err = GenerateECC_256()
    if err != nil {
      log.Fatal(err)
    }
    DumpECC_256(self.PrivKey, fname)
  }
 
  self.Kad.OurHash = ECC_256_KeyHash(self.PrivKey.PublicKey)
  log.Println("our hash is", FormatHash(self.Kad.OurHash))
  self.Kad.Init()
  self.Broadacst = make(chan *ARCMessage, 24)
  self.KadMessage = make(chan *ARCMessage, 24)
  self.Filter.Init() 
  self.hubs = make([]*HubHandler, 128)
}

func (self *Daemon) LoadPeers(fname string) {
  var err error
  err = self.PeerLoader.LoadFile(fname)
  if err != nil {
    log.Fatal("could not load peers ", err)
  }
  for idx := range(self.PeerLoader.Peers) {
    peer := self.PeerLoader.Peers[idx]
    self.AddPeer(peer.Net, peer.Addr, peer.PubKey)
  }
}

func (self *Daemon) AddPeer(net, addr, pubkey string) {
  log.Println("Add peer", net, addr, pubkey)
  go self.PersistHub(addr, pubkey)
}

func (self *Daemon) Bind(addr string) error {
  var err error
  var netaddr *net.TCPAddr
  netaddr, err = net.ResolveTCPAddr("tcp6", addr)
  if err != nil {
    return err
  }
  self.Listener, err = net.ListenTCP("tcp6", netaddr)
  if err == nil {
    log.Println("bound on", self.Listener.Addr())
  }
  self.Us.Addr = addr
  self.Us.Net = "ipv6"
  self.Us.PubKey = ECC_256_PubKey_ToString(self.PrivKey.PublicKey)
  log.Println("our public key is", self.Us.PubKey)
  return err
}

func (self *Daemon) PersistHub(addr, pubkey string) {
  if len(addr) == 0 {
    return
  }
  log.Println("persist hub at", addr)
  hash := ECC_256_KeyHash(ECC_256_UnPackPubKeyString(pubkey))
  for {
    time.Sleep(time.Second *1)
    conn, err := net.Dial("tcp6", addr)
    if err != nil {
      log.Println("failed to connect to hub", err)
      continue
    }
    log.Println("connect to hub", addr)
    var handler HubHandler
    handler.Init(self, conn)
    handler.TheirHash = hash
    handler.SendIdent()
    self.Kad.Insert(hash)
    go handler.WriteMessages()
    handler.ReadMessages()
    self.Kad.Remove(hash)
  }
}

func (self *Daemon) hubAdd(handler *HubHandler) {
  for idx := range(self.hubs) {
    if self.hubs[idx] == nil {
      self.hubs[idx] = handler
      return
    }
  }
  log.Fatal("too many hubs connected")
}

func (self *Daemon) hubRemove(handler *HubHandler) {
  for idx := range(self.hubs) {
    if handler == self.hubs[idx] {
      self.hubs[idx] = nil
    }
  }
}

func (self *HubHandler) SendIdent() {
  msg := NewArcIdentityMessage(self.daemon.Us, self.daemon.PrivKey)
  self.Broadacst <- msg
}

func (self *HubHandler) ReadMessages() {
  log.Println("reading...")
  for {
    msg := ReadARCMessage(self.conn)
    if msg == nil {
      log.Println("did not read arc message")
      self.conn.Close()
      self.daemon.hubRemove(self)
      return
    }
    buff := msg.Bytes()
    if self.daemon.Filter.Contains(buff) {
      log.Println("filter hit")
      continue
    }
    log.Println("Got Message of size", msg.MessageLength)
    
    if msg.MessageType == ARC_MESG_TYPE_DHT {
      peerHash := msg.DestHash
      peerHashStr := FormatHash(peerHash)
      if msg.GetPayloadString() == "NACK" {
        // backtrack
        val, ok := self.KadTries[peerHashStr]
        if ! ok {
          // we started this search
          val =  make([][]byte, MAX_KAD_TRIES)
          self.KadTries[peerHashStr] = val
        } 
        closest := self.daemon.Kad.GetClosestPeerExcludes(peerHash, val)
        var put bool
        for idx := range(val) {
          if val[idx] == nil {
            val[idx] = closest
            put = true
            break
          }
        }
        if put {
          msg.SetPayloadString("FIND")
          self.daemon.SendTo(closest, msg)
        } else {
          log.Println("giving up on kad search for", msg.DestHash)
          self.Broadacst <- msg
        }
      }
      
      if self.daemon.Kad.HashIsUs(peerHash) {
        self.daemon.KadMessage <- msg
      } else {
        closest := self.daemon.Kad.GetClosestPeerNotMe(peerHash)
        // we don't have any closest peers?
        if closest == nil {
          self.Broadacst <- NewArcKADMessage(peerHash, "NACK")
        } else {
          log.Println("relay kad message to", FormatHash(closest))
          self.daemon.SendTo(closest, msg)
        }
      }
    } else if msg.MessageType == ARC_MESG_TYPE_CHAT { 
      self.daemon.Broadacst <- msg
    } else  if msg.MessageType == ARC_MESG_TYPE_CTL {
      verified := msg.VerifyIdentity() 
      if verified {
        pubkey := msg.GetPubKey()
        hash := ECC_256_KeyHash(pubkey)
        if self.TheirHash == nil {
          self.TheirHash = hash
          self.daemon.Kad.Insert(hash)
          log.Println("hub identified as", FormatHash(hash))
        }
      }
    }
  }
}

func (self *Daemon) SendKad(target []byte, msg *ARCMessage) {
  closest := self.Kad.GetClosestPeerNotMe(target)
  if closest == nil {
    log.Println("we have no peers to send to for target", FormatHash(target))
  } else {
    log.Println("send kad message to", FormatHash(closest))
    self.SendTo(closest, msg)
  }
}

func (self *HubHandler) WriteMessages() {
  log.Println("writing...")
  for {
    msg := <- self.Broadacst 
    log.Println("hub write message")
    buff := msg.Bytes()
    _, err := self.conn.Write(buff)
    if err != nil {
      log.Println("Failed to write message", err)
      self.conn.Close()
      self.daemon.hubRemove(self)
      return
    }
  }
}

func (self *Daemon) SendTo(target []byte, msg *ARCMessage) {
  for idx := range(self.hubs) {
    if self.hubs[idx] != nil {
      hub := self.hubs[idx]
      if hub.TheirHash != nil {
        if bytes.Equal(hub.TheirHash, target) {
          hub.Broadacst <- msg
          return
        }
      }
    }
  }
}

func (self *Daemon) got_Broadcast(msg *ARCMessage, ircd *IRCD) {
  ircd.Broadcast <- string(msg.MessageData)
  buff := msg.Bytes()
  self.Filter.Add(buff)
  for idx := range(self.hubs) {
    if self.hubs[idx] != nil {
      log.Println("send to hub")
      self.hubs[idx].Broadacst <- msg
    }
  }
}

func (self *Daemon) got_KadMesssage(msg *ARCMessage) {
  log.Println("we got a kad message :D")
}

func (self *Daemon) Run(ircd *IRCD) {
  go self.Accept()
  var counter uint8
  log.Println("running hub")
  for {
    counter ++
    if counter == 0 {
      self.Filter.Decay()
    }
    var msg *ARCMessage
    select {
      case msg = <- self.Broadacst:
        self.got_Broadcast(msg, ircd)
      case msg = <- self.KadMessage:
        self.got_KadMesssage(msg)
    }
  }
}

func (self *Daemon) Accept() {
  for {
    conn, err := self.Listener.Accept()
    if err != nil {
      log.Fatal("error in Daemon::Accept()", err)
      return
    }
    log.Println("new incoming hub connection", conn.RemoteAddr())
    handler := new(HubHandler)
    handler.Init(self, conn)
    go handler.ReadMessages()
    go handler.WriteMessages()
  }
}