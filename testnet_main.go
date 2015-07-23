package main

import (
  "github.com/majestrate/arcd/arc"
  "github.com/majestrate/arcd/nacl"

  "bytes"
  "log"
  "time"
)

func main() {
  cfg := arcd.LoadConfig("arcd.ini")
  // make 20 nodes
  counter := 20
  var routers []arcd.ArcRouterMain
  for counter > 0 {
    var router arcd.ArcRouterMain
    router.Configure(cfg)
    routers = append(routers, router)
    counter --
  }
  for _, router := range(routers) {
    go router.Run()
  }
  router := routers[0]
  for {
    log.Println("go")
    // generate random test data
    payload := nacl.RandBytes(128)
    // insert the random test data
    log.Println("insert")
    rootHash := router.InsertData(payload)
    // wait for 1 second for the data to be inserted
    time.Sleep(time.Millisecond * 500)
    // Get the data
    data := router.GetDataViaHash(rootHash)
    // check it
    if bytes.Equal(data, payload) {
      // izgud :^3
      log.Println("GET SUCCESS !!!")
    } else {
      // failed
      log.Printf("GET FAIL: GOT %q instead of %q", data, payload)
    }
    
  }
}
