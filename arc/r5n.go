//
// r5n.go
// r5n kademlia dht implementation
//
package arcd

// the part of the kbucket that holds 1 entry
type kBucketEntry struct {
  // the key for this entry
  key NodeHash
  // the value that this key is associated with
  val PeerInfo
}
// a kbucket is a variable sized array of kbucket entries
type kBucket []kBucketEntry


// r5n kademlia dht routing table
type R5NKademlia struct {
  ourKey kBucketEntry
  routingTable []kBucket
}

// compute distance with xor
func (self R5NKademlia) Distance(k1, k2 NodeHash) (dist uint64) {
  var d NodeHash
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

func (self R5NKademlia) GetNextHop(msg Message) (nh NodeHash, success bool) {
  return nh, false
}

func (self R5NKademlia) AddPeer(info PeerInfo, key NodeHash) {
  
}

// obtain the kbucket that this nodehash would go into
func (self R5NKademlia) kbucketFor(key NodeHash) kBucket {
  dist := self.Distance(self.ourKey.key, key)
  idx := dist / uint64(len(self.routingTable))
  return self.routingTable[int(idx)]
}


func newR5NKad(ourInfo PeerInfo, key NodeHash, buckets int) R5NKademlia {
  return R5NKademlia{kBucketEntry{key,ourInfo}, make([]kBucket, buckets)}
}

  
