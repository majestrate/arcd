package arcd

import (
  "crypto/rsa"
  "log"
  "net"
  "time"
)

type Daemon struct {
  Listener *net.TCPListener
  Broadacst chan *ARCMessage
  Filter DecayingBloomFilter
  Kad RoutingTable
  PrivKey *rsa.PrivateKey
  PeerLoader PeerFileLoader
  hubs []*HubHandler
}

type HubHandler struct {
  conn net.Conn
  daemon *Daemon
  Broadacst chan *ARCMessage
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
    self.PrivKey, err = LoadRSA4K(fname)
    if err != nil {
      log.Fatal(err)
    }
  } else {
    log.Println("generate signing key")
    self.PrivKey, err = GenerateRSA4K()
    if err != nil {
      log.Fatal(err)
    }
    DumpRSA4K(self.PrivKey, fname)
  }
  log.Println("our public key is", FormatHash(self.PrivKey.PublicKey.N.Bytes()))
  self.Kad.OurHash = RSA4K_KeyHash(&self.PrivKey.PublicKey)
  self.Kad.Init()
  self.Broadacst = make(chan *ARCMessage, 24)
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
  go self.PersistHub(addr)
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
  return err
}

func (self *Daemon) PersistHub(addr string) {
  if len(addr) == 0 {
    return
  }
  log.Println("persist hub at", addr)
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
    go handler.WriteMessages()
    handler.ReadMessages()
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

func (self *Daemon) AddHub(addr string) {
  go self.PersistHub(addr)
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
    self.daemon.Filter.Add(buff)
    log.Println("Got Message of size", msg.MessageLength)
    self.daemon.Broadacst <- msg
  }
}

func (self *HubHandler) WriteMessages() {
  log.Println("writing...")
  for {
    msg := <- self.Broadacst 
    log.Println("hub write message")
    buff := msg.Bytes()
    if self.daemon.Filter.Contains(buff) {
      continue
    }
    self.daemon.Filter.Add(buff)
    _, err := self.conn.Write(buff)
    if err != nil {
      log.Println("Failed to write message", err)
      self.conn.Close()
      self.daemon.hubRemove(self)
      return
    }
  }
}

func (self *Daemon) SendTo(target []byte, line string) {
  msg := NewArcIRCLine(line)
  msg.Sign(self.PrivKey)
  copybytes(msg.DestHash, target, 0, 0, ARC_HASH_LEN)
  self.Broadacst <- msg
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
    msg := <- self.Broadacst
    log.Println("got message")
    if msg != nil {
      ircd.Broadcast <- string(msg.MessageData)
      for idx := range(self.hubs) {
        if self.hubs[idx] != nil {
          log.Println("send to hub")
          self.hubs[idx].Broadacst <- msg
        }
      }
    }
  }
}

func (self *Daemon) Accept() {
  for {
    conn, err := self.Listener.Accept()
    if err != nil {
      log.Fatal("error in Daemon::Accept()", err)
    }
    log.Println("new incoming hub connection", conn.RemoteAddr())
    var handler HubHandler
    handler.Init(self, conn)
    go handler.ReadMessages()
    go handler.WriteMessages()
  }
}