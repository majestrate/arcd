//
// link level udp protocol
//
// protocol version 0:
//
// a udp link layer connection allows 1 way authenticated communication where the initiator knows the identity of the remote peer
//
// to initiate a session with a remote peer the initiator sends an intro message
// the intro message is encrypted to the remote peer's long term encryption key
// the intro message's encrypted payload contains:
//   * an ephemeral public key to encrypt replies to
//   * the initial nounce to use with the ephemeral session key
//   * zero or more "options" for future use
//
// in reply to an intro message the remote peer sends to the initator an intro confirm message
// the intro confirm message is encrypted to the previously provided ephemeral public key with nounce
// the intro confirm message's encrypted payload contains:
//   * an ephemeral public key to use to in place of the long term encryption key
//   * the initial nounce to use with the ephemeral session key
//
// once the intro message confirm message is recv'd, the initiator can now send kad messages to the remote peer and recv replies from the remote peer
// the remote peer can only send replies that the initiator sent, the remote peer cannot send kad messages down this session to the initator.
//
package arcd

import (
  "encoding/binary"
  "errors"
  "log"
  "net"

  "github.com/majestrate/arcd/nacl"
)

// info about a udp peer
type udpPeerInfo struct {
  // our ephemeral private encryption key
  ourKey CryptoBox_PrivKey
  // current nounce in use
  nounce []byte
  // their public encryption key to use for this session
  theirKey CryptoBox_PubKey
  // their long term encryption key used for introduction
  identKey CryptoBox_PubKey
  // the udp address we are to send to
  endpoint *net.UDPAddr
  // extra options (for future use)
  options map[string][]byte
}

const udp_packet_mtu = 1280
const udp_packet_header = 16

func udpFragmentDataSize() int  {
  return udp_packet_mtu - udp_packet_header - nacl.CryptoBoxOverhead()
}
// messages are sent as fragments in random order
// this describes such fragment, everything is big endian byte order
//
// as of udp link protocol version 0:
// message fragments when sent are always 1280 Bytes
// a message fragment is padded to fit this size
// any packets that are over 1280 bytes have the trailing bytes ignored
// any packets that are under 1280 bytes are dropped
//
type udpLinkMessageFragment [udp_packet_mtu]byte

// the ID of message this fragment belongs to
// first 8 bytes
func (self udpLinkMessageFragment) MessageID() uint64 {
  return binary.BigEndian.Uint64(self[:8])
}
// this fragment's ID
// if this is 0 then there are no more fragments in this message
// 4 bytes
func (self udpLinkMessageFragment) FragmentID() uint32 {
  return binary.BigEndian.Uint32(self[8:12])
}

// flags
// 2 bytes
//
// as of udp link protocol version 0:
// if flags[0] is set to 0x04 then this is an ack to a message fragment
// if flags[0] is set to 0x03 then this is a message fragment
// if flags[0] is set to 0x02 then this is an intro confirm message, messageID and fragmentID must be zero
// if flags[0] is set to 0x01 then this is an intro message, messageID and fragmentID must be zero
// flags[0] is never zero and the rest of flags must be zero
//
func (self udpLinkMessageFragment) Flags() []byte {
  return self[12:14]
}

// fragment size
// 2 bytes
// size of the fragment data in bytes
// trailing data in the fragment is padding and is ignored
func (self udpLinkMessageFragment) FragmentSize() uint16 {
  return binary.BigEndian.Uint16(self[14:16])
}

// the data of the fragment
// ( 1280 - nacl.CryptoBoxOverhead() - 16 ) bytes
func (self udpLinkMessageFragment) FragmentData() []byte {
  fragmentLen := int(self.FragmentSize())
  if 16 + fragmentLen < len(self) {
    return self[16:16+fragmentLen]
  }
  log.Printf("fragment size specified is too large for packet: fragmentSize=%d packetSize =%d", fragmentLen, len(self))
  return nil
}

// put a udp link message fragment into a buffer
func putUdpLinkMessageFragment(frag []byte, msgId uint64, fragId uint32, flags, data []byte) {
  // ensure size
  if len(data) < 65536 {
    binary.BigEndian.PutUint64(frag[:], msgId)
    binary.BigEndian.PutUint32(frag[8:], fragId)
    copy(frag[12:], flags)
    binary.BigEndian.PutUint16(frag[14:], uint16(len(data)))
    copy(frag[16:], data)
  } else {
    log.Println("udp link message fragment data too large", len(data))
  }

}

// describes the state of a udp link message in transit
// can be either inbound or outbound
type udpLinkMessageTransitState struct {

}

// return true if we have completed transferring
func (udpLinkMessageTransitState) Completed() bool {
  return false
}

// session state for a remote udp endpoint
type udpPeerState struct {
  // the remote peer's info
  info udpPeerInfo
  // message we are recv-ing now
  messages_recv []udpLinkMessageTransitState
  // message we are sending now
  messages_send []udpLinkMessageTransitState
}

// create and encrypt an intro message
// this peer's info must be set before
func (self udpPeerState) MakeIntroMessage() []byte {
  // set our ephemeral private key
  keys := nacl.GenBoxKeypair()
  defer keys.Free()
  self.info.ourKey = keys.Secret()
  self.info.nounce = nacl.NewBoxNounce()
  // set options, empty for now
  self.info.options = make(map[string][]byte)
  pubkey := keys.Public()
  hdr_len := len(pubkey) + len(self.info.nounce)
  pkt_data := make([]byte, hdr_len)
  // packet format is public key + nounce
  copy(pkt_data, pubkey)
  copy(pkt_data[len(pubkey):], self.info.nounce)

  // set flags
  pkt_flags := make([]byte, 2)
  // intro message
  pkt_flags[0] = 1

  // make packet
  pkt := nacl.RandBytes(udp_packet_mtu - hdr_len)
  putUdpLinkMessageFragment(pkt, 0, 0, pkt_flags, pkt_data)
  // encrypt and return
  return self.info.identKey.EncryptAnon(pkt)
}

type udpLink struct {
  identKey CryptoBox_PrivKey
  udp_socket *net.UDPConn
}


func (self udpLink) Bind(addr string) error {
  if self.udp_socket != nil {
    return errors.New("already bound")
  }
  uaddr, err := net.ResolveUDPAddr("ipv6", addr)
  if err != nil {
    return err
  }
  self.udp_socket, err = net.ListenUDP("ipv6", uaddr)
  return err
}

func (self udpLink) Mainloop() {
  
}
