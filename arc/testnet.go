//
// testnet.go
// test network related utils
//

package arcd

import (
  "bytes"
  "io/ioutil"
)

type nodeTextFile string

func (self nodeTextFile) LookupNode(nh NodeHash) (PeerInfo, bool) {
  nodes, err := ioutil.ReadFile(string(self))
  if err != nil {
    lines := bytes.Split(nodes, []byte("\n"))
    for _, line := range(lines) {
      info := testnetPeerInfo(line)
      if info.NodeHash().Equal(nh) {
        return info, true
      }
    }
  }
  return nil, false
}
