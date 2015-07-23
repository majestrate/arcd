//
// link.go
// link level inter router communication interfaces
//
package arcd

import (
  "io"
)

// Information about a peer
type PeerInfo interface {
  // get the network address of this peer
  NetAddr() string
  // the network to use in net.Dial
  Net() string
  // the node hash of this peer
  NodeHash() CryptoHash
  // get string version
  String() string
}

type LinkMessageHeader interface {
  // protocol version
  Version() byte
  // signature of body
  GetSig() Signature
  // reader that reads the body of the message
  BodyReader() io.Reader
  // is this header correctly formed
  Valid() bool
  // raw bytes
  Bytes() []byte
}


// creator of link messages
type LinkMessageFactory interface {
  // create the link level messages we send for this dht message
  // splits and orders as needed
  CreateMessagesForDHT(msg DHTMessage) []Message
  //CreateMessagesForN2N(msg N2NMessage) []Message
}

// inter router link
type Link interface {
  // channel that others poll on for messages from this link
  MessageChan() chan Message
  // recv a message header
  // error != nil on error
  RecvHeader() (LinkMessageHeader, error)
  // send a message
  // sends header and then message body
  // Blocks
  SendMessage(msg Message, to_peer PeerInfo) error
  // run mainloop
  Mainloop()
  // get the message factory for this
  GetMessageFactory() LinkMessageFactory
}
