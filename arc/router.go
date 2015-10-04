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
  RegisterHub(h Hub) Router
  Run()
}


type kadRouter struct {
  ib chan Message
}

func (r kadRouter) InboundChan() chan Message {
  return r.ib
}

func (r kadRouter) RegisterHub(h Hub) Router {
  return r
}

func (r kadRouter) Run() {

}

// create kademlia message router
func NewKadRouter(keyfile string) Router {
  return kadRouter{
    ib: make(chan Message),
  }
}

type broadcastRouter struct {
  bc, ib chan Message
  hubs []Hub
}

func (r broadcastRouter) InboundChan() chan Message {
  return r.ib
}

func (r broadcastRouter) RegisterHub(h Hub) Router {
  r.hubs = append(r.hubs, h)
  return r
}

func (r broadcastRouter) Run() {
  log.Println("run router")
  for {
    select {
    case m, ok := <- r.bc:
      if ok {
        for _, h := range r.hubs {
          log.Println("send hubs")
          h.SendChan() <- m
        }
      } else {
        break
      }
    case m, ok := <- r.ib:
      if ok {
        log.Println(">>",m.URCLine())
        r.bc <- m
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
