//
// crypto_impl.go
// implemtation of cryptographic interfaces
//
package arcd

import (
  // nacl cgo bindings
  "github.com/majestrate/arcd/nacl"

  "bytes"
  "crypto/sha512"
  "encoding/base64"
)

// cryptographic hash
type CryptoHash [64]byte

func (self CryptoHash) String() string {
  return base64.StdEncoding.EncodeToString(self[:])
}

func (self CryptoHash) Equal(h CryptoHash) bool {
  return bytes.Equal(self[:], h[:])
}


type CryptoBox_PrivKey []byte
type CryptoBox_PubKey []byte

type CryptoSign_PrivKey []byte
type CryptoSign_PubKey []byte


// decrypt given data, nounce and public key of sender
func (self CryptoBox_PrivKey) Decrypt(data, nounce, pk []byte) []byte {
  return nacl.CryptoBoxOpen(data, nounce, self, pk)
}

func (self CryptoBox_PrivKey) DecryptAnon(data []byte) []byte {
  nounce_len := nacl.CryptoBoxOverhead()
  pubkey_len := nacl.CryptoBoxPubKeySize()
  return self.Decrypt(data[nounce_len+pubkey_len:], data[:nounce_len], data[nounce_len:nounce_len+pubkey_len])
}

// decrypt given data, nounce and secret key of sender
func (self CryptoBox_PubKey) Encrypt(data, nounce, sk []byte) []byte {
  return nacl.CryptoBoxOpen(data, nounce, self, sk)
}

// use an ephemeral keypair to encrypt
// bundle nounce and public key with encrypted message
func (self CryptoBox_PubKey) EncryptAnon(data []byte) []byte {
  keypair := nacl.GenBoxKeypair()
  defer keypair.Free()
  nounce := nacl.NewBoxNounce()
  pubkey := keypair.Public()
  body := self.Encrypt(data, nounce, keypair.Secret())
  hdr := append(nounce, pubkey...)
  return append(hdr, body...)
}

// wrapper for nacl.RandBytes
func randBytes(size int) []byte {
  return nacl.RandBytes(size)
}

// crypto hash function
func cryptoHash(data []byte) CryptoHash {
  var h CryptoHash
  digest := sha512.Sum512(data)
  copy(h[:], digest[:])
  return h
}
