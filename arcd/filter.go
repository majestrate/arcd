package arcd

import (
  //"math"
)

const BLOOM_FILTER_BITS = 20

type DecayingBloomFilter struct {
  array map[uint64]bool
  bins, probes uint
  counter uint16
}

func (self *DecayingBloomFilter) Init() {
  // initialize bins
  self.Decay()
}

func (self *DecayingBloomFilter) Decay() {
  self.array = make(map[uint64] bool)
}

func (self *DecayingBloomFilter) Add(data []byte) {
  idx := SHA1AsUInt64(data) 
  self.array[idx] = true
  self.counter ++
  if self.counter == 0 {
    self.Decay()
  }
}

func (self *DecayingBloomFilter) Contains(data []byte) bool {
  idx := SHA1AsUInt64(data) 
  return self.array[idx] 
}
