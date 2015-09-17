//
// util.go -- various utility functions
//
package arc

import (
  "os"
)

// check if a file exists
func checkFile(fname string) bool {
  if _, err := os.Stat(fname) ; os.IsNotExist(err) {
    return false
  }
  return true
}
