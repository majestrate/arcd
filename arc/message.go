//
// link level messages
//
package arcd

// message types
const (
  MSG_TYPE_NULL = iota
  MSG_TYPE_DHT // dht message
  MSG_TYPE_N2N // node to node message
)


// capacity flags for dht
type DHTCapacity [8]byte

// dht messages
type DHTMessage interface {
  // access raw byte slice
  Bytes() []byte
  // how many hops left in random walk
  Hops() int64
  // dht method, put/get/find/caps
  Method() string
  // is this a reply?
  Reply() bool
  // transaction id
  ID() int64
  // access the payload for a put
  PUT() DHTPutPayload
  // access the requested hash for a get
  GET() CryptoHash
  // access the capacity flags for a caps
  CAPS() DHTCapacity
}

// node to node messages
type NodeMessage []byte


// link level message
type Message interface {
  // sender of this message
  Source() PeerInfo
  // access as byte slice
  Bytes() []byte
  // message type
  Type() byte
  // parse this as a dht message
  AsDHTMessage() DHTMessage
  // parse this as a node to node message
  AsNodeMessage() NodeMessage
}


// maker of replies to DHT messages
type DHTMessageProcessor interface {
  // create a reply message
  // nil if we don't reply
  // returns dht message, who to send to
  CreateReplyFor(msg DHTMessage) (DHTMessage, PeerInfo)
}

// creator of dht messages
type DHTMessageFactory interface {
  // create a new dht message
  CreateMessage() DHTMessage
}
