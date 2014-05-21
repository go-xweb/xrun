package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

const (
	ForeBlack  = iota + 30 //30         40         黑色
	ForeRed                //31         41         紅色
	ForeGreen              //32         42         綠色
	ForeYellow             //33         43         黃色
	ForeBlue               //34         44         藍色
	ForePurple             //35         45         紫紅色
	ForeCyan               //36         46         青藍色
	ForeWhite              //37         47         白色
)

const (
	LevelTrace = iota + 1
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelCritical
)

type loggerFunc func(*log.Logger, int, string, ...interface{})

type loggerFuncf func(*log.Logger, int, string, ...interface{})

type Logger struct {
	*log.Logger
	log  loggerFunc
	logf loggerFuncf
}

func noColorLog(logger *log.Logger, color int, head string, params ...interface{}) {
	content := []interface{}{interface{}(head)}
	content = append(content, params...)
	logger.Println(content...)
}

func colorLog(logger *log.Logger, color int, head string, params ...interface{}) {
	content := []interface{}{interface{}(head)}
	content = append(content, params...)
	s := fmt.Sprintln(content...)
	logger.Printf("\033[%v;1m%s\033[0m", color, s)
}

func noColorLogf(logger *log.Logger, color int, format string, params ...interface{}) {
	logger.Printf(format, params...)
}

func colorLogf(logger *log.Logger, color int, format string, params ...interface{}) {
	s := fmt.Sprintf(format, params...)
	fmt.Println("====", s)
	logger.Printf("\033[%v;1m%s\033[0m", color, s)
}

var (
	DefaultLogger *Logger
)

func init() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	if runtime.GOOS == "windows" {
		DefaultLogger = &Logger{logger, noColorLog, noColorLogf}
	} else {
		DefaultLogger = &Logger{logger, colorLog, colorLogf}
	}
}

func (l *Logger) Error(params ...interface{}) {
	l.log(l.Logger, ForeRed, "[Error]", params...)
}

func (l *Logger) Errorf(format string, params ...interface{}) {
	l.logf(l.Logger, ForeRed, "[Error] "+format, params...)
}

func (l *Logger) Info(params ...interface{}) {
	l.log(l.Logger, ForeGreen, "[Info]", params...)
}

func (l *Logger) Infof(format string, params ...interface{}) {
	l.logf(l.Logger, ForeGreen, "[Info] "+format, params...)
}

func (l *Logger) Debug(params ...interface{}) {
	l.log(l.Logger, ForeBlue, "[Debug]", params...)
}

func (l *Logger) Debugf(format string, params ...interface{}) {
	l.logf(l.Logger, ForeBlue, "[Debug] "+format, params...)
}

func (l *Logger) Trace(params ...interface{}) {
	l.log(l.Logger, ForeCyan, "[Trace]", params...)
}

func (l *Logger) Tracef(format string, params ...interface{}) {
	l.logf(l.Logger, ForeCyan, "[Trace] "+format, params...)
}

func (l *Logger) Warn(params ...interface{}) {
	l.log(l.Logger, ForeYellow, "[Warn]", params...)
}

func (l *Logger) Warnf(format string, params ...interface{}) {
	l.logf(l.Logger, ForeYellow, "[Warn] "+format, params...)
}

func (l *Logger) Critical(params ...interface{}) {
	l.log(l.Logger, ForePurple, "[Critical]", params...)
}

func (l *Logger) Criticalf(format string, params ...interface{}) {
	l.logf(l.Logger, ForePurple, "[Critical] "+format, params...)
}

func Error(params ...interface{}) {
	DefaultLogger.Error(params...)
}

func Errorf(format string, params ...interface{}) {
	DefaultLogger.Errorf(format, params...)
}

func Info(params ...interface{}) {
	DefaultLogger.Info(params...)
}

func Infof(format string, params ...interface{}) {
	DefaultLogger.Infof(format, params...)
}

func Debug(params ...interface{}) {
	DefaultLogger.Debug(params...)
}

func Debugf(format string, params ...interface{}) {
	DefaultLogger.Debugf(format, params...)
}

func Trace(params ...interface{}) {
	DefaultLogger.Trace(params...)
}

func Tracef(format string, params ...interface{}) {
	DefaultLogger.Tracef(format, params...)
}

func Critical(params ...interface{}) {
	DefaultLogger.Critical(params...)
}

func Criticalf(format string, params ...interface{}) {
	DefaultLogger.Criticalf(format, params...)
}

func Warn(params ...interface{}) {
	DefaultLogger.Warn(params...)
}

func Warnf(format string, params ...interface{}) {
	DefaultLogger.Tracef(format, params...)
}
