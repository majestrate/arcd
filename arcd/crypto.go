package arcd

import (
  "bytes"
  "crypto"
  "crypto/rand"
  "crypto/rsa"
  "crypto/x509"
  "code.google.com/p/go.crypto/sha3"
  "crypto/sha512"
  "crypto/ecdsa"
  "crypto/elliptic"
  "io/ioutil"
  "log"
  "math/big"
)

func GenerateECC_256() (*ecdsa.PrivateKey, error) {
  return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func DumpECC_256(privkey *ecdsa.PrivateKey, fname string) error {
  data , err := x509.MarshalECPrivateKey(privkey)
  if err != nil {
    return err
  }
  return ioutil.WriteFile(fname, data, 0600)
}

func LoadECC_256(fname string) (*ecdsa.PrivateKey, error) {
  data, err := ioutil.ReadFile(fname)
  if err != nil {
    return nil, err
  }
  
  key, err := x509.ParseECPrivateKey(data)
  return key, err
}


func SignECC_256(data []byte, privkey *ecdsa.PrivateKey) ([]byte , error) {
  hash := SHA3_256(data)
  
  r, s, err := ecdsa.Sign(rand.Reader, privkey, hash)
  if err != nil {
    return nil , err
  }
  return packBigInts(r, s), nil
}

func VerifyECC_256(data, sig []byte, pubkey *ecdsa.PublicKey) bool {
  r, s := unpackBigInts(sig)
  if r == nil {
    log.Println("failed to parse signature")
    return false
  }
  hash := SHA3_256(data)
  return ecdsa.Verify(pubkey, hash, r, s)
}

func unpackBigInts(data []byte) (x, y *big.Int) {
  dlen := len(data) / 2
  return big.NewInt(0).SetBytes(data[:dlen]), big.NewInt(0).SetBytes(data[dlen:])
}

func packBigInts(x, y *big.Int) []byte {
  var buff bytes.Buffer
  buff.Write(x.Bytes())
  buff.Write(y.Bytes())
  return buff.Bytes()
}

func ECC_256_PubKeyBytes(pubkey ecdsa.PublicKey) []byte {
  return packBigInts(pubkey.X, pubkey.Y)
}

func ECC_256_UnPackPubKeyString(data string) ecdsa.PublicKey {
  buff := UnFormatHash(data)
  return ECC_256_UnPackPubKey(buff)
}

func ECC_256_UnPackPubKey(data []byte) ecdsa.PublicKey {
   x, y := unpackBigInts(data)
   var pubkey ecdsa.PublicKey
   pubkey.X = x
   pubkey.Y = y
   pubkey.Curve = elliptic.P256()
   return pubkey
}

func ECC_256_KeyHash(pubkey ecdsa.PublicKey) []byte {
  return SHA3_256(ECC_256_PubKeyBytes(pubkey))
}

func ECC_256_PubKey_ToString(pubkey ecdsa.PublicKey) string {
  return FormatHash(ECC_256_PubKeyBytes(pubkey))
}

func GenerateRSA4K() (*rsa.PrivateKey, error) {
  return rsa.GenerateKey(rand.Reader, 4096)
}

func DumpRSA4K(privkey *rsa.PrivateKey, fname string) error {
  data := x509.MarshalPKCS1PrivateKey(privkey)
  return ioutil.WriteFile(fname, data, 0600)
}

func LoadRSA4K(fname string) (*rsa.PrivateKey, error) {
  data, err := ioutil.ReadFile(fname)
  if err != nil {
    return nil, err
  }
  
  key, err := x509.ParsePKCS1PrivateKey(data)
  return key, err
}

func SignRSA4K(data []byte, privkey *rsa.PrivateKey) ([]byte, error) {
  //hash := SHA3_256(data)
  hash := SHA2_512(data)
  return rsa.SignPKCS1v15(rand.Reader, privkey, crypto.SHA512, hash)
}

func VerifyRSA4K(data, sig []byte,  pubkey *rsa.PublicKey) bool {
  //hash := SHA3_256(data)
  hash := SHA2_512(data)
  return rsa.VerifyPKCS1v15(pubkey, crypto.SHA512, hash, sig) == nil
}

func RSA4K_KeyHash(key rsa.PublicKey) []byte {
   return SHA2_512(key.N.Bytes())
}

func SHA3_256(data []byte) []byte {
  buff := make([]byte, 32)
  sum := sha3.Sum256(data)
  for idx := 0 ; idx < 32 ; idx ++ {
    buff[idx] = sum[idx]
  }
  return buff
}

func SHA2_512(data []byte) []byte {
  buff := make([]byte, 32)
  sum := sha512.Sum512(data)
  for idx := 0 ; idx < 32 ; idx ++ {
    buff[idx] = sum[idx]
  }
  return buff
}

func PubKeyHash_ToString(key rsa.PublicKey) string {
  return FormatHash(RSA4K_KeyHash(key))
}

func PubKey_ToString(key rsa.PublicKey) string {
  return FormatHash(key.N.Bytes())
}

func TestCrypto() {
  key, _ := GenerateECC_256()
  buff := make([]byte, 256)
  sig, _ := SignECC_256(buff, key)
  if ! VerifyECC_256(buff, sig, &key.PublicKey) {
    log.Fatal("crypto test failure")
  }
  log.Println("crypto test success")
}