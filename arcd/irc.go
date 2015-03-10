package arcd

import (
  "bufio"
  "bytes"
  "fmt"
  "log"
  "net"
  "strings"
)

type IRCD struct {
  BindAddr string
  Daemon *Daemon
  Broadcast chan string
  clients []*IRC
}

type IRC struct {
  Daemon *Daemon
  Broadcast chan string
  conn net.Conn
  reader *bufio.Reader
  Nick string
}

func (self *IRC) SendLine(data string) {
  self.Broadcast <- data
}

func (self *IRC) ReadLine() string {
  data, err  := self.reader.ReadBytes('\n')
  if err != nil {
    return ""
  }
  var line string
  line = string(data[:len(data)-2])
  log.Println("read ", line)
  return line
}

func (self *IRC) Init(conn net.Conn) {
  self.conn = conn
  self.reader = bufio.NewReader(self.conn)
  self.Broadcast = make(chan string, 8)
}

func (self *IRCD) Init(daemon *Daemon) {
  self.Daemon = daemon
  self.clients = make([]*IRC, 128)
  self.Broadcast = make(chan string)
}

func (self *IRCD) clientAdd(handler *IRC) {
  for idx := range(self.clients) {
    if self.clients[idx] == nil {
      self.clients[idx] = handler
      return
    }
  }
  log.Fatal("too many clients connected")
}

func (self *IRCD) clientRemove(handler *IRC) {
  for idx := range(self.clients) {
    if handler == self.clients[idx] {
      self.clients[idx] = nil
    }
  }
}

func (self *IRCD) Run() {
  go self.Accept()
  for {
    line := <- self.Broadcast
    for idx := range(self.clients) {
      if self.clients[idx] != nil {
        self.clients[idx].Broadcast <- line
      }
    }
  }
}

func (self *IRCD) Accept() {
  addr, err := net.ResolveTCPAddr("tcp6", self.BindAddr)
  listen, err := net.ListenTCP("tcp6", addr)
  if err != nil {
    log.Fatal(err)
  }
  log.Println("ircd bound at", listen.Addr())
  for {
    conn , err := listen.Accept()
    if err != nil {
      log.Fatal(err)
    }
    log.Println("inbound irc", conn.RemoteAddr())
    go self.InboundIRC(conn)
  }
}

func (self *IRCD) InboundIRC(conn net.Conn) {
  var irc IRC
  irc.Daemon = self.Daemon
  irc.Init(conn)
  self.clientAdd(&irc)
  go irc.WriteMessages()
  irc.ReadMessages()
  self.clientRemove(&irc)
}

func (self *IRC) WriteMessages() {
  for {
    line := <- self.Broadcast
    log.Println("write ", line)
    buff := bytes.NewBufferString(line)
    buff.WriteString("\n")
    _, err := self.conn.Write(buff.Bytes())
    if err != nil {
      break
    }
  }
}

func (self *IRC) ReadMessages() {
  for {
    line := self.ReadLine()
    if line == "" {
      break
    }
    if strings.HasPrefix(line, "PING ") {
      line := ":arcd PONG :"+ line[5:]
      self.SendLine(line)
      continue
    }
    if len(self.Nick) > 0 {
      if strings.HasPrefix(line , "PRIVMSG ") {
        ircline := fmt.Sprintf(":%s!user@arcd %s", self.Nick, line)
        self.Daemon.Broadacst <- NewArcIRCLine(ircline)
      }
      
      
    } else if strings.HasPrefix(line, "NICK ") {
      self.Nick = line[5:]
      self.Greet()
      self.JoinDefaultChannel()
    }
  }
}

func (self *IRC) JoinDefaultChannel() {
  self.SendLine(":"+self.Nick+"!user@arcd JOIN :#arcnet")
}

func (self *IRC) Motd() {
  
  self.SendLine(":arcd 375 "+self.Nick+" :- arcd MOTD")
  self.SendLine(":arcd 372 "+self.Nick+" :- benis")
  self.SendLine(":arcd 376 "+self.Nick+" :RPL_ENDOFMOTD")
}

// greet new user
func (self *IRC) Greet() {
  self.SendLine("001 :"+self.Nick)
  self.SendLine("002 :"+self.Nick+"!user@arcd")
  self.SendLine("003 :arcd")
  self.SendLine("004 arcd 0.0 :+")
  self.SendLine("005 NETWORK=ARCNET CHANTYPES=#&!+ CASEMAPPING=ascii CHANLIMIT=25 NICKLEN=25 TOPICLEN=128 CHANNELLEN=16 COLOUR=1 UNICODE=1 PRESENCE=0:")
  self.Motd()
}