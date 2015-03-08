package arcd

import (
  "crypto"
  "crypto/rand"
  "crypto/rsa"
  "crypto/x509"
  "code.google.com/p/go.crypto/sha3"
  "io/ioutil"
)


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
  hash := SHA3_256(data)
  return rsa.SignPSS(rand.Reader, privkey,crypto.SHA3_256, hash,  nil)
}

func VerifyRSA4K(data, sig []byte,  pubkey *rsa.PublicKey) bool {
  hash := SHA3_256(data)
  return rsa.VerifyPSS(pubkey, crypto.SHA3_256, hash, sig, nil) == nil
}

func RSA4K_KeyHash(key *rsa.PublicKey) []byte {
   return SHA3_256(key.N.Bytes())
}

func SHA3_256(data []byte) []byte {
  buff := make([]byte, 32)
  sum := sha3.Sum256(data)
  for idx := 0 ; idx < 32 ; idx ++ {
    buff[idx] = sum[idx]
  }
  return buff
}