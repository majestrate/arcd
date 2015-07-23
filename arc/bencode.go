//
// bencode.go
// implements a small subset of bittorrent encoding, specificly a dictionary holding ints and bytestrings
//
package arcd

import (
  "bufio"
  "errors"
  "fmt"
  "io"
  "strconv"
)

// encode a map
func bencodeEncodeMap(data map[string]interface{}, writer io.Writer) (err error) {
  _, err = io.WriteString(writer, "d")
  if err == nil {
    for k, v := range(data) {
      // write key
      _, err = io.WriteString(writer, fmt.Sprintf("%d:", len(k)))
      if err != nil {
        return
      }
      _, err = writer.Write([]byte(k))
      if err != nil {
        return
      }
      // write value
      switch t := v.(type) {
        // integers
        // signed
      case int:
      case int8:
      case int16:
      case int32:
      case int64:
      case uint:
      case uint8:
      case uint16:
      case uint32:
      case uint64:
        _, err = io.WriteString(writer, fmt.Sprintf("i%le", v))
        break
        
      case string:
        _, err = io.WriteString(writer, fmt.Sprintf("%d:", len(v.(string))))
        _, err = io.WriteString(writer, v.(string))
        break
      case []byte:
        _, err = io.WriteString(writer, fmt.Sprintf("%d:", len(v.([]byte))))
        _, err = writer.Write(v.([]byte))
        break
      default:
        err = errors.New(fmt.Sprintf("cannot encode type %s", t))
      }
    }
  }
    _, err = io.WriteString(writer, "e")
  return
}

// read a string
func bencode_readString(reader *bufio.Reader) ([]byte , error) {
  // read a string
  slenb, err := reader.ReadBytes(':')
  if err == nil {
    // read string length
    slen, err := strconv.ParseUint(string(slenb[:len(slenb)-1]), 10, 32)
    if err == nil {
      key := make([]byte, slen)
      _, err = reader.Read(key)
      if err == nil {
        return key, nil
      }
    }
  }
  return nil, err
}


// decode a map  
func bencodeDecodeMap(r io.Reader) (ret map[string]interface{}, err error) {
  reader := bufio.NewReader(r)
  // assert first entry is 'd'
  var b byte
  var p, key []byte
  b, err = reader.ReadByte()
  if err == nil {
    if b == 'd' {
      // read entries
      for {
        // peek, looking for 'e'
        p, err = reader.Peek(1)
        if err != nil {
          return nil, err
        }
        // we are at the end of this dict
        if p[0] == 'e' {
          // read last 'e'
          _, _ = reader.ReadByte()
          // we gud. return
          return 
        }
        // read key
        key, err = bencode_readString(reader)
        // could not read key
        if err != nil {
          return  
        }
        // peek next value
        p, err = reader.Peek(1)
        if err != nil {
          return 
        }
        if p[0] == 'i' {
          // read the 'i' and discard
          _, _ = reader.ReadByte()
          // read our number
          p, err = reader.ReadBytes('e')
          if err == nil {
            // parse it
            var num uint64
            num, err = strconv.ParseUint(string(p[:len(p)-1]), 10, 64)
            if err == nil {
              // it's an int
              // put it
              k := string(key)
              ret[k] = num
              // next entry
              continue
            }
          }
          // return if error
          return
        } else if p[0] >= '0' && p[0] <= '9' {
          // it's a string
          // read it
          p, err = bencode_readString(reader)
          if err != nil {
            return 
          }
          // put it
          k := string(p)
          ret[k] = p
          // next entry
          continue
        }
      }
    }
    return nil, errors.New("item not a dict")
  }
  return
}
