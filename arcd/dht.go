package arcd

import (
  "bytes"
  bencode "github.com/majestrate/bencode-go"
  "log"
)

const DHT_DEFAULT_HOPS uint = 0

type DHTMessage struct {
  Method string
  Hops uint
  Payload []byte
}

func NewDHTMessage(method string) *DHTMessage {
  msg := new(DHTMessage)
  msg.Method = method
  msg.Hops = DHT_DEFAULT_HOPS
  return msg
}

func (self *DHTMessage) Bytes() []byte {
  var buff bytes.Buffer
  err := bencode.Marshal(&buff, self)
  if err != nil {
    log.Println("failed to marshal dht message", err)
    return nil
  }
  return buff.Bytes()
}

func ParseDHTMessage(raw []byte) *DHTMessage {
  
  reader := bytes.NewReader(raw)
  msg := new(DHTMessage)
  err := bencode.Unmarshal(reader, msg)
  if err != nil {
    log.Println("failed to parse dht message", err)
    return nil
  }
  return msg
}