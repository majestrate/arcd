//
// message_dhtput.go dht put message impl
//

package arcd

import (
  "log"
)

// encrypted data that can be put
// padded to 2048 bytes with 2 byte length at first
type DHTPutPayload [2050]byte

// get a byte slice of the data
func (self DHTPutPayload) Bytes() []byte {
  slicelen := int( self[0] << 8 ) | int(self[1])
  return self[2:slicelen]
}

func (self DHTPutPayload) CalcHash() CryptoHash {
  return cryptoHash(self.Bytes())
}

// create a DHTPutPayload from a byteslice
func createChunk(data []byte) DHTPutPayload {
  var payload DHTPutPayload
  payload_len := len(data)
  if payload_len > len(payload) - 2 {
    log.Fatalf("createChunk() cannot make a chunk of size %s too big", payload_len)
  }
  // put length
  payload[0] = byte((payload_len & 0xff00) >> 8)
  payload[1] = byte(payload_len & 0x00ff)
  // copy data
  copy(payload[2:], data)
  // return it
  return payload
}

// break up a byteslice into a DHTPutPayload slice to be inserted
// pads as needed
func createInsertChunks(data []byte) []DHTPutPayload {
  var chunks []DHTPutPayload
  // XXX: hardcoded
  for len(data) > 2048 {
    chunk := createChunk(data[:2048])
    data = data[2048:]
    chunks = append(chunks, chunk)
  }
  return append(chunks, createChunk(data))
}


// handler of PUT messages
type dhtPutHandler struct {
  kad Kademlia
  store ChunkStore
}

// reply to a dht put
func (self dhtPutHandler) CreateReplyFor(msg DHTMessage) (dmsg DHTMessage, info PeerInfo) {
  // immediately store
  payload := msg.PUT()
  log.Println("storing chunk")
  self.store.StoreChunk(payload)
  d_msg := bencDHTMessage{
    reply: true,
    tid: msg.ID(),
    method: "PUT",
    hops: 0,
  }
  // call dump
  // this makes the dht message raw byte slice have data
  err := d_msg.Dump()
  if err == nil {
    return d_msg, nil
  }
  log.Printf("failed to create reply to DHT PUT, %s", err)
  return nil, nil
}

// create a DHTMessageProcessor for PUT
func createPutReplier(kad Kademlia, store ChunkStore) DHTMessageProcessor {
  return dhtPutHandler{kad: kad, store: store}
}
