//
// config.go -- config loader
//

package arc

import (
  "encoding/json"
  "log"
  "os"
)

type RemoteHubConfig struct {
  Addr string
  Port int
  ProxyAddr string
  ProxyPort int
  ProxyType string
}

type LocalHubConfig struct {
  Bind string
  Keys string
}

type Config struct {
  Remote []RemoteHubConfig
  Local LocalHubConfig
}
 
// save to file
func (cfg Config) Save(fname string) error {
  f, err := os.Create(fname)
  if err == nil {
    enc := json.NewEncoder(f)
    err = enc.Encode(cfg)
    f.Close()
  }
  return err
}

// generate default config
func genConfig() (cfg Config) {

  // urc hub for ayb
  aybHub := RemoteHubConfig{
    Addr: "allyour4nert7pkh.onion",
    Port: 6789,
    ProxyType: "socks",
    ProxyAddr: "127.0.0.1",
    ProxyPort: 9050,
  }

  
  cfg.Remote = append(cfg.Remote, aybHub)
  cfg.Local.Bind = "[::]:6789"
  cfg.Local.Keys = "privkey.dat"

  return cfg  
}

func LoadConfig(fname string) (cfg Config) {
  if ! checkFile(fname) {
    cfg = genConfig()
    err := cfg.Save(fname)
    if err != nil {
      log.Fatal("failed to save initial config", err)
    }
  }
  f, err := os.Open(fname)
  if err == nil {
    dec := json.NewDecoder(f)
    
    err = dec.Decode(&cfg)
    if err == nil {
      return cfg
    }
  }
  log.Fatal("failed to load config", err)
  return
}

