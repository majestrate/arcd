//
// irc.go -- irc s2s bridge
//
package arc

import (
  "bufio"
  "fmt"
  "io"
  "log"
  "net"
  "strings"
  "time"
)


type ircLine string

func (line ircLine) Command() (cmd string) {
  l := string(line)
  if strings.HasPrefix(l, ":") {
    parts := strings.Split(l, " ")
    if len(parts) > 1 {
      cmd = parts[1]
    }
  }
  return
}

func (line ircLine) Param() (param string) {
  l := string(line)
  // get index of : minus the first char
  idx := strings.Index(l[1:], ":")
  return l[1+idx:]
}

type ircBridge struct {
  io.ReadWriteCloser
}

// read lines and send ircLines down a channel to be processed
func (irc ircBridge) produce(chnl chan Message) (err error) {
  sc := bufio.NewScanner(irc)
  // for each line
  for sc.Scan() {
    
    line := sc.Text()
    log.Println(line)
    l := ircLine(line)
    cmd := l.Command()
    // accept certain commands
    for _, c := range []string{"NOTICE", "PRIVMSG", "JOIN", "PART", "QUIT"} {
      if cmd == c {
        chnl <- urcMessageFromURCLine(line)
        break
      }
    }
  }
  err = sc.Err()
  return
}

type ircAuthInfo string


// linkname component
func (info ircAuthInfo) Name() string {
  return "arcnet.tld"
}

// linkpass component
func (info ircAuthInfo) Pass() string {
  return string(info)
}

// write a line
func (irc ircBridge) Line(format string, args ...interface{}) (err error) {
  _, err = fmt.Fprintf(irc, format, args...)
  _, err = io.WriteString(irc, "\n")
  return
}

func (irc *ircBridge) handshake(auth ircAuthInfo) (err error) {
  sc := bufio.NewScanner(irc)
  // send pass line
  err = irc.Line("PASS %s", auth.Pass())
  err = irc.Line("SERVER %s 1", auth.Name())
  for err == nil && sc.Scan() {
    line := ircLine(sc.Text())
    log.Println("irchub hanshake:", line)
    if line.Command() == "PING" {
      // send a pong if we got a ping
      err = irc.Line(":%s PONG :%s", auth.Name(), line.Param())
      // we have handshaked
      log.Println("irchub handshake good")
      return
    }
  }
  return
}


// an irc / urc bridge
type ircHub struct {
  ib, ob chan Message
  // register / deregister ircline writer
  regis, deregis chan chan ircLine
  router Router
}

func (h ircHub) Send(m Message) {
  h.ob <- m
}


func (h *ircHub) runConnection(c io.ReadWriteCloser, auth ircAuthInfo) (err error) {
  chnl := make(chan ircLine)
  irc := ircBridge{c}
  err = irc.handshake(auth)
  if err == nil {
    h.regis <- chnl
    // write messages
    go func() {
      for {
        line, ok := <- chnl
        if ok {
          irc.Line("%s", line)
        } else {
          return
        }
      }
    }()
    // read messages
    irc.produce(h.ib)
    h.deregis <- chnl
  }
  return
}

func (h ircHub) Persist(c RemoteHubConfig) {
  for {
    // sleep for backoff
    time.Sleep(time.Second)
    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.Addr, c.Port))
    if err == nil {
      err = h.runConnection(conn, ircAuthInfo(c.Password))
      if err == nil {
        log.Println("irchub connection ended")
      } else {
        log.Println("irchub connection error", err)
      }
    } else {
      // error
      log.Println("irchub error", err)
    }
  }
}

func (h ircHub) Run() {
  log.Println("run irc hub")
  conns := make(map[chan ircLine]bool)
  for {
    select {
    case m, ok := <- h.ib:
      if ok {
        h.router.InboundChan() <- m
      }
    case m, ok := <- h.ob:
      if ok {
        line := m.Line()
        for chnl := range conns {
          chnl <- line
        }
      }
    case chnl, ok := <- h.regis:
      if ok {
        log.Println("irchub connection registerd")
        conns[chnl] = true
      }
    case chnl, ok := <- h.deregis:
      if ok {
        delete(conns, chnl)
      }
    }
  }
}

func (h ircHub) Close() {
  close(h.ib)
  close(h.ob)
  close(h.regis)
  close(h.deregis)
}

func NewIRCHub(r Router) Hub {
  return ircHub{
    ib: make(chan Message),
    ob: make(chan Message),
    regis: make(chan chan ircLine),
    deregis: make(chan chan ircLine),
    router: r,
  }
}
