//
// router.go -- message router
//
package arc

import (
  "log"
)

// generic router interface
// routes messages as needed
type Router interface {
  InboundChan() chan Message
  Run(hubs ...Hub)
}

type broadcastRouter struct {
  bc, ib chan Message
  filter bloomFilter
}

func (r broadcastRouter) InboundChan() chan Message {
  return r.ib
}

func (r broadcastRouter) Run(hubs ...Hub) {
  log.Println("run broadcast router")
  for {
    select {
    case m, ok := <- r.bc:
      if ok {
        for _, h := range hubs {
          h.Send(m)
        }
      } else {
        break
      }
    case m, ok := <- r.ib:
      if ok {
        b := m.RawBytes()
        if r.filter.Contains(b) {
          // filter hit
          log.Println("fitler hit")
        } else {
          // filter pass
          r.filter.Add(b)
          log.Println("offer")
          r.bc <- m
        }
      }
    }
  }
  log.Println("broadcast router exited")
}

// create broadcast style message 'router'
func NewBroadcastRouter(keyfile string) Router {
  return broadcastRouter{
    bc: make(chan Message, 16),
    ib: make(chan Message, 32),
  }
}


type kadRouter struct {
  ib chan Message
  dht kadDHT
  bc chan Message
  filter bloomFilter
}

func (r kadRouter) InboundChan() chan Message {
  return r.ib
}

func (r kadRouter) Run(hubs ...Hub) {
  log.Println("run kad router")
  for {
    select {
    case m, ok := <- r.ib:
      if ok {
        t := m.Type()
        line := m.Line()
        if t == URC_DHT_KAD {
          // for our router
          log.Println("dht message:", line)
        } else {
          // for broacdast router
          log.Println("hub message:", line)
          b := m.RawBytes()
          if r.filter.Contains(b) {
            // filter hit
            log.Println("fitler hit")
          } else {
            // filter pass
            r.filter.Add(b)
            log.Println("offer")
            r.bc <- m
          }
        }
      } else {
        // error on channel?
        break
      }
    case m, ok := <- r.bc:
      if ok {
        for _, h := range hubs {
          h.Send(m)
        }
      } else {
        break
      }
    }
  }
  log.Println("kad router exited")
}

// create kademlia message router
func NewKadRouter(keyfile string) Router {
  return kadRouter{
    ib: make(chan Message, 16),
    bc: make(chan Message, 32),
  }
}
