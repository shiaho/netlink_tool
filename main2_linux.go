package main

import (
	"encoding/binary"
	"fmt"
	"github.com/hkwi/nlgo"
	"io/ioutil"
	"strconv"
	"syscall"
	"time"
	"unsafe"
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

var read_byte map[int]uint32 = make(map[int]uint32)

func main() {
	nlsk := nlgo.NlSocketAlloc()
	nlgo.NlConnect(nlsk, syscall.NETLINK_GENERIC)
	familyID := uint16(22)
	fmt.Println(familyID)
	files, _ := ioutil.ReadDir("/proc/")
	var pids []int
	for _, f := range files {
		if f.Name()[0] >= '0' && f.Name()[0] <= '9' {
			pid, _ := strconv.ParseInt(f.Name(), 10, 32)
			pids = append(pids, int(pid))
		}
	}
	for i := 0; i < 50; i++ {
		time.Sleep(1000 * time.Millisecond)

		for _, p := range pids {
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
									var now_read_byte uint32
									if last_read_byte, ok := read_byte[p]; ok {
										now_read_byte = uint32(*(*uint32)(unsafe.Pointer(&attr.Data[248])))
										if now_read_byte-last_read_byte > 0 {
											fmt.Println(p, "read_byte", now_read_byte-last_read_byte)
										}
									}
									read_byte[p] = now_read_byte
									break
								}
							}
							switch msg.Header.Type {
							case syscall.NLMSG_DONE:
								return nil
							case syscall.NLMSG_ERROR:
								return fmt.Errorf("NlMsgerr=%s", nlgo.NlMsgerr(msg))
							default:
								return fmt.Errorf("unexpected NlMsghdr=%s", msg.Header)
							}
						}
					}
				}
			}()

		}
	}
}

const (
	TASKSTATS_CMD_UNSPEC = iota /* Reserved */
	TASKSTATS_CMD_GET           /* user->kernel request/get-response */
	TASKSTATS_CMD_NEW           /* kernel->user event */
	__TASKSTATS_CMD_MAX
)

const (
	TASKSTATS_TYPE_UNSPEC    = iota /* Reserved */
	TASKSTATS_TYPE_PID              /* Process id */
	TASKSTATS_TYPE_TGID             /* Thread group id */
	TASKSTATS_TYPE_STATS            /* taskstats structure */
	TASKSTATS_TYPE_AGGR_PID         /* contains pid + stats */
	TASKSTATS_TYPE_AGGR_TGID        /* contains tgid + stats */
	__TASKSTATS_TYPE_MAX
)

const (
	TASKSTATS_CMD_ATTR_UNSPEC = iota
	TASKSTATS_CMD_ATTR_PID
	TASKSTATS_CMD_ATTR_TGID
	TASKSTATS_CMD_ATTR_REGISTER_CPUMASK
	TASKSTATS_CMD_ATTR_DEREGISTER_CPUMASK
	__TASKSTATS_CMD_ATTR_MAX
)

func parse_attributes(data []byte) map[int]Attr {
	attrs := make(map[int]Attr, 0)
	//fmt.Println(data)
	for len(data) != 0 {

		attr_len := binary.LittleEndian.Uint16(data[0:2])
		//fmt.Println(data[:2], attr_len, uint16(data[1])<<8)
		attr_type := binary.LittleEndian.Uint16(data[2:4])

		attrs[int(attr_type)] = Attr{
			int(attr_type),
			data[4:attr_len],
		}
		//fmt.Println(nlgo.NLMSG_ALIGN(int(attr_len)))
		data = data[nlgo.NLMSG_ALIGN(int(attr_len)):]
	}
	return attrs
}
