//
// router.go
// arc router interfaces
//
package arcd

import (
  "bytes"
)

type RouterMain interface {
  Configure(map[string]string) RouterMain
  Run()
}

// non-colliding hash of a node on the network
type NodeHash [64]byte

func (self NodeHash) Equal(other NodeHash) bool {
  return bytes.Equal(self[:], other[:])
}

// measures the distance between keys
type RoutingMetric interface {
  // get distance between the hash of 2 items
  Distance(k1, k2 NodeHash) uint64
}

type RoutingTable interface {
  // Given a dht message find the next node to forward it to
  // return the node's hash and if the operation succeeded or not
  GetNextHop(msg DHTMessage) (NodeHash, bool)
  // put a peer into the routing table
  // remember their info is associated with their key
  // update routing table after
  AddPeer(info PeerInfo, key NodeHash)
  // forget about a peer
  // update routing table after
  DelPeer(info PeerInfo)
}

type NodeDatabase interface {
  // lookup a node ident given a hash
  // return a PeerInfo and true if success
  // return nil and false if failed
  LookupNode(nh NodeHash) (PeerInfo, bool)
}
