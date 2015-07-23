//
// crypto.go
// cryptography interfaces
//
package arcd

// generic signature
type Signature []byte

// public key for signing
type PubSigKey interface {
  // verify a signature of data
  VerifySig(data []byte, sig Signature)
  // get bytes
  Bytes() []byte
}


// private key for signing
type PrivSigKey interface {
  // sign data and create signature
  Sign(data []byte) Signature
  // get the public key for this private key
  PubKey() PubSigKey
  // get the raw bytes
  Bytes() []byte
}

// public encryption key
type PubEncKey interface {
  // encrypt data for this public key
  Encrypt(data []byte) []byte
  Bytes() []byte
}

type PrivEncKey interface {
  // decrypt data for this private key
  Decrypt(data []byte) []byte
  Public() PubEncKey
  Bytes() []byte
}
