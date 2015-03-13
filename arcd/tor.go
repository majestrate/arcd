//
// tor control 
//
package arcd

import (
  "fmt"
  "io/ioutil"
  "log"
  "os/exec"
)

type TorProc struct {
  Cmd *exec.Cmd
  torrcPath string
  socksPort int
}

func SpawnTor(port int) *TorProc {
  proc := new(TorProc)
  proc.torrcPath = "contrib/torrc"
  proc.socksPort = port
  proc.Cmd = exec.Command("tor", "-f", proc.torrcPath)
  return proc
}

func (self *TorProc) SocksAddr() string {
  return fmt.Sprintf("127.0.0.1:%d",self.socksPort)
}

func (self *TorProc) Start() {
  log.Println("starting tor...")
  self.Cmd.Start()
}

func (self *TorProc) Wait() {
  self.Cmd.Wait()
  log.Println("tor exited")
}

func (self *TorProc) Stop() {
  log.Println("stopping tor")
  self.Cmd.Process.Kill()
  self.Wait()
}


func (self *TorProc) GetOnion() string {
  data, err := ioutil.ReadFile("tordata/arcd/hostname")
  if err != nil {
    return ""
  }
  return string(data)
}