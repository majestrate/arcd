package arcd

import (
  "bytes"
  "log"
)

const KBUCKET_HASH_BITS uint = 256
const KBUCKET_SCALE uint = 8
const KBUCKET_NUMBER uint = KBUCKET_HASH_BITS / KBUCKET_SCALE

type KBucket struct {
  Data []byte
  Next *KBucket
}

func (self *KBucket) Contains(hash []byte) bool {
  bucket := self
  for {
    if bucket.Data == nil {
      return false
    }
    if bytes.Equal(bucket.Data, hash) {
      return true
    }
    if bucket.Next == nil {
      return false
    }
    bucket = bucket.Next
  }
}

func (self *KBucket) Append(hash []byte ) {
  bucket := self
  for {
    if bucket.Data == nil {
      break
    }
    bucket = bucket.Next
  }
  bucket.Next = new(KBucket)
  bucket.Data = hash
}

func (self *KBucket) Remove(hash []byte) {
  
  current := self
  prev := self
  
  for {
    // it's not here
    if current.Data == nil {
      return
    }
    if bytes.Equal(current.Data, hash) {
      if current.Next == nil {
        // first / last
        current.Data = nil
        prev.Next = nil
        return
      } else {
        // middle node
        prev.Next = current.Next
        current.Data = nil
        return
      }
    } else {
      prev = current
      current = prev.Next
    }
  }
}

// maps FileBlockHash to RouterInfos
type RoutingTable struct {
  Buckets [KBUCKET_NUMBER]KBucket
  OurHash []byte
}

func (self *RoutingTable) Init() {
  log.Println("initiailize routing table")
  
  self.Insert(self.OurHash)
}

func (self *RoutingTable) Insert(hash []byte) {
  log.Println("insert hash into routing table", FormatHash(hash))
  idx := bucketIndexForHash(hash)
  self.Buckets[idx].Append(hash)
}

func (self *RoutingTable) Remove(hash []byte) {
  idx := bucketIndexForHash(hash)
  self.Buckets[idx].Remove(hash)
}

// xor left and right to get distance
// assumes left and right are the same size
func xorBytes(left, right []byte) []byte {
  dist := make([]byte, len(left))
  for idx := range(left) {
    dist[idx] = left[idx] ^ right[idx] 
  }
  return dist
}


func countBits(data []byte) uint {
  var bits uint
  for idx := range(data) {
    b := data[idx]
    for c := 0 ; c < 8 ; c++ {
      if b & 0x01 == 1 {
        bits ++
      }
      b = b >> 1
    }
  }
  
  return bits
}

func getHashDistance(left, right []byte) uint {
  dist := xorBytes(left, right)
  return countBits(dist)
}

func bucketIndexForHash(hash []byte) uint {
  return countBits(hash) /  KBUCKET_SCALE
}

func (self *RoutingTable) GetClosestPeer(target []byte) []byte {
  dist := getHashDistance(self.OurHash, target)
  log.Println("kad dist=", dist)
  return nil
}