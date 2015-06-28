//
// tcp link implementation
//
package arcd

import (
  "bytes"
  "io"
)

// tcp based link
type tcpLink struct {

  // implements Link
  Link

  sock_reader io.Reader
  sock_writer io.Writer
  sock_closer io.Closer

  // peer's public signing key
  sigkey PubSigKey
  // peer's public encryption key
  // currently unused
  enckey PubEncKey
}

// a tcp link level message
// inter router communication message
//
// format for version 0 is 70 byte header
// 1 byte protocol version
// 1 byte message type
// 2 byte message length
// 12 bytes padding (any value)
// 64 bytes ed25519 signature of body
type _tcpLinkMessageHeader [70]byte

// link level protocol version
func (self _tcpLinkMessageHeader) Version() byte {
  return self[0]
}

// what kind of message is this?
func (self _tcpLinkMessageHeader) Type() byte {
  return self[1]
}

// how big the Body of the message is
func (self _tcpLinkMessageHeader) BodyLength() int {
  return int((self[2] << 8) | self[3])
}

func (self _tcpLinkMessageHeader) Signature() Signature {
  return self[16:]
}

func (self _tcpLinkMessageHeader) ValidSize() bool {
  return len(self) == 1 + 1 + 2 + 12 + 64
}

type tcpLinkMessageHeader struct {
  LinkMessageHeader

  hdr _tcpLinkMessageHeader
  sock_reader io.Reader
}

func (self tcpLinkMessageHeader) Version() byte {
  return self.hdr.Version()
}

func (self tcpLinkMessageHeader) Type() byte {
  return self.hdr.Type()
}

func (self tcpLinkMessageHeader) Signature() Signature {
  return self.hdr.Signature()
}

func (self tcpLinkMessageHeader) ValidSize() bool {
  return self.hdr.ValidSize()
}

func (self tcpLinkMessageHeader) BodyReader() io.Reader {
  msglen := self.hdr.BodyLength()
  buff := make([]byte, msglen)
  _, err := io.ReadFull(self.sock_reader, buff)
  if err != nil {
    return nil
  }
  return bytes.NewBuffer(buff)
}

func (self tcpLink) RecvHeader() (LinkMessageHeader, error) {
  var tcpHdr tcpLinkMessageHeader
  tcpHdr.sock_reader = self.sock_reader
  _, err := io.ReadFull(self.sock_reader, tcpHdr.hdr[:])
  if err != nil {
    return nil, err
  }
  return tcpHdr, nil
}
