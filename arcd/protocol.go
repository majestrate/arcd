package arcd

import (
  "bytes"
  "crypto/rsa"
  "io"
  "log"
)

const ARC_HASH_LEN uint = 32
const ARC_HEADER_LEN uint16 = uint16(ARC_HASH_LEN * 2 + 2 + 2 + 8 + 1)
const (
  // control message
  ARC_MESG_TYPE_CTL = iota
  // plain chat message
  ARC_MESG_TYPE_CHAT
  // encrypted chat message
  ARC_MESG_TYPE_CRYPT_CHAT
  // dht routed message
  ARCH_MESG_TYPE_DHT
)


type ARCMessage struct {
  // begin header
  ProtocolByte uint8 // right now it's 0x00
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
  mesg.Init(0)
  // protocol zero
  if hdr[0] == 0 {
    mesg.ProtocolByte = hdr[0]
    copybytes(mesg.SourceHash, hdr, 0, 1, ARC_HASH_LEN)
    copybytes(mesg.DestHash, hdr, 0, 1 + ARC_HASH_LEN, ARC_HASH_LEN)
    mesg.MessageType = getshort(hdr, 1 + (ARC_HASH_LEN * 2))
    mesg.MessageLength = getshort(hdr, 1 + 2 + (ARC_HASH_LEN * 2))
    mesg.MessageTime = getlong(hdr, 1 + 2 + 2 + (ARC_HASH_LEN * 2))
    mesg.MessageData = make([]byte, mesg.MessageLength)
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
  copybytes(self.MessageData, data, 0, 0, uint(self.MessageLength))
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

func (self *ARCMessage) Sign(privkey *rsa.PrivateKey) {
  self.MessageLength += 32
  buff := self.Bytes()
  sig, err := SignRSA4K(buff, privkey)
  if err != nil {
    log.Fatal(err)
  }
  log.Println("sig len=", len(sig))
}

func NewArcPing() *ARCMessage {
  buff := bytes.NewBufferString("PING")
  msg := new(ARCMessage)
  msg.Init(ARC_MESG_TYPE_CTL)
  msg.SetPayload(buff.Bytes())
  msg.StampTime()
  return msg
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