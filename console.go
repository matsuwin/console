package console

import (
	"fmt"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"runtime"
	"strings"
	"sync"
	"time"
)

const _INFO, _DEBUG, _WARN, _ERROR = 0, 1, 2, 3

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
func ERROR(err error, msg ...string) error {
	if control.Error {
		if len(msg) != 0 {
			err = errors.Wrap(err, msg[0])
		}
		pc, file, line, _ := runtime.Caller(1)
		msg := err.Error()
		push(_ERROR, &pc, fileLine(file, line), &msg)
	}
	return err
}

var (
	control     = Options{}
	rateLimit   = 1024
	engine      = make(chan structure, rateLimit)
	wg          = &sync.WaitGroup{}
	consoleExit = "internal:consoleExit"
)

// Options ...
type Options struct {
	Info, Debug, Warning, Error, Print bool
	LogFileSizeMB                      int
	MaxBackups                         int
	Filename                           string
}

// L T F N M
type structure struct {
	L uint8     // Level
	T time.Time // timestamp
	F uintptr   // Function
	N string    // FileLine
	M string    // Message

	cpu float64
}

// Wait 等待缓冲区的所有数据被消费
func (*structure) Wait() {
	engine <- structure{M: consoleExit}

	// 注意!!! 如果您在程序中并行调用了 cgo，很有可能会产生 panic，无法恢复。
	// panic:
	//   runtime: unexpected return pc for runtime.chanrecv called from 0xc000409008
	//   fatal error: unknown caller pc

	wg.Wait()
}

// New 激活并启用功能
func New(options *Options) interface{ Wait() } {
	if options != nil {
		control = *options
	} else {
		control = Options{Info: true, Debug: true, Warning: true, Error: true, Print: true}
	}
	wg.Add(2)

	go monitorCpu(wg)
	go loop(wg)

	return &structure{}
}

// ManuallyClose 手动关闭日志
func ManuallyClose() {
	if _, has := <-engine; has {
		engine <- structure{M: consoleExit}
	}
}

func push(l uint8, f *uintptr, n string, m *string, a ...interface{}) {
	if len(a) != 0 {
		*m = fmt.Sprintf(*m, a...)
	}
	if len(engine) < rateLimit {
		engine <- structure{l, time.Now(), *f, n, *m, CPUPercent}
		return
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				println(err)
			}
		}()
		engine <- structure{l, time.Now(), *f, n, *m, CPUPercent}
	}()
}

func fileLine(file string, line int) string {
	if index := strings.LastIndex(file, "/"); index != -1 {
		file = file[index+1:]
	}
	return fmt.Sprintf("%s:%d", file, line)
}

/**
 * Utils
 */

// Time2String time.Time format to string
func Time2String(t time.Time) string {
	return t.Local().Format("2006-01-02.15:04:05")
}

// Timekeeper 函数运行计时
func Timekeeper(name string) func() {
	pc := make([]uintptr, 1)
	runtime.Callers(3, pc)
	start := time.Now()
	DEBUG("[ Timekeeper ] Start %s", name)
	return func() {
		DEBUG("[ Timekeeper ] Finish %s T:%s", name, time.Since(start))
	}
}

// Json 格式化输出JSON
func Json(val interface{}, indent string) []byte {
	jss, _ := JsonMarshalIndent(val, indent)
	fmt.Printf("%s\n", jss)
	return jss
}

// JsonMarshalIndent SetEscapeHTML=false
func JsonMarshalIndent(val interface{}, indent string) (ret []byte, err error) {
	ret, err = jsonIteratorConfig.MarshalIndent(val, "", indent)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return
}

func JsonMarshal(val interface{}) (ret []byte, err error) {
	ret, err = jsonIteratorConfig.Marshal(val)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return
}

var jsonIteratorConfig = jsoniter.ConfigFastest
