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


// parse an ircline
// :soource action target :message
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

// check if we should accept a line and pass it up to the user
func (self *IRC) acceptMessage(line string) bool {
  source, action, target, message := parseIRCLine(line)
  var nick string
  // extract nickname
  if strings.Count(source, "!") > 0 {
    nick = strings.Split(source, "!")[0]
  }

  // check for privmsg
  if action == "PRIVMSG" {
    if channelNameValid(target) {
      _, ok := self.channels[target]
      // this is a private message from a valid channel not from us
      return ok && self.Nick != nick
    }
    // this is a private message to us
    return target == self.Nick
  }
  // check for JOIN/PART
  if action == "JOIN" || action == "PART" {
    if channelNameValid(message) {
      // valid channel join/part
      _, ok := self.channels[message]
      return ok 
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

// handle irc client messages
func (self *IRC) ReadMessages() {
  for {

    line := self.ReadLine()
    // no more data
    if line == "" {
      self.end()
      break
    }
    // check for ping
    if strings.HasPrefix(line, "PING ") {
      if line[5] == ':' {
        line = ":arcd PONG " + line[5:]
      } else {
        line = ":arcd PONG :" + line[5:]
      }
      // send pong
      self.SendLine(line)
      continue
    }
    // are we registered?
    if len(self.Nick) > 0 {
      // is this a private message?
      if strings.HasPrefix(line , "PRIVMSG ") {
        ircline := fmt.Sprintf(":%s!user@arcd %s", self.Nick, line)
        // broadcast line
        self.Daemon.Broadacst <- NewArcIRCLine(ircline)
      } else if strings.HasPrefix(line, "JOIN ") {
        // join channel message
        chans := strings.Split(line[5:], ",")
        // join the channels
        for idx := range(chans) {
          self.JoinChannel(chans[idx])
        }
      } else if strings.HasPrefix(line, "PART ") {
        // part channel message
        chans := strings.Split(line[5:], ",")
        // part the channels
        for idx := range(chans) {
          self.PartChannel(chans[idx])
        }
      }
    } else if strings.HasPrefix(line, "NICK ") {
      // initial register process
      // i hate irc clients
      idx := strings.Index(line, ":" )
      if idx > 0 {
        self.Nick = line[idx+1:]
      } else {
        self.Nick = line[5:]
      }
      log.Println(self.Nick)
      // greet on register
      self.Greet()
    } else {
      // log line if unknown(?)
      log.Println(line)
    }
  }
}

// send a numeric response
func (self *IRC) SendNum(num, target, data string) {
  var line string
  if target == "" {
    line = fmt.Sprintf(":arcd %s %s", num, data)
  } else {
    line = fmt.Sprintf(":arcd %s %s :%s", num, target, data)
  }
  self.SendLine(line)
}

// determine if a channel name is valid
// TODO: make rfc compliant
func channelNameValid(name string) bool {
  if ! strings.HasPrefix(name, "#")  {
    return false
  }
  if strings.Contains(name, " ") {
    return false
  }
  return true
}

// join channel logic
func (self *IRC) JoinChannel(chname string) {
  // check for channel name validity
  if ! channelNameValid(chname) {
    self.SendNum("403", chname ,"No such channel")
    return
  }
  
  _, ok := self.channels[chname]

  // are we already in the channel?
  if ok {
    self.SendNum("443", fmt.Sprintf("%s %s", self.Nick, chname), "already on channel")
    return
  }
  self.channels[chname] = true
  // tell ourselves that we joined
  line := fmt.Sprintf(":%s!user@arcd JOIN %s", self.Nick, chname)
  self.Send <- line
  // send a broadcast line
  self.Daemon.Broadacst <- NewArcIRCLine(line)
}

// part the channel
func (self *IRC) PartChannel(chname string) {
  // channel name valid etc
  if ! channelNameValid(chname) {
    self.SendNum("403", chname ,"No such channel")
    return
  }

  // delete channel presence
  var put bool
  _, ok := self.channels[chname]
  if ok {
    delete(self.channels, chname)
    put = true
  }

  // announce result
  if put {
    line := fmt.Sprintf(":%s!user@arcd PART :%s", self.Nick, chname)
    // tell us we parted
    self.send <- line
    // broadcast the part
    self.Daemon.Broadacst <- NewArcIRCLine(line)
  } else {
    // we weren't on the channel
    self.SendNum("442", chname, "you are not on that channel")
  }
}

// send motd
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
