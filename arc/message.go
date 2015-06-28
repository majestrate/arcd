//
// link level messages
//
package arcd


// message types
const (
  MSG_TYPE_NULL = iota
  MSG_TYPE_DHT
  MSG_TYPE_N2N
)

// dht messages
type DHTMessage []byte
// node to node messages
type NodeMessage []byte


type Message interface {
  // access as byte slice
  Bytes() []byte
  // message type
  Type() byte
  // parse this as a dht message
  AsDHTMessage() DHTMessage
  // parse this as a node to node message
  AsNodeMessage() NodeMessage
}

