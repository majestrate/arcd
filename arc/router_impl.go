//
// router_testnet.go
// arc router implementation
//

package arcd

import (
  "encoding/hex"
  "fmt"
  "log"
  "net"
  "strings"
)

type ArcRouterMain struct {
  ourInfo PeerInfo
  // dht
  kad Kademlia 
  // our node database
  nodes nodeTextFile
  // our udp link level transport
  udp testnetLink
  // datastore
  store ChunkStore
  // insert data chunk channel
  insertChan chan DHTPutPayload
  // message repliers
  dhtget_reply DHTMessageProcessor
  dhtput_reply DHTMessageProcessor
}

// configure our router
func (self ArcRouterMain) Configure(args map[string]string) RouterMain {
  // insert chunk channel
  self.insertChan = make(chan DHTPutPayload)
  // 24K pieces per router
  self.store = createRamChunkStore(1024 * 24)
  // XXX: testnet
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
  localaddr := self.udp.udp_socket.LocalAddr()
  log.Println("bound to", localaddr)

  // make a random key
  // XXX: testnet
  randkey := randBytes(len(new(NodeHash)))
  self.ourInfo = testnetPeerInfo(fmt.Sprintf("%s %s", localaddr, hex.EncodeToString(randkey)))
  self.nodes.AddInfo(self.ourInfo)

  log.Println("our info is", self.ourInfo)

  // set udp link peer info
  self.udp.ourInfo = self.ourInfo
  // make dht engine
  self.kad = newKad(self.ourInfo, 32)
  // set up dht message repliers
  self.dhtget_reply = createGetReplier(self.kad, self.store)
  self.dhtput_reply = createPutReplier(self.kad, self.store)
  
  // populate nodedb for udp link
  // XXX: testnet
  nodes := self.nodes.AllNodes()
  for _, info := range nodes {
    self.udp.nodes[info.NetAddr()] = info 
  }
  
  // start link level mainloop
  go self.udp.Mainloop()

  // return 
  return self
}

// runit :^3
func (self ArcRouterMain) Run() {
  chnl := self.udp.MessageChan()
  for {
    select {
    case msg := <- chnl:
      switch (msg.Type()) {
      case MSG_TYPE_DHT:
        self.got_dht_message(msg.AsDHTMessage(), msg.Source())
        break
      default:
        log.Println("got invalid link message")
      }
    case insert := <- self.insertChan:
      // we have an insert we gotta do
      // get the hash
      log.Println("insert a piece")
      h := insert.CalcHash()
      // find the closest hop
      info, success := self.kad.GetNextHop(h)
      if success {
        msg := bencDHTMessage{
          data: insert[:],
          tid: self.createTID(),
          hops: 0,
          reply: false,
          method: "PUT",
          
        }
        self.send_dht_message_to(info, msg)  
      } else {
        log.Printf("failed to do insert for %s, no closest peers found D:", h) 
      }
    }
  }
}

// return random tid that we don't have
func (self ArcRouterMain) createTID() int64 {
  // TODO: implement 
  return 4; // :^3
}

// send a dht message to a peer that we know
func (self ArcRouterMain) send_dht_message_to(info PeerInfo, msg DHTMessage) {
  // create link messages
  msgs := self.udp.GetMessageFactory().CreateMessagesForDHT(msg)
  if msgs == nil {
    log.Println("failed to generate link level messages?")
  } else {
    for _, lmsg := range msgs {
      self.udp.SendMessage(lmsg, info)
    }
  }
}

// obtain a chukn of data via its hash
func (self ArcRouterMain) GetDataViaHash(h CryptoHash) []byte {
  // TODO: implement
  return nil
}

// insert some data into the dht
// return the root block's hash
func (self ArcRouterMain) InsertData(payload []byte) CryptoHash {
  // create chunks
  chunks := createInsertChunks(payload)
  // insert them
  for _, chunk := range chunks {
    self.insertChan <- chunk
  }
  // return the root hash
  return chunks[0].CalcHash()
}

// called when we recv a dht message
func (self ArcRouterMain) got_dht_message(msg DHTMessage, from PeerInfo) {
  log.Printf("got dht message from %s", from.NodeHash())
  method := strings.ToUpper(msg.Method())
  if method == "GET" {
    
  } else if method == "PUT"{
    
  }
}

