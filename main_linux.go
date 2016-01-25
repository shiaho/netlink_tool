package main

import (
	"encoding/binary"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"syscall"
)

type Attr struct {
	Type int
	Data []byte
}

func (a *Attr) Dump() []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint16(bs, uint16(len(a.Data)+4))
	binary.LittleEndian.PutUint16(bs[2:], uint16(a.Type))
	length := len(a.Data)
	pad := ((length + 4 - 1) & (^3)) - length
	bs = append(bs, a.Data...)
	for i := 0; i < pad; i++ {
		bs = append(bs, 0)
	}
	return bs
}

type Message struct {
	Type    int
	Flags   int
	Seq     int
	Pid     int
	Payload []byte
}

func (m *Message) Len() int {
	return len(m.Payload)
}

func (m *Message) Serialize() []byte {
	fmt.Println(m.Payload)
	return m.Payload
}

var (
	NLM_F_REQUEST = 0x1
)

func main() {
	lo, _ := netlink.LinkByName("lo")
	fmt.Println(lo)
	addr, _ := netlink.ParseAddr("127.0.0.2/8")
	fmt.Println(addr)
	netlink.AddrAdd(lo, addr)
	fmt.Println(netlink.AddrList(lo, netlink.FAMILY_ALL))

	req := nl.NewNetlinkRequest(syscall.NLMSG_MIN_TYPE, syscall.NLM_F_REQUEST)

	data := append([]byte("TASKSTATS"), 0)
	a := &Attr{
		Type: 2,
		Data: data,
	}
	Hdr := []byte{
		3,
		0,
		0,
		0,
	}
	m := &Message{
		Type:    16,
		Pid:     -1,
		Seq:     -1,
		Flags:   NLM_F_REQUEST,
		Payload: append(Hdr, a.Dump()...),
	}

	req.AddData(m)
	res, _ := req.Execute(syscall.NETLINK_GENERIC, 0)
	fmt.Println(res)
	fmt.Println(parse_attributes(res[0][4:])[1])
}

func parse_attributes(data []byte) map[int]Attr {
	attrs := make(map[int]Attr, 0)
	for len(data) != 0 {

		attr_len := binary.LittleEndian.Uint16(data[:2])
		fmt.Println(data[:2], attr_len, uint16(data[1])<<8)
		attr_type := binary.LittleEndian.Uint16(data[2:4])

		attrs[int(attr_type)] = Attr{
			int(attr_type),
			data[4:attr_len],
		}
		fmt.Println((int(attr_len+4-1) & int(^3)))
		data = data[(int(attr_len+4-1) & int(^3)):]
	}
	return attrs
}
