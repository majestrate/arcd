//
// hub.go -- link level message hub
//
package arc

import (
  "fmt"
  "log"
  "net"
  "time"
)

// base type for arcd router
type Hub interface {
  // get a channel for others to send messages on
  SendChan() chan Message
  // persist a connection
  Persist(addr string, port int, proxyType, proxyAddr string, proxyPort int)
  // run all operations
  Run()
  // close hub
  Close()
}

type basicHub struct {
  // bind address
  bind string
  // privkey file
  keyfile string
  // send broadcast message channel
  broadcast chan Message
  // register connection channel
  registerConn chan Connection
  // register connection channel
  deregisterConn chan Connection
  // connection map
  conns map[Connection]bloomFilter
  // message router
  router Router
}

func (h basicHub) SendChan() chan Message {
  return h.broadcast
}

func (h basicHub) Close() {
  close(h.broadcast)
  close(h.registerConn)
  close(h.deregisterConn)
}

// handle a urc connection inbound outbound doesn't matter
func (h basicHub) handleURC(conn Connection) {
  // register our connection
  h.registerConn <- conn
  // new protocol state
  urc := urcProtocol{}
  var err error
  // get our filter
  f := h.conns[conn]
  for {
    var umsg urcMessage
    // read a message
    umsg, err = urc.ReadMessage(conn)
    if err == nil {
      // add the raw bytes of this message to our bloom filter
      f.Add(umsg.RawBytes())
      // tell router of inbound message
      h.router.InboundChan() <- umsg
    } else {
      // error is fatal
      log.Println("error in urc handler", err)
      break
    }
  }
  // deregister connection we are done
  h.deregisterConn <- conn
}

// persist a connection to a remote hub
func (h basicHub) Persist(addr string, port int, proxyType, proxyAddr string, proxyPort int) {
  if proxyType == "socks" {
    // use socks proxy
    log.Printf("persist hub %s:%d proxy=%s://%s:%d", addr, port, proxyType, proxyAddr, proxyPort)
    go func() {
      for {
        // cooldown
        time.Sleep(time.Second)
        log.Println("connecting to hub", addr)
        // connect to socks proxy
        conn, err := socksConnect(proxyAddr, proxyPort, addr, port)
        if err == nil {
          log.Println("connected to", addr)
          // handle connection
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
        // cooldown
        time.Sleep(time.Second)
        log.Println("connecting to hub", addr)
        // dial out
        conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
        log.Println("connected to", addr)
        if err == nil {
          // handle connection
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
      // register a connection
      // give it a new bloom filter
      h.conns[c] = bloomFilter{}
    case c := <- h.deregisterConn:
      // deregeister a connection
      // delete it from the list of connections
      delete(h.conns, c)
      // close the connection
      c.Close()
    case m := <- h.broadcast:
      // we want to send a broadcast line
      b := m.RawBytes()
      // for each connection
      for c, f := range h.conns {
        // check filter
        if f.Contains(b) {
          // filter hit, don't send it this way
          continue
        }
        // add to bloom filter
        f.Add(b)
        // relay it
        send := len(b)
        sent := 0
        log.Println("broadcast urc")
        for {
          // TODO: go routine for each connection?
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


// create a new hub
// parameters are bind address and private key file
func CreateHub(addr, keyfile string, r Router) Hub {
  return basicHub{
    bind: addr,
    keyfile: keyfile,
    broadcast: make(chan Message),
    registerConn: make(chan Connection),
    deregisterConn: make(chan Connection),
    conns: make(map[Connection]bloomFilter),
    router: r,
  }
}


