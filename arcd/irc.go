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
  Send chan string
  conn net.Conn
  reader *bufio.Reader
  Nick string
  channels map[string]bool
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
  log.Println("too many clients connected")
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
        if self.clients[idx].acceptMessage(line) {
          self.clients[idx].Send <- line
        }
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

func (self *IRC) Init(conn net.Conn) {
  self.conn = conn
  self.reader = bufio.NewReader(self.conn)
  self.Send = make(chan string, 8)
  self.channels = make(map[string]bool)
}

func (self *IRC) SendLine(data string) {
  self.Send <- data
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

func (self *IRC) end() {
  self.conn.Close() 
}

func parseIRCLine(line string) (source, action, target, message string) {
  if ! strings.HasPrefix(line, ":") {
    return 
  }
  parts := strings.Split(line, " ")
  if len(parts) < 3 {
    return 
  }
  
  if len(parts) >= 4 {
    idx := strings.Index(parts[3], ":")
    if idx == -1 {
      return 
    }
    message = line[idx+1:]
  }

  source = parts[0][1:]
  action = strings.ToUpper(parts[1])
  target = parts[2]
  
  return
}

func (self *IRC) acceptMessage(line string) bool {
  source, action, target, _ := parseIRCLine(line)
  var nick string
  if strings.Count(source, "!") > 0 {
    nick = strings.Split(source, "!")[0]
  }
  
  if action == "PRIVMSG" {
    if channelNameValid(target) {
      _, ok := self.channels[target]
      return ok && self.Nick != nick
    }
    return target == self.Nick
  }
  if action == "JOIN" || action == "PART" {
    if channelNameValid(target) {
      _, ok := self.channels[target]
      return ok //|| self.Nick == nick
    } 
  }
  return false
}

func (self *IRC) WriteMessages() {
  for {
    line := <- self.Send
    log.Println("write ", line)
    buff := bytes.NewBufferString(line)
    buff.WriteString("\n")
    _, err := self.conn.Write(buff.Bytes())
    if err != nil {
      self.end()
      break
    }
  }
}

func (self *IRC) ReadMessages() {
  for {
    line := self.ReadLine()
    if line == "" {
      self.end()
      break
    }
    if strings.HasPrefix(line, "PING ") {
      if line[5] == ':' {
        line = ":arcd PONG " + line[5:]
      } else {
        line = ":arcd PONG :" + line[5:]
      }
      self.SendLine(line)
      continue
    }
    if len(self.Nick) > 0 {
      if strings.HasPrefix(line , "PRIVMSG ") {
        ircline := fmt.Sprintf(":%s!user@arcd %s", self.Nick, line)
        self.Daemon.Broadacst <- NewArcIRCLine(ircline)
      } else if strings.HasPrefix(line, "JOIN ") {
        chans := strings.Split(line[5:], ",")
        for idx := range(chans) {
          self.JoinChannel(chans[idx])
        }
      } else if strings.HasPrefix(line, "PART ") {
        chans := strings.Split(line[5:], ",") 
        for idx := range(chans) {
          self.PartChannel(chans[idx])
        }
      }
    } else if strings.HasPrefix(line, "NICK ") {
      idx := strings.Index(line, ":" )
      if idx > 0 {
        self.Nick = line[idx+1:]
      } else {
        self.Nick = line[5:]
      }
      log.Println(self.Nick)
      self.Greet()
    } else {
      log.Println(line)
    }
  }
}

func (self *IRC) SendNum(num, target, data string) {
  var line string
  if target == "" {
    line = fmt.Sprintf(":arcd %s %s", num, data)
  } else {
    line = fmt.Sprintf(":arcd %s %s :%s", num, target, data)
  }
  self.SendLine(line)
}

func channelNameValid(name string) bool {
  if ! strings.HasPrefix(name, "#")  {
    return false
  }
  if strings.Contains(name, " ") {
    return false
  }
  return true
}

func (self *IRC) JoinChannel(chname string) {
  if ! channelNameValid(chname) {
    self.SendNum("403", chname ,"No such channel")
    return
  } 
  _, ok := self.channels[chname]
  if !ok {
    self.channels[chname] = true
    line := fmt.Sprintf(":%s!user@arcd JOIN :%s", self.Nick, chname)
    self.Send <- line
    self.Daemon.Broadacst <- NewArcIRCLine(line)
  } else {
    self.SendNum("443", fmt.Sprintf("%s %s", self.Nick, chname), "already on channel")
  }

}

func (self *IRC) PartChannel(chname string) {
  if ! channelNameValid(chname) {
    self.SendNum("403", chname ,"No such channel")
    return
  } 
  var put bool
  _, ok := self.channels[chname]
  if ok {
    delete(self.channels, chname)
    put = true
  }
  if put {
    line := fmt.Sprintf(":%s!user@arcd PART :%s", self.Nick, chname)
    self.Daemon.Broadacst <- NewArcIRCLine(line)
  } else {
    self.SendNum("442", chname, "you are not on that channel")
  }
}

func (self *IRC) Motd() {
  self.SendNum("375", self.Nick, "- arcd MOTD")
  self.SendNum("372", self.Nick, "- benis")
  self.SendNum("376", self.Nick, "- RPL_ENDOFMOTD")
}

// greet new user
func (self *IRC) Greet() {
  self.SendNum("001", self.Nick, "Welcome to the Internet Relay Network "+self.Nick+"!user@arcd")
  self.SendNum("002", self.Nick, "Your host is arcd, running version arcd 0.0.1")
  self.SendNum("003", self.Nick, "This Server was created sometime")
  self.SendNum("004", self.Nick, "arcd 0.0.1 xbw mb")
  self.Motd()
}