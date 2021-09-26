package console

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

const _INFO, _DEBUG, _WARN, _ERROR = 0, 1, 2, 3

var (
	control     = Options{Print: true}
	engine      = make(chan structure, 1024)
	wg          = sync.WaitGroup{}
	consoleExit = "internal:consoleExit"
)

// Options ...
type Options struct {
	Print, Info, Debug, Warning, Error bool
	LogFileSizeMB                      int
	MaxBackups                         int
	Filename                           string
}

// L T F N M
type structure struct {
	L int       // Level
	T time.Time // timestamp
	F uintptr   // Function
	N string    // FileLine
	M string    // Message

	cpu float64
}

// Wait 等待缓冲区的所有数据被消费
func (*structure) Wait() {
	engine <- structure{M: consoleExit}
	wg.Wait()
}

// New 激活并启用功能
func New(options *Options) interface{ Wait() } {
	if options != nil {
		control = *options
	}
	wg.Add(2)

	go func() {
		defer func() {
			if ei := recover(); ei != nil {
			}
			wg.Done()
		}()
		cpuMonitor()
	}()

	go func() {
		defer func() {
			if ei := recover(); ei != nil {
			}
			wg.Done()
		}()
		loop()
	}()

	return &structure{}
}

// INFO 程序执行的过程信息
func INFO(msg string, a ...interface{}) {
	if control.Info {
		pc, file, line, _ := runtime.Caller(1)
		push(_INFO, &pc, fileLine(file, line), &msg, a...)
	}
}

// DEBUG 调试信息
func DEBUG(msg string, a ...interface{}) {
	if control.Debug {
		pc, file, line, _ := runtime.Caller(1)
		push(_DEBUG, &pc, fileLine(file, line), &msg, a...)
	}
}

// WARN 警告信息
func WARN(msg string, a ...interface{}) string {
	if control.Warning {
		pc, file, line, _ := runtime.Caller(1)
		push(_WARN, &pc, fileLine(file, line), &msg, a...)
	}
	return msg
}

// ERROR 错误信息
func ERROR(err error) error {
	if control.Error {
		pc, file, line, _ := runtime.Caller(1)
		msg := err.Error()
		push(_ERROR, &pc, fileLine(file, line), &msg)
	}
	return err
}

func fileLine(file string, line int) string {
	if index := strings.LastIndex(file, "/"); index != -1 {
		file = file[index+1:]
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func push(l int, f *uintptr, n string, m *string, a ...interface{}) {
	if len(a) != 0 {
		*m = fmt.Sprintf(*m, a...)
	}
	go func() {
		defer func() {
			if ei := recover(); ei != nil {
			}
		}()
		engine <- structure{l, time.Now(), *f, n, *m, CPUPercent}
	}()
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

// ManuallyClose 手动关闭日志
func ManuallyClose() {
	engine <- structure{M: consoleExit}
	wg.Wait()
	fos.Close()
}

// Timekeeper 函数运行计时
func Timekeeper(name string) func() {
	pc := make([]uintptr, 1)
	runtime.Callers(3, pc)
	start := time.Now()
	DEBUG("timekeeper START for %s", name)
	return func() {
		DEBUG("timekeeper END for %s T:%s", name, time.Since(start))
	}
}

// Json 格式化输出JSON
func Json(data interface{}, indent string) (buf []byte) {
	if indent != "" {
		buf, _ = json.MarshalIndent(data, "", indent)
		fmt.Println(string(buf))
	} else {
		buf, _ = json.Marshal(data)
		fmt.Println(string(buf))
	}
	return
}
