//
// datastore.go -- interfaces for storing chunk data
//
package arcd

// a way to store chunks
type ChunkStore interface {
  // store a chunk
  // remember its hash
  StoreChunk(data DHTPutPayload)
  // get a chunk given its hash, with true if it was found
  GetChunk(h CryptoHash) (DHTPutPayload, bool)
  // get total capacity in bytes
  TotalCapacity() uint64
  // how many bytes are used right now
  CapacityUsed() uint64
}

