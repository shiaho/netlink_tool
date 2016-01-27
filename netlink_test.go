package main
import "testing"


import (
	"github.com/hkwi/nlgo"
	"io/ioutil"
	"strconv"
	"syscall"
	"github.com/shiaho/netlink_tool/netlink"
	"github.com/kr/pretty"
	"fmt"
	"strings"
)

var (
	_ = pretty.Sprintf("")
	taskstatsMap = make(map[int]*netlink.Taskstats)
	state = make(map[string]int64)
	pids []int
)


func init() {
	files, _ := ioutil.ReadDir("/proc/")
	for _, f := range files {
		if f.Name()[0] >= '0' && f.Name()[0] <= '9' {
			pid, _ := strconv.ParseInt(f.Name(), 10, 32)
			pids = append(pids, int(pid))
		}
	}
}

func BenchmarkNetlink(b *testing.B) {
	nlsk := nlgo.NlSocketAlloc()
	nlgo.NlConnect(nlsk, syscall.NETLINK_GENERIC)
	for i := 0; i < b.N; i++ {
		for _, p := range pids {
			t := netlink.GetTaskStats(nlsk, p)
			if t != nil {
				if taskstatsMap[p] == nil {
					taskstatsMap[p] = t
				} else {
					taskstatsMap[p] = t
				}

			}
		}
	}
}



func BenchmarkProc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, p := range pids {
			data, _ := ioutil.ReadFile(fmt.Sprintf("/proc/%d/io", p))

			for _, line := range strings.Split(string(data), "\n") {
				if line == "" {
					continue
				}
				d := strings.Split(line, ": ")
				v, _ := strconv.ParseInt(d[1],10, 32)
				state[d[0]] = v
			}
		}
	}
}

func BenchmarkAllPid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		files, _ := ioutil.ReadDir("/proc/")
		var pids []int
		for _, f := range files {
			if f.Name()[0] >= '0' && f.Name()[0] <= '9' {
				pid, _ := strconv.ParseInt(f.Name(), 10, 32)
				pids = append(pids, int(pid))
			}

		}
	}
}

/*
BenchmarkNetlink     500           2503186 ns/op
BenchmarkProc        500           3554837 ns/op
BenchmarkAllPid     2000            878981 ns/op
 */