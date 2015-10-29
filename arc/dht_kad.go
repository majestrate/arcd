//
// dht_kad.go -- kademlia dht implementation
//
package arc


type kBucketEntry struct {
  key DHTKey
  value DHTValue
}

// a kbucket in our routing table
type kBucket struct {
  entries []kBucketEntry
}

type kadDHT struct {
  DHTHandler
  
  kbuckets [32]*kBucket
}


