//
// irc.go -- irc s2s bridge
//
package arc

import (
  "bufio"
  "fmt"
  "io"
)


type ircLine string

type ircBridge struct {
  io.ReadWriteCloser
}

// reads irc lines and processes them
type ircReader struct {
  io.Reader
}

// run the reader, send ircLines down a channel to be processed
func (r ircReader) Process(chnl chan ircLine) (err error) {
  sc := bufio.NewScanner(r)
  for sc.Scan() {
    chnl <- ircLine(sc.Text())
  }
  err = sc.Err()
  return
}

type ircAuthInfo string


// linkname component
func (info ircAuthInfo) Name() string {
  return "arcd.irc.bridge.tld"
}

// linkpass component
func (info ircAuthInfo) Pass() string {
  return string(info)
}

// write a line
func (irc ircBridge) Line(format string, args ...interface{}) (err error) {
  _, err = fmt.Fprintf(irc, format, args)
  _, err = io.WriteString(irc, "\n")
  return
}

// handshake with a server we are connected to
// use auth to authenticate
func (irc ircBridge) handshake(auth ircAuthInfo) (err error) {
  err = irc.Line("PASS %s", auth.Pass())
  err = irc.Line("SERVER %s 1", auth.Name())
  return
}


type ircHub struct {
  ib, ob chan Message
}
