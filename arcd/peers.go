package arcd

import (
  //"bufio"
  "bytes"
  "io/ioutil"
  //"log"
  "strings"
)

type Peer struct {
  Addr string
  Net string
  PubKey string
}

// serialize to bytes
func (self *Peer) Bytes() []byte {
  var buff bytes.Buffer
  buff.WriteString(self.Net)
  buff.WriteString(" ")
  buff.WriteString(self.Addr)
  buff.WriteString(" ")
  buff.WriteString(self.PubKey)
  return buff.Bytes()
}

func (self *Peer) Parse(line string) bool {
  parts := strings.Split(line, " ")
  if len(parts) != 3 {
      return false
  }
  self.PubKey = parts[2]
  self.Addr = parts[1]
  self.Net = parts[0]
  return true
}

type PeerFileLoader struct {
  Peers []Peer
}

func (self *PeerFileLoader) LoadFile(fname string) error {
  data, err := ioutil.ReadFile(fname)
  if err != nil {
    return err
  }
  lines := bytes.Split(data, []byte{'\n'})
  peers := make([]Peer, len(lines))
  peer_count := 0
  for idx := range(lines) {
    line := lines[idx]
    if bytes.HasPrefix(line, []byte{'#'}) {
      continue
    }
    var peer Peer
    if peer.Parse(string(line)) {
      peers[peer_count] = peer
      peer_count ++
    }
  }
  self.Peers = make([]Peer, peer_count)
  for idx := range(self.Peers) {
    self.Peers[idx] = peers[idx]
  }
  return nil
}