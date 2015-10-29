//
// util.go -- various utility functions
//
package arc

import (
  "encoding/base64"
  "encoding/hex"
  "os"
  "time"
)

// check if a file exists
func checkFile(fname string) bool {
  if _, err := os.Stat(fname) ; os.IsNotExist(err) {
    return false
  }
  return true
}

func timeNow() uint64 {
  // because taia96
  return uint64(time.Now().Unix() + 4611686018427387914)
}

func dumpBuffer(b [32]byte) string {
  return "0x" + hex.EncodeToString(b[:])
}

func keyToBytes(b64 string) (b []byte) {
  d, err := base64.StdEncoding.DecodeString(b64)
  if err == nil {
    b = make([]byte, 32)
    // decode okay
    copy(b, d[:32])
    return
  }
  // decode failed
  return nil
}
