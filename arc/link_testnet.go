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

func (self testnetPeerInfo) NetAddr() string {
  return strings.Split(string(self), " ")[0]
}

func (self testnetPeerInfo) NodeHash() NodeHash {
  b, err := hex.DecodeString(strings.Split(string(self), " ")[1])
  if err != nil {
    log.Fatalf("cannot decode node hash from peerinfo: [%s]", self)
  }
  var nh NodeHash
  copy(nh[:], b)
  return nh
}

func (self testnetPeerInfo) Net() string {
  return "udp6"
}



type testnetLinkMessage []byte

func (self testnetLinkMessage) AsDHTMessage() DHTMessage {
  return DHTMessage(self[2:])
}

func (self testnetLinkMessage) AsNodeMessage() NodeMessage {
  return NodeMessage(self[2:])
}


func (self testnetLinkMessage) Bytes() []byte {
  return self
}

func (self testnetLinkMessage) Version() byte {
  return self[0]
}

func (self testnetLinkMessage) Type() byte {
  return self[1]
}

// no sigs on testnet
func (self testnetLinkMessage) GetSig() Signature {
  return nil
}

func (self testnetLinkMessage) BodyReader() io.Reader {
  return bytes.NewBuffer(self[2:])
}

// always valid on testnet
func (self testnetLinkMessage) Valid() bool {
  return true
}

type testnetLink struct {

  udp_socket *net.UDPConn
  ibMsgChan chan Message
}

func (self testnetLink) RecvHeader() (LinkMessageHeader, error) {
  // fixed buffer for recv
  recvBuff := make([]byte, 2048)
  n, uaddr, err := self.udp_socket.ReadFromUDP(recvBuff)
  if err != nil {
    return nil, err
  }
  log.Printf("got %d from %q", n, uaddr)
  return testnetLinkMessage(recvBuff[:n]), nil
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
    self.ibMsgChan <- testnetLinkMessage(linkmsg.Bytes())
  }
}

func (self testnetLink) MessageChan() chan Message {
  return self.ibMsgChan
}
