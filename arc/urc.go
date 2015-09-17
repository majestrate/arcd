//
// urc.go -- urc link protocol
//
package arc

import (
  "encoding/binary"
  "io"
)

type urcHeader [26]byte

func (h urcHeader) Length() uint16 {
  return binary.BigEndian.Uint16(h[:2])
}

type urcMessage struct {
  hdr urcHeader
  body []byte
}

func (u urcMessage) RawBytes() (b []byte) {
  b = append(b, u.hdr[:]...)
  b = append(b, u.body...)
  return
}

func (u urcMessage) Sent() uint64 {
  return binary.BigEndian.Uint64(u.hdr[2:10])
}

func (u urcMessage) URCLine() string {
  if u.hdr[14] == '\x00' {
    // plaintext
    return string(u.body)
  }
  return ""
}

type urcProtocol struct {
  
}

// read a urc link message
func (urc urcProtocol) ReadMessage(r io.Reader) (msg urcMessage, err error) {
  _, err = io.ReadFull(r, msg.hdr[:])
  if err == nil {
    l := int(msg.hdr.Length())
    msg.body = make([]byte, l)
    _, err = io.ReadFull(r, msg.body)
  }
  return
}


type urcConnection io.ReadWriteCloser
