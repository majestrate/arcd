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


func extractNick(str string) (nick string) {
  idx := strings.Index(str, "!")
  if idx > 1 {
    nick = str[:idx]
  } else {
    nick = str
  }
  return
}

type ircLine string

func (line ircLine) Command() (cmd string) {
  l := string(line)
  if strings.HasPrefix(l, ":") {
    parts := strings.Split(l, " ")
    if len(parts) > 1 {
      cmd = parts[1]
    }
  } else {
    parts := strings.Split(l, " ")
    cmd = parts[0]
  }
  return
}

func (line ircLine) Source() (src string) {
  l := string(line)
  if strings.HasPrefix(l, ":") {
    src = strings.Split(l[1:], " ")[0]
  }
  return
}

func (line ircLine) Target() (targ string) {
  l := string(line)
  if strings.HasPrefix(l, ":") {
    parts := strings.Split(l, " ")
    if len(parts) > 2 {
      targ = parts[2]
    }
  } else {
    parts := strings.Split(l, " ")
    if len(parts) > 1 {
      targ = parts[1]
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
  name, nick string
  // urc user -> last message
  urcusers map[string]int64
  // channel -> presence
  chans map[string]ircChannel
  // nick -> is from irc server
  nicks map[string]bool
  
}

// an irc channel presence
type ircChannel struct {
  nicks []string
  topic string
}

// read lines and send ircLines down a channel to be processed
func (irc *ircBridge) produce(chnl chan Message) (err error) {
  // for each line
  sc := bufio.NewScanner(irc)
  log.Println("irchub produce")
  for sc.Scan() {
    
    line := sc.Text()
    log.Println("irchub server2hub", line)
    l := ircLine(line)
    cmd := l.Command()
    target := l.Target()
    
    switch cmd  {
    case "PING":
      // server ping
      irc.Line(":%s PONG %s", irc.name, l.Param())
      break
    case "SERVER":
      // we got a server command from the remote, we are connected
      log.Println("we have connected to the ircd")
      irc.Line("NICK %s :1", irc.nick)
      irc.Line(":%s USER serverlink arcd arcd :arc network", irc.nick)
      irc.Line(":%s JOIN #status", irc.nick)
      irc.Line(":%s PRIVMSG #status :arcnet link up", irc.nick)
      irc.nicks[irc.nick] = false
      break
    case "NICK":
      // register nick
      irc.nicks[target] = true
      break
    }
    // accept certain commands
    for _, c := range []string{"NOTICE", "PRIVMSG", "JOIN", "PART", "QUIT"} {
      if cmd == c {
        m := urcMessageFromURCLine(line)
        chnl <- m
        break
      }
    }
  }
  err = sc.Err()
  return
}

// consume messages from hub
func (irc *ircBridge) consume(chnl chan ircLine) {
  for {
    line, ok := <- chnl
    log.Println("irchub consume", line)
    if ok {
      src := line.Source()
      target := line.Target()
      nick := extractNick(src)
      cmd := line.Command()
      param := line.Param()
      switch cmd {
      case "NICK":
        local, ok := irc.nicks[nick]
        if ok {
          // we are tracking this guy
          if local {
            // this nick is local to the irc server
            // don't forward it
          } else {
            local, ok = irc.nicks[target]
            if ok {
              if local {
                // this nick is local to the irc server
                // someone spoofed the nickchange
                // don't forward
              } else {
                // this nick is not local but it is tracked
                // change names
                delete(irc.nicks, nick)
                irc.nicks[target] = false
                irc.Line("%s", line)
              }
            } else {
              // we don't have this nick, track it, it's remote
              irc.nicks[target] = false
              irc.Line("%s", line)
            }
          }
        }
        break
      case "PRIVMSG":
        local, ok := irc.nicks[nick]
        if ok {
          // nick presence tracked
          if local {
            // this nick is local to the irc server
            // don't forward it
            break
          } else {
            irc.Line("%s", line)
          }
        } else {
          // not tracked
          // is remote
          irc.nicks[nick] = false
          irc.Line("NICK %s :1", nick)
          irc.Line(":%s USER user arcd arcd :remote user", nick)
          irc.Line(":%s MODE %s +i", nick, nick)
          irc.Line(":%s JOIN %s", nick, target)
          irc.Line(":%s PRIVMSG %s :%s", nick, target, param)
        }
        break
      }
    } else {
      return
    }
  }
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
  log.Println("irchub hub2server", fmt.Sprintf(format, args...))
  _, err = io.WriteString(irc, "\n")
  return
}

func (irc *ircBridge) handshake(auth ircAuthInfo) (err error) {
  // send pass line
  err = irc.Line("PASS %s", auth.Pass())
  err = irc.Line("SERVER %s 1", auth.Name())
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
  irc := ircBridge{c, auth.Name(), "archub", make(map[string]int64),make(map[string]ircChannel), make(map[string]bool)}
  err = irc.handshake(auth)
  if err == nil {
    h.regis <- chnl
    // write messages
    go irc.consume(chnl)
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
