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


type kadRouter struct {
  ib chan Message
}

func (r kadRouter) InboundChan() chan Message {
  return r.ib
}

func (r kadRouter) Run(hubs ...Hub) {

}

// create kademlia message router
func NewKadRouter(keyfile string) Router {
  return kadRouter{
    ib: make(chan Message),
  }
}

type broadcastRouter struct {
  bc, ib chan Message
  filter bloomFilter
}

func (r broadcastRouter) InboundChan() chan Message {
  return r.ib
}

func (r broadcastRouter) Run(hubs ...Hub) {
  log.Println("run router")
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
        } else {
          // filter pass
          r.filter.Add(b)
          r.bc <- m
        }
      }
    }
  }
  log.Println("router exited")
}

// create broadcast style message 'router'
func NewBroadcastRouter(keyfile string) Router {
  return broadcastRouter{
    bc: make(chan Message, 16),
    ib: make(chan Message, 32),
  }
}
