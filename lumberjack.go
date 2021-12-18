package console

import "gopkg.in/natefinch/lumberjack.v2"

var stream *lumberjack.Logger

func NewLumberjack(filename string, filesizeMB, backups int) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    filesizeMB,
		MaxBackups: backups,
		LocalTime:  true,
		Compress:   true,
	}
}
