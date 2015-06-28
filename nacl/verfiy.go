package nacl

// #include <sodium.h>
// #cgo pkg-config: libsodium
import "C"

import (
  "log"
)

// verify a signed message
func CryptoVerify(smsg, pk []byte) bool {
  smsg_buff := NewBuffer(smsg)
  defer smsg_buff.Free()
  pk_buff := NewBuffer(pk)
  defer pk_buff.Free()

  if pk_buff.size != C.crypto_sign_publickeybytes() {
    log.Println("nacl.CryptoVerify() invalid public key size", len(pk))
    return false
  }
  mlen := C.ulonglong(0)
  msg := malloc(C.size_t(len(smsg)))
  defer msg.Free()
  return C.crypto_sign_open(msg.uchar(), &mlen, smsg_buff.uchar(), C.ulonglong(len(smsg)), pk_buff.uchar()) == 0
}

// verfiy a detached signature
// return true on valid otherwise false
func CryptoVerifyDetached(msg, sig, pk []byte) bool {
  msg_buff := NewBuffer(msg)
  defer msg_buff.Free()
  sig_buff := NewBuffer(sig)
  defer sig_buff.Free()
  pk_buff := NewBuffer(pk)
  defer pk_buff.Free()

  if pk_buff.size != C.crypto_sign_publickeybytes() {
    log.Println("nacl.CryptoVerifyDetached() invalid public key size", len(pk))
    return false
  }
  
  // invalid sig size
  if sig_buff.size != C.crypto_sign_bytes() {
    log.Println("nacl.CryptoVerifyDetached() invalid signature length", len(sig))
    return false
  }
  return C.crypto_sign_verify_detached(sig_buff.uchar(), msg_buff.uchar(), C.ulonglong(len(msg)), pk_buff.uchar()) == 0
}
