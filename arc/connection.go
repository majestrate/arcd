//
// connection.go -- network connections
//
package arc

import (
  "errors"
  "fmt"
  "io"
  "net"
)

type Connection io.ReadWriteCloser


func socksConnect(socksaddr string, socksport int, remoteaddr string, remoteport int) (conn Connection, err error) {
  conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d",socksaddr, socksport))
  if err == nil {
    req := make([]byte, len(remoteaddr) + 11)
    req[0] = '\x04'
    req[1] = '\x01'
    req[2] = byte(remoteport & 0xff00 >> 8)
    req[3] = byte(remoteport & 0x00ff)
    req[7] = '\x01'
    req[8] = '\x30'
    copy(req[10:], []byte(remoteaddr))
    _, err = conn.Write(req)
    if err == nil {
      resp := make([]byte, 8)
      _, err = io.ReadFull(conn, resp)
      if resp[1] != '\x5a' {
        err = errors.New("cannot connect via socks proxy")
      }
    }
  }
  return
}
