//
// message.go -- chat message protocol
//
package arc

import (
  "encoding/binary"
  "io"
)

type Message interface {

  // get raw byteslice
  RawBytes() []byte
  // get the ircline out of this message
  Line() ircLine
  // time sent
  Sent() uint64
  // get the message command type
  Type() uint32
}


var URC_PLAINTEXT = uint32(0)
var URC_DHT_KAD = binary.BigEndian.Uint32([]byte{0x01, 0x01 , 0x01, 0x01}[:])




type MessageReader interface {
  ReadMessage(r io.Reader) (msg Message, err error)
}

type MessageWriter interface {
  WriteMessage(w io.Writer, msg Message) (err error)
}
