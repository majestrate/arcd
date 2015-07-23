//
// link protocol for testnet
//

package arcd

import (
  "bytes"
  "encoding/hex"
  "time"
  "io"
  "log"
  "net"
  "strings"
)

type testnetPeerInfo string

func (self testnetPeerInfo) String() string {
  return string(self)
}

func (self testnetPeerInfo) NetAddr() string {
  return strings.Split(string(self), " ")[0]
}

func (self testnetPeerInfo) NodeHash() CryptoHash {
  b, err := hex.DecodeString(strings.Split(string(self), " ")[1])
  if err != nil {
    log.Fatalf("cannot decode node hash from peerinfo: [%s]", self)
  }
  var nh CryptoHash
  copy(nh[:], b)
  return nh
}

func (self testnetPeerInfo) Net() string {
  return "udp6"
}



type testnetLinkMessage struct {
  from PeerInfo
  data []byte
}

func (self testnetLinkMessage) AsDHTMessage() DHTMessage {
  msg := bencDHTMessage{self.data[2:], "", false, 0, 0, nil}
  msg.Parse()
  return msg
}

func (self testnetLinkMessage) AsNodeMessage() NodeMessage {
  return NodeMessage(self.data[2:])
}

func (self testnetLinkMessage) Source() PeerInfo {
  return self.from
}

func (self testnetLinkMessage) Bytes() []byte {
  return self.data
}

func (self testnetLinkMessage) Version() byte {
  return self.data[0]
}

func (self testnetLinkMessage) Type() byte {
  return self.data[1]
}

// no sigs on testnet
func (self testnetLinkMessage) GetSig() Signature {
  return nil
}

func (self testnetLinkMessage) BodyReader() io.Reader {
  return bytes.NewBuffer(self.data[2:])
}

// always valid on testnet
func (self testnetLinkMessage) Valid() bool {
  return true
}

type testnetLink struct {
  ourInfo PeerInfo
  nodes map[string]PeerInfo
  udp_socket *net.UDPConn
  ibMsgChan chan Message
}

func (self testnetLink) RecvHeader() (LinkMessageHeader, error) {
  // fixed buffer for recv
  recvBuff := make([]byte, 4096)
  n, uaddr, err := self.udp_socket.ReadFromUDP(recvBuff)
  if err != nil {
    return nil, err
  }
  log.Printf("got %d from %q", n, uaddr)
  msg :=  testnetLinkMessage{
    data: recvBuff[:n],
    from: self.nodes[uaddr.String()],
  }
  return msg, nil
}

// send a message in 1 udp packet
func (self testnetLink) SendMessage(msg Message, info PeerInfo) error {
  // resolve the address
  addr := info.NetAddr()
  uaddr, err := net.ResolveUDPAddr(info.Net(), addr)
  if err != nil {
    return err
  }
  // just send it
  n, err := self.udp_socket.WriteToUDP(msg.Bytes(), uaddr)
  log.Printf("sent %d to %s", n, addr)
  return err
}

func (self testnetLink) Mainloop() {
  log.Println("begin testnet udp link mainloop")
  for {
    linkmsg, err := self.RecvHeader()
    if err != nil {
      log.Println("testnetLink::RecvHeader()", err)
      time.Sleep(time.Second)
    }
    self.ibMsgChan <- testnetLinkMessage{from: self.ourInfo, data: linkmsg.Bytes()}
  }
}

func (self testnetLink) MessageChan() chan Message {
  return self.ibMsgChan
}

func (self testnetLink) CreateMessagesForDHT(msg DHTMessage) []Message {
  // header
  buff := make([]byte, 2)
  buff[0] = 0 // version
  buff[1] = byte(MSG_TYPE_DHT) // type
  // body
  d := msg.Bytes()
  buff = append(buff, d...)

  var msgs []Message
  // our single link message
  m := testnetLinkMessage{data: buff}
  return append(msgs, m)
}

func (self testnetLink) GetMessageFactory() LinkMessageFactory {
  return self
}
