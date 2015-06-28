package nacl

// #include <sodium.h>
// #cgo pkg-config: libsodium
//
// unsigned char * deref_uchar(void * ptr) { return (unsigned char*) ptr; }
//
//
import "C"

import (
  "encoding/hex"
  "log"
  "reflect"
  "unsafe"
)

// wrapper arround malloc/free
type Buffer struct {
  ptr unsafe.Pointer;
  length C.int;
  size C.size_t;
  
}

// wrapper arround nacl.malloc
func Malloc(size int) *Buffer {
  if size > 0 {
    return malloc(C.size_t(size))
  }
  return nil
}

// does not check for negatives
func malloc(size C.size_t) *Buffer {
  ptr := C.malloc(size)
  C.sodium_memzero(ptr, size)
  buffer := &Buffer{ptr: ptr, size: size , length: C.int(size)}
  return buffer
}

// create a new buffer copying from a byteslice
func NewBuffer(buff []byte) *Buffer {
  buffer := Malloc(len(buff))
  if buffer == nil {
    log.Println("nacl.NewBuffer() nacl.Malloc() failed")
    return nil
  }
  if copy(buffer.Data(), buff) != len(buff) {
    log.Println("nacl.NewBuffer() did not copy all bytes")
    return nil
  }
  return buffer
}

func (self *Buffer) uchar() *C.uchar {
  return C.deref_uchar(self.ptr)
}

func (self *Buffer) Length() int {
  return int(self.length)
}

// get immutable byte slice
func (self *Buffer) Bytes() []byte {
  buff := make([]byte, self.Length())
  copy(buff, self.Data())
  return buff
}

// get underlying byte slice
func (self *Buffer) Data() []byte {
  hdr := reflect.SliceHeader{
    Data: uintptr(self.ptr),
    Len: self.Length(),
    Cap: self.Length(),
  }
  return *(*[]byte)(unsafe.Pointer(&hdr))
}

func (self *Buffer) String() string {
  return hex.EncodeToString(self.Data())
}

// zero out memory and then free
func (self *Buffer) Free() {
  C.sodium_memzero(self.ptr, self.size)
  C.free(self.ptr)
}
