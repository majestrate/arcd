//
// util.go -- various utility functions
//
package arc

import (
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
  return uint64(time.Now().Unix() + 4611686018427387914)
}
