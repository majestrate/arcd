package arcd

import (
  "bytes"
  //"crypto/rsa"
  "crypto/ecdsa"
  "io"
  "log"
)

const ARC_HASH_LEN uint = 32
const ARC_SIG_LEN uint = 64
const ARC_HEADER_LEN uint16 = uint16((ARC_HASH_LEN * 2) + 2 + 2 + 8 + 1)
const (
  // control message
  ARC_MESG_TYPE_CTL = iota +1
  // plain chat message
  ARC_MESG_TYPE_CHAT
  // encrypted chat message
  ARC_MESG_TYPE_CRYPT_CHAT
  // dht routed message
  ARC_MESG_TYPE_DHT
)


const ARC_PROTOCOL_NUM uint8 = 0x03

type ARCMessage struct {
  // begin header
  ProtocolByte uint8 
  // for dht
  SourceHash []byte
  DestHash []byte
  // general
  MessageType uint16 
  MessageLength uint16
  MessageTime uint64
  // end header
  MessageData []byte
}

func ReadARCMessage(reader io.Reader) *ARCMessage {
  var hdr []byte
  hdr = make([]byte, ARC_HEADER_LEN)
  _, err := reader.Read(hdr)
  if err != nil {
    return nil
  }
  mesg := new(ARCMessage)
  
  // protocol zero
  if hdr[0] == ARC_PROTOCOL_NUM {
    mesg.Init(0)
    mesg.ProtocolByte = hdr[0]
    copybytes(mesg.SourceHash, hdr, 0, 1, ARC_HASH_LEN)
    copybytes(mesg.DestHash, hdr, 0, 1 + ARC_HASH_LEN, ARC_HASH_LEN)
    mesg.MessageType = getshort(hdr, 1 + (ARC_HASH_LEN * 2))
    mesg.MessageLength = getshort(hdr, 1 + 2 + (ARC_HASH_LEN * 2))
    mesg.MessageTime = getlong(hdr, 1 + 2 + 2 + (ARC_HASH_LEN * 2))
    var mlen uint16
    mlen = mesg.MessageLength
    
    mesg.MessageData = make([]byte, mlen)
    log.Println("len=", mlen)
    log.Println("len=", len(mesg.MessageData))
    _, err = reader.Read(mesg.MessageData)
    if err != nil {
      log.Println("failed to read arc message payload of size", mesg.MessageLength)
      return nil
    }
    return mesg
  } else {
    log.Println("invalid protocol number", hdr[0])
    return nil
  }
}

func (self *ARCMessage) Init(mtype uint16) {
  self.ProtocolByte = ARC_PROTOCOL_NUM
  self.SourceHash = make([]byte, ARC_HASH_LEN)
  self.DestHash = make([]byte, ARC_HASH_LEN)
  self.MessageType = mtype
}

func (self *ARCMessage) StampTime() {
  self.MessageTime = TimeNow()
}

func (self *ARCMessage) SetPayload(data []byte) {
  self.MessageLength = uint16(len(data))
  log.Println("set payload of size", self.MessageLength)
  self.MessageData = make([]byte,  self.MessageLength)
  copybytes(self.MessageData, data, 0, 0, uint(len(data)))
  log.Println("pay is", len(self.MessageData))
  
}

func (self *ARCMessage) Bytes() []byte {
  bufflen := uint(ARC_HEADER_LEN + self.MessageLength)
  buff :=  make([]byte, bufflen)
  // make header
  buff[0] = self.ProtocolByte
  copybytes(buff, self.SourceHash, 1, 0, ARC_HASH_LEN)
  copybytes(buff, self.DestHash, ARC_HASH_LEN + 1, 0, ARC_HASH_LEN)
  putshort(self.MessageType, buff, (ARC_HASH_LEN * 2) + 1)
  putshort(self.MessageLength, buff, (ARC_HASH_LEN * 2) + 2 + 1)
  putlong(self.MessageTime, buff, (ARC_HASH_LEN * 2)+ 2 + 2 + 1)
  copybytes(buff, self.MessageData, uint(ARC_HEADER_LEN), 0, uint(self.MessageLength))
  return buff
}

