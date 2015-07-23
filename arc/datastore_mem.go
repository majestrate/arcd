//
// datastore_mem.go -- in memory chunk store
//
package arcd

import (
  "log"
)

type ramChunkStore struct {
  total int
  left int
  store map[CryptoHash]DHTPutPayload
}

func (self ramChunkStore) StoreChunk(data DHTPutPayload) {
  h := data.CalcHash()
  if _, ok := self.store[h] ; ok {
    // we already have it?
    log.Printf("did not put hash %s already exists?", h)
  } else {
    self.store[h] = data
    self.left --
  }
}

func (self ramChunkStore) GetChunk(h CryptoHash) (val DHTPutPayload, ok bool) {
  val, ok = self.store[h]
  return
}

func (self ramChunkStore) TotalCapacity() uint64 {
  return uint64(self.total) * uint64(2048)
}

func (self ramChunkStore) CapacityUsed() uint64 {
  return self.TotalCapacity() - ( uint64(self.left) * uint64(2048) )
}

func createRamChunkStore(chunks int) ChunkStore {
  return ramChunkStore{chunks, chunks, make(map[CryptoHash]DHTPutPayload)}
}

