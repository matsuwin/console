package console

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	"os"
	"runtime"
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

func loop() {
	defer fos.Close()

	if control.Filename == "" {
		control.Filename = "log/execution.log"
	}

	fos = NewLumberjack(control.Filename, control.LogFileSizeMB, control.MaxBackups)
	writeLine := ""

	for elem := range engine {
		if elem.M == consoleExit {
			over = true
			return
		}

		date := elem.T.Local().Format(time.RFC3339)
		function := runtime.FuncForPC(elem.F).Name()

		switch elem.L {
		case _INFO:
			writeLine = fmt.Sprintf("INFO   %s (%s) %s (CPU:%.1f%%) %s\n", date, function, elem.N, elem.cpu, elem.M)
		case _DEBUG:
			writeLine = fmt.Sprintf("DEBUG  %s (%s) %s (CPU:%.1f%%) %s\n", date, function, elem.N, elem.cpu, elem.M)
		case _WARN:
			writeLine = fmt.Sprintf("WARN   %s (%s) %s (CPU:%.1f%%) %s\n", date, function, elem.N, elem.cpu, elem.M)
		case _ERROR:
			writeLine = fmt.Sprintf("ERROR  %s (%s) %s (CPU:%.1f%%) %s\n", date, function, elem.N, elem.cpu, elem.M)
		}
		_, _ = fos.Write([]byte(writeLine))

		if control.Print {
			// fmt.Printf("\033[0;31;48m%s\033[0m\n", "RED")
			//
			// 前景 背景 颜色
			// ---------------------------------------
			// 30  40  黑色
			// 31  41  红色
			// 32  42  绿色
			// 33  43  黄色
			// 34  44  蓝色
			// 35  45  紫红色
			// 36  46  青蓝色
			// 37  47  白色
			//
			// 代码 意义
			// -------------------------
			//  0  终端默认设置
			//  1  高亮显示
			//  4  使用下划线
			//  5  闪烁
			//  7  反白显示
			//  8  不可见
			switch elem.L {
			case _INFO:
				fmt.Printf("INFO  %s (%s) %s (CPU:%.1f%%) %s\n", date, function, elem.N, elem.cpu, elem.M)
			case _DEBUG:
				fmt.Printf("\u001B[0;34;48mDEBUG %s (%s) %s (CPU:%.1f%%) %s\u001B[0m\n", date, function, elem.N, elem.cpu, elem.M)
			case _WARN:
				fmt.Printf("\u001B[0;33;48mWARN  %s (%s) %s (CPU:%.1f%%) %s\u001B[0m\n", date, function, elem.N, elem.cpu, elem.M)
			case _ERROR:
				fmt.Printf("\u001B[0;31;48mERROR %s (%s) %s (CPU:%.1f%%) %s\u001B[0m\n", date, function, elem.N, elem.cpu, elem.M)
			}
		}
	}
}