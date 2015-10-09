//
// filter.go -- bloom filter
//

package arc

import (
  "crypto/sha256"
  "encoding/binary"
  "math"
)

type bloomFilter [1024*32]byte

func (f *bloomFilter) getProbes(data []byte) []uint64 {
  h := sha256.Sum256(data)
  return []uint64{
    binary.LittleEndian.Uint64(h[:16]),
    binary.LittleEndian.Uint64(h[16:32]),
  }
}

func (f *bloomFilter) Contains(b []byte) bool {
  for _, probe := range f.getProbes(b) {
    if f[int(probe%uint64(len(f)))] & byte(math.Pow(2, float64(probe % 8))) != 0 {
      return true
    }
  }
  return false
}

func (f *bloomFilter) Add(b []byte) {
  for _, probe := range f.getProbes(b) {
    f[int(probe%uint64(len(f)))] |= byte(math.Pow(2, float64(probe % 8)))
  }
}
