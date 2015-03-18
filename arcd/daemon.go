package arcd

import (
  "bytes"
  "crypto/ecdsa"
  "errors"
  "log"
  "net"
  "strconv"
  "strings"
  "time"
)

type Daemon struct {
  Listener *net.TCPListener
  Broadacst chan *ARCMessage
  KadMessage chan *ARCMessage
  Kad RoutingTable
  torproc *TorProc
  Us Peer
  PrivKey *ecdsa.PrivateKey
  PeerLoader PeerFileLoader
  hubs []*HubHandler
  numhubs uint
  KnownPeers []Peer
}

type HubHandler struct {
  conn net.Conn
  daemon *Daemon
  Broadacst chan *ARCMessage
  TheirHash []byte
  them Peer
  KadTries map[string][][]byte // key = target, val = keys tried
  Filter DecayingBloomFilter
}

func (self *HubHandler) Init(daemon *Daemon, conn net.Conn) {
  self.daemon = daemon
  self.conn = conn
  self.Broadacst = make(chan *ARCMessage, 24)
  daemon.hubAdd(self)
  self.Filter.Init() 
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
  self.hubs = make([]*HubHandler, 128)
  self.KnownPeers = make([]Peer, 128)

}

func (self *Daemon) LoadPeers(fname string) {
  var err error
  err = self.PeerLoader.LoadFile(fname)
  if err != nil {
    log.Fatal("could not load peers ", err)
  }
  for idx := range(self.PeerLoader.Peers) {
    peer := self.PeerLoader.Peers[idx]
    self.AddPeer(peer)
  }
}

func (self *Daemon) HasPeer(peer Peer) bool {
  for idx := range(self.KnownPeers) {
    if self.KnownPeers[idx].PubKey == peer.PubKey {
      return true
    }
  }
  return false
}

func (self *Daemon) AddPeer(peer Peer) {
  if peer.PubKey == self.Us.PubKey {
    log.Println("not adding self")
    return
  }
  if self.HasPeer(peer) {
    log.Println("already have peer", peer.Addr)
    return 
  }
  self.AddPeerStr(peer.Net, peer.Addr, peer.PubKey)
  for idx := range(self.KnownPeers) {
    if self.KnownPeers[idx].PubKey == "" {
      self.KnownPeers[idx] = peer
      return
    }
  }
}

func (self *Daemon) AddPeerStr(net, addr, pubkey string) {
  if self.numhubs < uint(len(self.hubs)) {
    log.Println("Add peer", net, addr, pubkey)
    go self.PersistHub(net, addr, pubkey)
  } else {
    log.Println("Not adding hub, too many connections")
  }
}

func (self *Daemon) Bind(addr string, socksport int) error {
  var err error
  var netaddr *net.TCPAddr
  netaddr, err = net.ResolveTCPAddr("tcp6", addr)
  if err != nil {
    return err
  }
  self.Listener, err = net.ListenTCP("tcp6", netaddr)
  
  if err != nil {
    return err
  }
  ouraddr := self.Listener.Addr()
  log.Println("bound on", ouraddr)
  
  
  // spawn tor
  self.torproc = SpawnTor(socksport)
  self.torproc.Start()
  time.Sleep(time.Second)
  onion := self.torproc.GetOnion()
  // load / run daeoms
  
  self.Us.Net = "tor"
  self.Us.Addr = onion+":11001"
  self.Us.PubKey = ECC_256_PubKey_ToString(self.PrivKey.PublicKey)
  log.Println("our public key is", self.Us.PubKey)
  log.Println("our onion is", onion)
  return err
}

func (self *Daemon) connectForNet(network, addr string) (net.Conn, error) {

  if network == "tor" {
    // connect to socks proxy
    conn, err := net.Dial("tcp", self.torproc.SocksAddr())
    if err != nil {
      return conn, err
    }
    var buff bytes.Buffer
    // socks request
    host := strings.Split(addr, ":")[0]
    port, err := strconv.Atoi(strings.Split(addr, ":")[1])
    if err != nil {
      conn.Close()
      return conn, err
    }
    
    buff.Write([]byte{ 4, 1, byte( (port >> 8) ), byte( (port & 0xff) ), 0, 0, 0, 1, 54, 0})
    buff.WriteString(host)
    buff.Write([]byte{0})
    conn.Write(buff.Bytes())
    recvbuff := make([]byte, 8)
    conn.Read(recvbuff)
    if recvbuff[1] != 0x5a {
      conn.Close()
      return conn, errors.New("failed to connect via socks proxy")
    }
    return conn, nil
  }
  return net.Dial("tcp6", "")
}

func (self *Daemon) PersistHub(network, addr, pubkey string) {
  if len(addr) == 0 {
    return
  }
  log.Println("persist hub at", addr)
  hash := ECC_256_KeyHash(ECC_256_UnPackPubKeyString(pubkey))
  for {
    time.Sleep(time.Second *1)
    conn, err := self.connectForNet(network, addr) 
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
  if self.numhubs < uint(len(self.hubs)) {
    for idx := range(self.hubs) {
      if self.hubs[idx] == nil {
        self.hubs[idx] = handler
        self.numhubs ++
        return
      }
    }
  }
  log.Println("too many connections")
  handler.conn.Close()
}

func (self *Daemon) hubRemove(handler *HubHandler) {
  for idx := range(self.hubs) {
    if handler == self.hubs[idx] {
      self.hubs[idx] = nil
      self.numhubs --
    }
  }
}

func (self *Daemon) GetPeers(maxnum int) []Peer {
  peers := make([]Peer, maxnum)
  for idx := range(self.hubs) {
    handler := self.hubs[idx] 
    if handler != nil {
      if handler.them.PubKey != "" {
        peers[maxnum-1] = handler.them
        maxnum --;
      }
    }
  }
  peers[0] = self.Us
  return peers
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
    if self.Filter.Contains(buff) {
      log.Println("filter hit")
      continue
    }
    self.Filter.Add(buff)
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
      if self.TheirHash == nil {
        verified := msg.VerifyIdentity() 
        if verified {
          self.them = msg.GetPeer()
          pubkey := ECC_256_UnPackPubKeyString(self.them.PubKey)
          hash := ECC_256_KeyHash(pubkey)
          log.Println("hub identified as", FormatHash(hash))
          self.TheirHash = hash
          self.daemon.Kad.Insert(hash)
        } else {
          log.Println("hub failed to identify")
          self.conn.Close()
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
    self.Filter.Add(buff)
    wr, err := self.conn.Write(buff)
    if err != nil {
      log.Println("Failed to write message", err)
      self.conn.Close()
      self.daemon.hubRemove(self)
      return
    }
    log.Println("wrote", wr)
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
  log.Println("running hub")
  for {
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