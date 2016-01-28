package netlink

import (
	"encoding/binary"
	"github.com/hkwi/nlgo"
	"unsafe"
	"syscall"
	"fmt"
)

type MSG struct {
	Len  uint16
	Type uint16
	Pid  uint32
}

type Attr struct {
	Type int
	Data []byte
}

func GetTaskStats(nlsk *nlgo.NlSock, p int) (t *Taskstats) {
	nlsk.Flags |= nlgo.NL_NO_AUTO_ACK
	const familyID  =  22
	m := &MSG{
		Len:  8,
		Type: TASKSTATS_CMD_ATTR_PID,
		Pid:  uint32(p),
	}

	hdr := (*[nlgo.SizeofGenlMsghdr]byte)(unsafe.Pointer(&nlgo.GenlMsghdr{
		Cmd:     TASKSTATS_CMD_GET,
		Version: 0,
	}))[:]
	req := ((*[8]byte)(unsafe.Pointer(m)))[:]
	length := 4
	pad := ((length + 4 - 1) & (^3)) - length
	for i := 0; i < pad; i++ {
		req = append(req, 0)
	}
	hdr = append(hdr, req...)
	nlgo.NlSendSimple(nlsk, familyID, 1, hdr[:])

	func() error {
		for {
			buf := make([]byte, 16384)
			if nn, _, err := syscall.Recvfrom(nlsk.Fd, buf, 0); err != nil {
				return err
			} else if nn > len(buf) {
				return nlgo.NLE_MSG_TRUNC
			} else {
				buf = buf[:nn]
			}
			if msgs, err := syscall.ParseNetlinkMessage(buf); err != nil {
				return err
			} else {
				for _, msg := range msgs {

					switch msg.Header.Type {
					case syscall.NLMSG_DONE:
						return nil
					case syscall.NLMSG_ERROR:
						return fmt.Errorf("NlMsgerr=%s", nlgo.NlMsgerr(msg))
					case 22:
						genl := (*nlgo.GenlMsghdr)(unsafe.Pointer(&msg.Data[0]))
						_ = genl
						attrs := parse_attributes(msg.Data[nlgo.GENL_HDRLEN:])
						for _, attr := range attrs {
							if attr.Type == TASKSTATS_TYPE_AGGR_PID {
								attrs = parse_attributes(attr.Data)
								break
							}
						}
						for _, attr := range attrs {
							if attr.Type == TASKSTATS_TYPE_STATS {
								_ = uint32(*(*uint32)(unsafe.Pointer(&attr.Data[248])))
								t = (*Taskstats)(unsafe.Pointer(&attr.Data[0]))
								break
							}
						}
						return nil
					default:
						return fmt.Errorf("unexpected NlMsghdr=%s", msg.Header)
					}
				}
			}
		}
	}()
//	if err != nil {
//		fmt.Println(err, err.Error())
//	}
	return
}


func parse_attributes(data []byte) map[int]Attr {
	attrs := make(map[int]Attr, 0)

	for len(data) != 0 {

		attr_len := binary.LittleEndian.Uint16(data[0:2])

		attr_type := binary.LittleEndian.Uint16(data[2:4])

		attrs[int(attr_type)] = Attr{
			int(attr_type),
			data[4:attr_len],
		}

		data = data[nlgo.NLMSG_ALIGN(int(attr_len)):]
	}
	return attrs
}