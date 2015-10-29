
//
// dht.go -- kademlia routed dht 
//
package arc

import (
  "bytes"
  "io"
)

type DHTKey [32]byte

type DHTDistance [32]byte

func maxDist() (d DHTDistance) {
  for idx, _ := range d {
    d[idx] = 0xff
  }
  return 
}

// return true if the distance is zero
func (d DHTDistance) Zero() bool {
  for _, b := range d {
    if b != 0 {
      return false
    }
  }
  return true
}

// return true if we are less than the other
func (d DHTDistance) LessThan(other DHTDistance) bool {
  for idx, b := range other {
    if b == d[idx] {
      continue
    } else if d[idx] < b {
      return true
    } else {
      return false
    }
  }
  return false
}

// is this key equal to other
func (d DHTDistance) Equal(other DHTDistance) bool {
  return bytes.Compare(d[:], other[:]) == 0
}

// get distance between this key and another key
func (k DHTKey) Distance(other DHTKey) (d DHTDistance) {
  for idx, b := range k {
    d[idx] = b ^ other[idx]
  }
  return
}

// a value that can be inserted into the dht
// could be on disk, in memory, inside goatse, etc
type DHTValue interface {
  io.Reader
  io.Writer
  Hash() [32]byte
}

const (
  DHT_SUCCESS = iota // operation success
  DHT_FAIL_OVERLOAD // operation failed because we are overloaded
  DHT_FAIL_TIMEOUT // operation failed because the transaction timed out
  DHT_FAIL_PENDING // operation failed because we are already fucking doing it GOD DAMN
  DHT_FAIL_CANCEL // operation failed because a relayer shut down
)

type DHTRequester interface {
  // get a value given a key
  // returns value and 0 on success
  // returns nil and non-zero on fail
  GET(k DHTKey) (DHTValue, int)

  // put a value locally
  // return 0 on success otherwise error code
  PUT(v DHTValue) (int)
}

type DHTHandler interface {
  // return key and value if the request is accepted locally
  // return key and nil if the request is to be relayed
  // return nil and nil if the request is to be rejected
  HandleGET(k DHTKey) (DHTKey, DHTValue)

  // return non-zero on fail or zero on success
  HandlePUT(v DHTValue) (int)
}
