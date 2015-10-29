//
// urc.go -- urc link protocol
//
package arc

import (
  "bytes"
  "crypto/rand"
  "encoding/binary"
  "fmt"
  "io"
  "log"
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
  return binary.LittleEndian.Uint32(u.hdr[14:18])
}

func (u urcMessage) RawBytes() (b []byte) {
  b = append(b, u.hdr[:]...)
  b = append(b, u.body...)
  return
}

func (u urcMessage) Sent() uint64 {
  return binary.BigEndian.Uint64(u.hdr[2:10])
}

func (u urcMessage) Line() ircLine {
  t := u.Type()
  if t == 0 {
    // plaintext
    return ircLine(u.body)
  } else if t == URC_DHT_KAD {
    // dht message
    idx := bytes.Index(u.body, []byte{10})
    return ircLine(u.body[:idx+1])
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
  binary.BigEndian.PutUint16(hdr[:], uint16(len(line)))
  binary.BigEndian.PutUint64(hdr[2:], timeNow())
  return urcMessage{
    body: []byte(line),
    hdr: hdr,
  }
}

func Privmsg(from, to, msg string) Message {
  return urcMessageFromURCLine(fmt.Sprintf(":%s!arc@arcnet PRIVMSG %s :%s\n", from, to, msg))
}

func urcDHT(method, k_b64, our_hexkey string, body io.Reader) Message {
  var hdr urcHeader
  io.ReadFull(rand.Reader, hdr[18:])
  // mark as dht message
  hdr[14] = 1
  hdr[15] = 1
  hdr[16] = 1
  hdr[17] = 1

  line := ""
  if body == nil {
    line = fmt.Sprintf(":arcd!%s@arcnet PRIVMSG #dht :%s %s\n", our_hexkey, method, k_b64)
    binary.BigEndian.PutUint16(hdr[:], uint16(len(line)))
  } else {
    // body attached
    var buff bytes.Buffer
    n, err := io.Copy(&buff, body)
    if err == nil {
      line = fmt.Sprintf(":arcd!%s@arcnet PRIVMSG #dht :%s %d\n",our_hexkey, method, n)
      l := n
      l += int64(len(line))
      if l <= 65536 {
        binary.BigEndian.PutUint16(hdr[:], uint16(l))
        body := make([]byte, int(l))
        copy(body, []byte(line))
        copy(body, buff.Bytes())
        return urcMessage{
          body: body,
          hdr: hdr,
        }
      } else {
        // too big
        log.Println("urc message too big", l, "bytes")
        return nil
      }
    } else {
      // error copying
      log.Println("error creating urc DHT message", err)
      return nil
    }
  }
  binary.BigEndian.PutUint64(hdr[2:], timeNow())
  
  return urcMessage{
    body: []byte(line),
    hdr: hdr,
  }
}
