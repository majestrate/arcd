//
// r5n.go
// r5n kademlia dht implementation
//
package arcd

import (
  "time"
)

type DistanceFunc func (k1, k2 CryptoHash) uint64

type kBucket struct {
  data []PeerInfo
}

// add a peer info to this kbucket
func (self kBucket) AddPeer(info PeerInfo) {
  if self.data == nil {
    self.data = make([]PeerInfo, 1)
  }
  self.data = append(self.data, info)
}

// get the closest PeerInfo to CryptoHash
func (self kBucket) GetClosest(h CryptoHash, dist DistanceFunc) (info PeerInfo, success bool) {
  // check for non empty bucket
  if self.data != nil && len(self.data) > 0 {
    var mindist uint64
    // big number
    mindist = 0xffffffffffffffff
    for _, val := range self.data {
      d := dist(h, val.NodeHash())
      if d < mindist {
        mindist = d
        info = val
        // we found something
        success = true
      }
    }
  }
  return
}

// the state of a relayed or self created transaction
type KadTransaction interface {
  StartTime() time.Time
  ExpireTime() time.Time
  Method() string
  Target() CryptoHash
}


// the result of a transaction
type TransactionResult int

const (
  TRANS_RESULT_OK = iota // transaction was okay
  TRANS_RESULT_TIMEOUT // transaction timed out
  TRANS_RESULT_REJECT // transaction was rejected right away
  TRANS_RESULT_FAIL // transaction failed somehow
)

// return string representation
func (self TransactionResult) String() string {
  switch self {
  case TRANS_RESULT_OK: return "[OK]"
  case TRANS_RESULT_TIMEOUT: return "[TIMEOUT]"
  case TRANS_RESULT_REJECT: return "[REJECTED]"
  case TRANS_RESULT_FAIL: return "[FAILED]"
  default: return "[UNKNOWN STATE]"
  }
}

// a running transaction
// acts as a future
type RunningKadTransaction interface {
  // return the result of this transaction when it's done
  Result() chan TransactionResult
}

type Kademlia interface {
  // get function that computes distance
  Distance() DistanceFunc
  // add a peer to the routing table
  AddPeer(info PeerInfo)
  // remove a peer from the routing table
  DelPeer(info PeerInfo)
  // run a transaction
  // register it with us so we don't replay it
  // return a Running Transaction that we wait for
  RegisterTransaction(t KadTransaction, id int64) RunningKadTransaction
  // do we have a transaction with this ID and CryptoHash?
  HasTransaction(h CryptoHash, id int64) bool
  // get the next hop for a CrpytoHash
  // return hash, success
  // does not affect state of routing table
  // does not register the hash as a transaction
  GetNextHop(h CryptoHash) (PeerInfo, bool)
}


// transaction tracker
type kTracker struct {
  // transactions in this ktracker
  data map[int64] KadTransaction
}

func createKTracker() kTracker {
  return kTracker{
    data: make(map[int64]KadTransaction),
  }
}

// kademlia dht routing table
type kademlia struct {
  // our nodehash
  ourHash CryptoHash
  // our routing table
  routingTable []kBucket
  // pending relayed transactions
  trans kTracker
}

// implements Kadelmia.Distance
func (self kademlia) Distance() DistanceFunc {
  return self.xorDistance
}

// compute distance with xor
func (self kademlia) xorDistance(k1, k2 CryptoHash) (dist uint64) {
  var d CryptoHash
  // xor it
  for idx, _ := range(d) {
    d[idx] = k1[idx] ^ k2[idx]
  }
  // count bits
  for _, b := range d {
    // un rolled
    if b & 0x01 == 1 {
      dist += 1
    }
    if ( b >> 1 ) & 0x01 == 1 {
      dist += 1
    }
    if ( b >> 2 ) & 0x01 == 1 {
        dist += 1
    }
    if ( b >> 3 ) & 0x01 == 1 {
        dist += 1
    }
    if ( b >> 4 ) & 0x01 == 1 {
        dist += 1
    }
    if ( b >> 5 ) & 0x01 == 1 {
        dist += 1
    }
    if ( b >> 6 ) & 0x01 == 1 {
        dist += 1
    }
    if ( b >> 7 ) & 0x01 == 1 {
        dist += 1
    }
  }
  
  return
}

// get the node hash of the next hop for a DHT message
func (self kademlia) GetNextHop(h CryptoHash) (info PeerInfo, success bool) {
  // get the closest kbucket
  idx := self.kbucketIndexFor(h)
  bucket := self.routingTable[idx]

  for {
    // get next closest bucket if we don't have it
    for idx < len(self.routingTable) {
      idx ++ 
      bucket = self.routingTable[idx]
    }
    if idx < len(self.routingTable) {
      info, success = bucket.GetClosest(h, self.Distance()) 
      // return if the lookup succeeded
      if success {
        return // lookup success
      }
    } else {
      break // no more buckets
    }
  }
  // lookup failed
  return 
}

// add a peer into our routing table
func (self kademlia) AddPeer(info PeerInfo) {
  hash := info.NodeHash()
  idx := self.kbucketIndexFor(hash)
  // add to kbucket
  self.routingTable[idx].AddPeer(info)
  
}

// obtain the index of the kbucket that this nodehash would go into
func (self kademlia) kbucketIndexFor(key [64]byte) int {
  dist := self.Distance()(self.ourHash, key)
  idx := dist / uint64(len(self.routingTable))
  return int(idx)
}

// delete a peer info from our routing table
func (self kademlia) DelPeer(info PeerInfo) {
  
}

// do we have a transaction?
func (self kademlia) HasTransaction(h CryptoHash, id int64) bool {
  tr , ok := self.trans.data[id]
  if ok {
    return tr.Target() == h
  }
  return false
}

// register a transaction, run it
func (self kademlia) RegisterTransaction(kt KadTransaction, id int64) RunningKadTransaction {
  // don't check for duplicates
  self.trans.data[id] = kt
  // don't create any running transaction yet
  // TODO: create running transaction
  return nil
}

func newKad(ourInfo PeerInfo, buckets int) Kademlia {
  return kademlia{
    ourHash: ourInfo.NodeHash(),
    routingTable: make([]kBucket, buckets),
    trans: createKTracker(),
  }
}

  
