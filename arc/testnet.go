//
// testnet.go
// test network related utils
//

package arcd

import (
  "bytes"
  "io/ioutil"
  "log"
  "os"
)

type nodeTextFile string

func (self nodeTextFile) AddInfo(info PeerInfo) {
  log.Printf("add info to %s: %s", self, info)
  f, err := os.OpenFile(string(self), os.O_CREATE | os.O_WRONLY | os.O_APPEND | os.O_SYNC, 0600)
  if err != nil {
    log.Fatalf("cannot open %s: %s", self, err)
  }
  f.Write([]byte(info.String()))
  f.Write([]byte("\n"))
  err = f.Sync()
  if err != nil {
    log.Fatalf("cannot write to %s: %s", self, err)
  }
  f.Close()
}

// get every node's peer info
func (self nodeTextFile) AllNodes() []PeerInfo {
  var nodes []PeerInfo
  data, err := ioutil.ReadFile(string(self))
  if err != nil {
    lines := bytes.Split(data, []byte("\n"))
    for _, line := range(lines) {
      info := testnetPeerInfo(line)
      nodes = append(nodes, info)
    }
  }
  return nodes
}


func (self nodeTextFile) LookupNode(h CryptoHash) (PeerInfo, bool) {
  nodes, err := ioutil.ReadFile(string(self))
  if err != nil {
    lines := bytes.Split(nodes, []byte("\n"))
    for _, line := range(lines) {
      info := testnetPeerInfo(line)
      if info.NodeHash().Equal(h) {
        return info, true
      }
    }
  }
  return nil, false
}
