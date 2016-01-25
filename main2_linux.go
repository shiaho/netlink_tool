package main

import (
	"fmt"
	"github.com/hkwi/nlgo"
	"io/ioutil"
	"strconv"
	"syscall"
	"time"
	"github.com/kr/pretty"
	"github.com/shiaho/netlink_tool/netlink"
)

var (
	_ = pretty.Sprintf("")
	taskstatsMap = make(map[int]*netlink.Taskstats)
)

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
			t := netlink.GetTaskStats(nlsk, p)
			fmt.Printf("pid: %d ", p)
			if t != nil {
				if taskstatsMap[p] == nil {
					taskstatsMap[p] = t
				} else {
					fmt.Println(t.Read_bytes - taskstatsMap[p].Read_bytes,
						t.Write_bytes - taskstatsMap[p].Write_bytes)
					taskstatsMap[p] = t
				}

			}
		}
	}
}







