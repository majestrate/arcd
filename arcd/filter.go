package arcd

import (
  //"math"
)

const BLOOM_FILTER_BITS = 20

type DecayingBloomFilter struct {
  array []byte
  bins, probes uint
}

func (self *DecayingBloomFilter) Init() {
  // initialize bins
  self.Decay()
}

func (self *DecayingBloomFilter) Decay() {
  self.array = nil
  self.array = make([]byte, 1024 * 8)
}

func (self *DecayingBloomFilter) Add(data []byte) {
  idx := SHA1AsUInt64(data) % uint64(len(self.array))
  self.array[idx] = 1
}

func (self *DecayingBloomFilter) Contains(data []byte) bool {
  idx := SHA1AsUInt64(data) % uint64(len(self.array))
  return self.array[idx] == 1
  /*
  probes := self.get_probes(data)
  for idx := range(probes) {
    i := probes[idx]
    i2 := uint( i  % 8 )
    if self.array[i / 8] & byte( math.Pow( 2.0, float64(i2) ) ) != 0 {
      return true
    }
  }
  return false
  */
}
