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
    parts := strings.Split(string(line), " ")
    if len(parts) != 3 {
      continue
    }
    peer.PubKey = parts[2]
    peer.Addr = parts[1]
    peer.Net = parts[0]
    peers[peer_count] = peer
    peer_count ++
  }
  self.Peers = make([]Peer, peer_count)
  for idx := range(self.Peers) {
    self.Peers[idx] = peers[idx]
  }
  return nil
}