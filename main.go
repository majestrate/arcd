package main

import (
  "github.com/majestrate/arcd/arc"
  "time"
)

func main() {
  counter := 15
  for counter > 0 {
    go func () {
      var router arcd.RouterMain
      router = new(arcd.TestnetRouterMain)
      config := arcd.LoadConfig("arcd.ini")
      router = router.Configure(config)
      router.Run()
    }()
    counter --
  }
  for {
    time.Sleep(time.Second)
  }
}
