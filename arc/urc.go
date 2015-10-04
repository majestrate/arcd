//
// urc.go -- urc link protocol
//
package arc

import (
  "crypto/rand"
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

func (u urcMessage) Type() uint32 {
  return binary.BigEndian.Uint32(u.hdr[14:18])
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
  if u.Type() == 0 {
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

func urcMessageFromURCLine(line string) urcMessage {
  var hdr urcHeader
  // random bytes
  io.ReadFull(rand.Reader, hdr[18:])
  // length
  binary.BigEndian.PutUint16(hdr[:2], uint16(len(line)))
  return urcMessage{
    body: []byte(line),
    hdr: hdr,
  }
}
