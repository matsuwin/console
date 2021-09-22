package console

import (
	"context"
	"github.com/shirou/gopsutil/v3/process"
	"os"
	"time"
)

var CPUPercent float64
var over bool

func cpuMonitor() {
	ctx := context.Background()
	proc, _ := process.NewProcess(int32(os.Getpid()))
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		if over {
			return
		}
		CPUPercent, _ = proc.CPUPercentWithContext(ctx)
	}
}
