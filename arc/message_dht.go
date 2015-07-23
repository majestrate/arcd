//
// message_dht.go
// dht message implementation
//

package arcd

import (
  "bytes"
  "errors"
  "log"
)


type bencDHTMessage struct {
  // raw data
  raw []byte
  // dht fields
  method string
  reply bool
  tid int64
  hops int64
  data []byte
}

// dump values into raw field
func (self bencDHTMessage) Dump() (err error) {
  data := make(map[string]interface{})
  // set reply 
  if self.reply {
    data["r"] = "1"
  }
  // set method
  data["m"] = self.method
  // set transaction id
  data["t"] = self.tid
  // set hop count
  data["h"] = self.hops
  // set data
  data["d"] = self.data
  var buff bytes.Buffer
  // encode
  err = bencodeEncodeMap(data, &buff)
  self.data = buff.Bytes()
  return err
}

func (self bencDHTMessage) Parse() (err error) {
  buff := bytes.NewBuffer(self.raw)
  data, err := bencodeDecodeMap(buff)
  if err == nil {
    var ok bool
    var val interface{}
    // if this is a reply get the method
    val, self.reply = data["r"]
    if ! self.reply {
      val , ok = data["m"]
      if ! ok {
        return errors.New("message has no method")
      }
    }
    switch val.(type) {
    case string:
      self.method = string(val.([]byte))
      break
    default:
      return errors.New("method not string")
    }
    
    // get the transaction id
    var tid interface{}
    tid, ok = data["t"]
    if ok {
      switch tid.(type) {
      case int64:
        self.tid = tid.(int64)
        break
      default:
        return errors.New("transaction id isn't int")
      }
    }
    // get the hop count
    val , ok = data["h"]
    if ok {
      switch val.(type) {
      case int64:
        self.hops = val.(int64)
        break
      default:
        return errors.New("hops isn't int")
      }
    } else {
      // hops not defined
      self.hops = 0
    }
    // data field, this is interpreted differently for each method
    val, ok = data["d"]
    if ok {
      switch val.(type) {
      case []byte:
        copy(self.data, val.([]byte))
      }
    }
  }
  return
}

func (self bencDHTMessage) PUT() DHTPutPayload {
  var payload DHTPutPayload
  if len(self.data) > len(payload) {
    log.Printf("PUT data too big, was %d bytes", len(self.data))
  } else {
    copy(payload[:], self.data)
  }
  return payload
}

func (self bencDHTMessage) GET() CryptoHash {
  var hash CryptoHash
  if len(self.data) > len(hash) {
    log.Printf("GET data too big, was %d bytes", len(self.data))
  } else {
    copy(hash[:], self.data)
  }
  return hash
}

func (self bencDHTMessage) Method() string {
  return self.method
}

func (self bencDHTMessage) Hops() int64 {
  return self.hops
}

func (self bencDHTMessage) Reply() bool {
  return self.reply
}

func (self bencDHTMessage) Bytes() []byte {
  return self.data
}

func (self bencDHTMessage) CAPS() DHTCapacity {
  // TODO: implement
  var caps [8]byte
  return caps
}

func (self bencDHTMessage) ID() int64 {
  return self.tid
}
