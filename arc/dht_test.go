//
// dht_test.go -- tests for dht
//
package arc

import (
  "testing"
)

func TestDistance(t *testing.T) {
  t.Parallel()
  var k1,k2,k3 DHTKey

  dmax := maxDist()
  
  d1 := k1.Distance(k2)
  
  if ! d1.Zero() {
    t.Log("d1 is non zero")
    t.Fail()
  }

  if d1.Equal(dmax) {
    t.Log("d1 == dmax")
    t.Fail()
  }

  if d1.LessThan(dmax) {
    // this is okay
  } else {
    t.Log("d1 < dmax failed")
    t.Fail()
  }
  
  k2[0] = 0xff

  d1 = k1.Distance(k2)

  if d1.Equal(dmax) {
    t.Log("d1 == dmax")
    t.Fail()
  }

  if d1.LessThan(dmax) {
    // this is okay
  } else {
    t.Log("d1 < dmax failed")
    t.Fail()
  }

  k3[1] = 0xff

  d2 := k3.Distance(k1)

  if d1.LessThan(d2) {
    t.Log("d2 < d1 failed")
    t.Logf("d1 = %s d2 = %s", dumpBuffer(d1), dumpBuffer(d2))
    t.Fail()
  }

  if d1.Equal(d2) {
    t.Logf("d1 = %s d2 = %s", dumpBuffer(d1), dumpBuffer(d2))
    t.Log("d1 == d2 failed")
    t.Fail()    
  }
  
}

func TestDHTGet(t *testing.T) {
  
}
