//
// message.go -- chat message protocol
//
package arc

import (
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


type MessageReader interface {
  ReadMessage(r io.Reader) (msg Message, err error)
}

type MessageWriter interface {
  WriteMessage(w io.Writer, msg Message) (err error)
}