func (self *ARCMessage) Sign(privkey *ecdsa.PrivateKey) {

  self.MessageLength += uint16(ARC_SIG_LEN)
  nosig := make([]byte, ARC_SIG_LEN)
  
  var oldmsg bytes.Buffer
  oldmsg.Write(self.MessageData)
  oldmsg.Write(nosig)
  self.MessageData = oldmsg.Bytes()
  
  buff := self.Bytes()
  //log.Println(buff)
  sig, err := SignECC_256(buff, privkey)
  if err != nil {
    log.Fatal(err)
  }
  //log.Println(sig)
  copybytes(self.MessageData, sig, uint(uint(self.MessageLength) - ARC_SIG_LEN), 0, uint(ARC_SIG_LEN))
  
}

func (self *ARCMessage) Verify(pubkey *ecdsa.PublicKey) bool {
  buff := self.Bytes()
  idx := len(buff) - int(ARC_SIG_LEN)
  sig := make([]byte, ARC_SIG_LEN)
  copybytes(sig, buff, 0, uint(idx), ARC_SIG_LEN)
  
  // zero out sig
  for c := 0 ; c < int(ARC_SIG_LEN) ; c++ {
    buff[idx+c] = 0
  }
  return VerifyECC_256(buff, sig, pubkey)
}

func (self *ARCMessage) GetPeer() Peer {
  idx := len(self.MessageData)-int(ARC_SIG_LEN)
  log.Println(idx)
  data := self.MessageData[:idx]
  var peer Peer
  if ! peer.Parse(string(data)) {
    log.Println("invalid peer data", string(data))
  }
  return peer
}

func (self *ARCMessage) VerifyIdentity() bool {
  pubkey := ECC_256_UnPackPubKeyString(self.GetPeer().PubKey)
  return self.Verify(&pubkey)
}

func (self *ARCMessage) SetPayloadString(data string) {
  var buff bytes.Buffer
  buff.WriteString(data)
  self.SetPayload(buff.Bytes())
}

func (self *ARCMessage) GetPayloadString() string {
  return string(self.MessageData)
}

func NewArcIRCLine(line string) *ARCMessage {
  log.Println("new irc line", line)
  buff := bytes.NewBufferString(line)
  msg := new(ARCMessage)
  msg.Init(ARC_MESG_TYPE_CHAT)
  msg.SetPayload(buff.Bytes())
  msg.StampTime()
  return msg
}

func NewArcIdentityMessage(us Peer, privkey *ecdsa.PrivateKey) *ARCMessage {
  msg := new(ARCMessage)
  msg.Init(ARC_MESG_TYPE_CTL)
  msg.SetPayload(us.Bytes())
  msg.StampTime()
  msg.Sign(privkey)
  return msg
}

func NewArcPeersMessage(peers []Peer) *ARCMessage {
  msg := new(ARCMessage)
  msg.Init(ARC_MESG_TYPE_CTL)
  var buff bytes.Buffer
  // max 30 peers
  counter := 30
  for idx := range(peers) {
    if counter == 0 {
      break
    }
    peer := peers[idx]
    if peer.PubKey != ""  {
      buff.Write(peer.Bytes())
      buff.WriteString("\n")
    }
    counter --
  }
  msg.SetPayload(buff.Bytes())
  msg.StampTime()
  return msg
}


func NewArcKADMessage(target []byte, data string) *ARCMessage {
  msg := new(ARCMessage)
  msg.Init(ARC_MESG_TYPE_DHT)
  l := len(target)
  copybytes(msg.DestHash, target, 0, 0, uint(l))
  var buff bytes.Buffer
  buff.WriteString(data)
  msg.SetPayload(buff.Bytes())
  msg.StampTime()
  return msg
}