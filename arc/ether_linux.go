// +build linux
//
// ether.go -- ethernet hub
//

package arc

/*

#include <unistd.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <linux/if_packet.h>
#include <linux/if_ether.h>
#include <linux/if_arp.h>

// open ethernet given interface name
int ether_open(int if_idx, const unsigned char * hw_addr) {
  int fd = -1;
  fd = socket(AF_PACKET, SOCK_RAW, htons(0xD1CE));
  if ( fd != -1 ) {
    struct sockaddr_ll addr;
    memset(&addr, 0, sizeof(struct sockaddr_ll));
    addr.sll_protocol = htons(0xD1CE);
    addr.sll_family = AF_PACKET;
    addr.sll_halen = ETH_ALEN;
    memcpy(&addr.sll_addr, hw_addr, addr.sll_halen);
    if ( bind(fd, (struct sockaddr*)&addr, sizeof(struct sockaddr_ll)) == -1 ) {
      close(fd);
      return -1;
    }
    return fd;
  } 
  return -1;
}

size_t ether_recv(int fd, int if_idx, void * result) {
  struct sockaddr_ll addr;
  memset(&addr, 0, sizeof(struct sockaddr_ll));
  addr.sll_hatype = ARPHRD_ETHER;
  addr.sll_pkttype = PACKET_BROADCAST;
  return recvfrom(fd, result, ETH_FRAME_LEN, 0, (struct sockaddr*)&addr, NULL);
}

// ethernet broadcast frame
int ether_broadcast(int fd, int if_idx, const unsigned char * if_hwaddr, const char * dataptr, const size_t datalen) {

  if ( datalen + 14 > ETH_FRAME_LEN ) {
    // invalid size
    return -1;
  }
  char frame[ETH_FRAME_LEN];

  char* head = frame;
  char* data = head + 14;
  struct ethhdr * eh = (struct ethhdr *) head;
  memcpy(data, dataptr, datalen);

  // broadcast address
  struct sockaddr_ll addr;
  addr.sll_family = AF_PACKET;
  addr.sll_ifindex = if_idx;
  addr.sll_halen = ETH_ALEN;
  memset(addr.sll_addr, 0xff, ETH_ALEN);

  // broadcast dest addr
  memset(eh->h_dest, 0xff, ETH_ALEN);
  // source addr is our network interface
  memcpy(eh->h_source, if_hwaddr, ETH_ALEN);
  // ethernet protocol is 0xd1ce
  eh->h_proto = htons(0xD1CE);
  int result = -1;

  result = sendto(fd, (void*)&frame, datalen + 14, 0, (struct sockaddr *)&addr, sizeof(addr));
  return result;

}

void ether_close(int fd) {
  if (fd != -1) close(fd);
}

*/
import "C"

import (
  "encoding/binary"
  "errors"
  "log"
  "net"
  "time"
)

type etherHub struct {
  fd C.int
  iface *net.Interface
  hwaddr [6]C.uchar
  send chan Message
  ib chan Message
  router Router
  filter bloomFilter
}

// bind to a network interface
func (eh *etherHub) bind(iface string) (err error) {
  eh.iface, err = net.InterfaceByName(iface)
  if err == nil {
    if len(eh.iface.HardwareAddr) == 6 {
      log.Println("binding to", eh.iface.HardwareAddr)
      for n, c := range eh.iface.HardwareAddr {
        eh.hwaddr[n] = C.uchar(c)
      }
      eh.fd = C.ether_open(C.int(eh.iface.Index), &eh.hwaddr[0])
      if eh.fd == -1 {
        err = errors.New("cannot bind to "+iface)
        return
      }
    } else {
      err = errors.New("hardware address != 6")
    }
  }
  return 
}

func (eh etherHub) Persist(_ string, _ int, _, _ string, _ int) {
  return
}

func (eh etherHub) SendChan() chan Message {
  return eh.send
}

// broadcast raw data
func (eh *etherHub) broadcast(data []byte) (err error) {
  // add to filter
  eh.filter.Add(data)
  log.Println("broadcast", len(data), eh.iface.Index, eh.hwaddr)
  d := make([]C.char, len(data))
  for i, c := range data {
    d[i] = C.char(c)
  }
  res := C.ether_broadcast(eh.fd, C.int(eh.iface.Index), &eh.hwaddr[0], &d[0], C.size_t(len(data)))
  if res == -1 {
    err = errors.New("failed to send ethernet frame")
  }
  log.Println("broadcasted", len(data))
  return
}

// run main
func (eh etherHub) Run() {
  log.Println("run ethernet hub")
  go eh.sendLoop()
  eh.recvLoop()
}

func (eh *etherHub) sendLoop() {
  for {
    select {
    case msg, ok := <- eh.ib:
      if ok {
        eh.router.InboundChan() <- msg
      } else {
        return
      }
    case msg, ok := <- eh.send:
      log.Println("got send")
      if ok {
        // broadcast minus first 2 bytes
        err := eh.broadcast(msg.RawBytes()[2:])
        if err == nil {
          // we gud
          continue
        } else {
          log.Println("failed to broadcast over ethernet", err)
        }
      }
    }
  }
}

// run recv loop from ethernet
func (eh *etherHub) recvLoop() {
  // big buffer
  ptr := C.malloc(1518)
  buff := C.GoBytes(ptr, 1518)
  if buff == nil {
    log.Println("failed to malloc")
    return
  }
  defer C.free(ptr)
  for {
    // low level recv
    idx := C.int(eh.iface.Index)
    rsize := C.ether_recv(eh.fd, idx, ptr)
    recv_size := int(rsize)
    if recv_size >= 38 && recv_size <= 1518 {
      log.Println("eth recv", recv_size)
      var msg urcMessage
      msg.body = make([]byte, recv_size - 38)
      // exclude ethernet header
      copy(msg.hdr[:2], buff[14:38])
      // put urc header length
      binary.BigEndian.PutUint16(msg.hdr[:2], uint16(len(msg.body)))
      // put urc body
      copy(msg.body, buff[38:])
      r := msg.RawBytes()
      if eh.filter.Contains(r) {
        // filter hit
        continue
      } else {
        // filter pass
        eh.filter.Add(r)
      }
      eh.ib <- msg
    } else {
      log.Println("invalid ether_recv size:", recv_size)
      time.Sleep(time.Second)
    }
  }
}

func (eh etherHub) Close() {
  C.ether_close(eh.fd)
}

func CreateEthernetHub(ifname string, r Router) Hub {
  log.Println("create ethernet hub")
  h := etherHub{
    send: make(chan Message),
    router: r,
  }
  err := h.bind(ifname)
  if err == nil {
    return h
  }
  log.Fatal("failed to create ethernet hub: ", err.Error())
  return nil
}
