//
// hub.go -- message router
//
package arc

import (
  "fmt"
  "log"
  "net"
  "time"
)

type Hub interface {
  // persist a remote urc connection
  PersistURC(addr string, port int, proxyType, proxyAddr string, proxyPort int)
  // run all operations
  Run()
}

type basicHub struct {
  bind string
  keyfile string
  broadcast chan Message
  registerConn chan Connection
  deregisterConn chan Connection
  conns map[Connection]bloomFilter
}

func (h basicHub) handleURC(conn Connection) {
  h.registerConn <- conn
  urc := urcProtocol{}
  var err error
  f := h.conns[conn]
  for {
    var umsg urcMessage
    umsg, err = urc.ReadMessage(conn)
    if err == nil {
      f.Add(umsg.RawBytes())
      h.broadcast <- umsg
    } else {
      // error is fatal
      log.Println("error in urc handler", err)
      break
    }
  }
  h.deregisterConn <- conn
}

func (h basicHub) PersistURC(addr string, port int, proxyType, proxyAddr string, proxyPort int) {
  if proxyType == "socks" {
    log.Printf("persist hub %s:%d proxy=%s://%s:%d", addr, port, proxyType, proxyAddr, proxyPort)
    go func() {
      for {
        time.Sleep(time.Second)
        log.Println("connecting to hub", addr)
        conn, err := socksConnect(proxyAddr, proxyPort, addr, port)
        if err == nil {
          log.Println("connected to", addr)
          h.handleURC(conn)
        } else{ 
          log.Println("cannot connect to", addr, err)
        }
      }
    }()
  } else {
    log.Printf("persist hub %s:%d", addr, port)
    go func() {
      for {
        time.Sleep(time.Second)
        log.Println("connecting to hub", addr)
        conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
        log.Println("connected to", addr)
        if err == nil {
          h.handleURC(conn)
        } else {
          log.Println("cannot connect to", addr, err)
        }
      }
    }()
  }
}

func (h basicHub) Run() {
  log.Println("run hub")
  // connection -> is inbound
  for {
    select {
    case c := <- h.registerConn:
      h.conns[c] = bloomFilter{}
    case c := <- h.deregisterConn:
      delete(h.conns, c)
      c.Close()
    case m := <- h.broadcast:
      log.Println(m.URCLine())
      b := m.RawBytes()
      for c, f := range h.conns {
        if f.Contains(b) {
          // filter hit
          continue
        }
        // add to bloom filter
        f.Add(b)
        // relay it
        send := len(b)
        sent := 0
        for {
          n, err := c.Write(b[sent:])
          if err == nil {
            sent += n
            if sent < send {
              // continue sending it's a short write
              continue
            } else {
              break
            }
          } else {
            // error writing
            log.Println("failed to write message", err)
            h.deregisterConn <- c
            break
          }
        }
      }
    }
  }
}



func CreateHub(addr, keyfile string) Hub {
  return basicHub{
    bind: addr,
    keyfile: keyfile,
    broadcast: make(chan Message),
    registerConn: make(chan Connection),
    deregisterConn: make(chan Connection),
    conns: make(map[Connection]bloomFilter),
  }
}


