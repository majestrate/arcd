//
// configuration parser
//
package arcd

import (
  "github.com/majestrate/configparser"
  "log"
)


func LoadConfig(fname string) map[string]string {
  conf, err := configparser.Read(fname)
  if err != nil {
    log.Fatal("error loading config: ", err)
  }
  sect, err := conf.Section("arcd")
  if err != nil {
    log.Fatal("error parsing config: ", err)
  }
  return sect.Options()
}
