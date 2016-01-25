package main

import (
	"fmt"
	"github.com/hkwi/nlgo"
	"os"
	"syscall"
	"unsafe"
)

const (
	CN_IDX_PROC = 1
	CN_VAL_PROC = 1

	NLMSG_NOOP    = 1
	NLMSG_ERROR   = 2
	NLMSG_DONE    = 3
	NLMSG_OVERRUN = 4

	PROC_CN_MCAST_LISTEN = 1
	PROC_CN_MCAST_IGNORE = 2

	PROC_EVENT_NONE = 0
)

var PROC_EVENT_WHAT = map[uint64]string{
	0:          "PROC_EVENT_NONE",
	1:          "PROC_EVENT_FORK",
	2:          "PROC_EVENT_EXEC",
	4:          "PROC_EVENT_UID",
	0x40:       "PROC_EVENT_GID",
	0x80:       "PROC_EVENT_SID",
	0x80000000: "PROC_EVENT_EXIT",
}

type ppp struct {
	pid uint32
}

type AddPair struct {
	Host []byte
	Port uint32
}

func main() {
	nlsk := nlgo.NlSocketAlloc()
	nlgo.NlConnect(nlsk, syscall.NETLINK_GENERIC)
	syscall.Bind(nlsk.Fd, &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Pid:    uint32(os.Getpid()),
		Pad:    CN_IDX_PROC,
	})
	data := []byte{
		40,
		0,
		0,
		0,
		3,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		77,
		26,
		0,
		0,
		1,
		0,
		0,
		0,
		1,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		4,
		0,
		0,
		0,
		1,
		0,
		0,
		0,
	}

	p := (*ppp)(unsafe.Pointer(&data[12]))
	p.pid = uint32(os.Getpid())

	fmt.Println(data)

	err := syscall.Sendto(nlsk.Fd, data, 0, &nlsk.Peer)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("send success")
	}

	func() error {
		for {
			buf := make([]byte, 1024)
			if _, _, err := syscall.Recvfrom(nlsk.Fd, buf, 0); err != nil {
				return err
			}
			fmt.Println(buf)
		}
	}()

}
