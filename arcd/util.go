package arcd

import (
  "crypto/sha1"
  "encoding/base32"
  "os"
  "strings"
  "time"
)

func copybytes(dst, src []byte, doff, soff, count uint) {
  var idx uint 
  for idx = 0 ; idx < count ; idx++ {
    dst[idx+doff] = src[idx+soff]
  }
}

func getshort(buff []byte, off uint) uint16 {
  var retval uint16
  retval = uint16(buff[off]) & 0xff
  return retval | (uint16( buff[1+off]) << 8 )
}

func putshort(num uint16, buff []byte, off uint) {
  buff[off+1] = byte(num) >> 8
  buff[off] = byte(num) & 0xff
}

func getlong(buff []byte, off uint) uint64 {
  var retval uint64
  retval = 0
  retval = retval | uint64(buff[0+off] << 56)
  retval = retval | uint64(buff[1+off] << 48)
  retval = retval | uint64(buff[2+off] << 40)
  retval = retval | uint64(buff[3+off] << 32)
  retval = retval | uint64(buff[4+off] << 24)
  retval = retval | uint64(buff[5+off] << 16)
  retval = retval | uint64(buff[6+off] << 8)
  retval = retval | uint64(buff[7+off])
  return retval
}

func putlong(num uint64, buff []byte, off uint) {
  buff[0+off] = byte(0xff00000000000000 & num >> 56)
  buff[1+off] = byte(0x00ff000000000000 & num >> 48)
  buff[2+off] = byte(0x0000ff0000000000 & num >> 40)
  buff[3+off] = byte(0x000000ff00000000 & num >> 32)
  buff[4+off] = byte(0x00000000ff000000 & num >> 24)
  buff[5+off] = byte(0x0000000000ff0000 & num >> 16)
  buff[6+off] = byte(0x000000000000ff00 & num >> 8)
  buff[7+off] = byte(0x00000000000000ff & num)
}

func TimeNow() uint64 {
  return uint64(time.Now().UTC().UnixNano())
}

func SHA1AsUInt64(data []byte) uint64 {
  digest := sha1.Sum(data)
  var retval uint64
  retval = 0
  for idx := 0 ; idx < 20 ; idx ++ {
    retval = retval | uint64(digest[idx] << uint(idx * 8))
  }
  return retval
}


func FileExists(fname string) bool {
  if _, err := os.Stat(fname) ; os.IsNotExist(err) {
    return false
  }
  return true
}

func FormatHash(data []byte) string {
  return strings.ToLower(strings.Trim(base32.HexEncoding.EncodeToString(data), "="))
}