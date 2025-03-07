package wklog

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger
var errorLogger *zap.Logger
var warnLogger *zap.Logger
var panicLogger *zap.Logger
var atom = zap.NewAtomicLevel()

var opts *Options

func Configure(op *Options) {
	atom.SetLevel(op.Level)
	opts = op

	// ====================== info ==========================
	infoWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(opts.LogDir, "info.log"),
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(newEncoderConfig()),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(infoWriter)),
		atom,
	)
	if opts.LineNum {
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))
	} else {
		logger = zap.New(core)
	}

	// ====================== error ==========================
	errorWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(opts.LogDir, "error.log"),
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	core = zapcore.NewCore(
		zapcore.NewJSONEncoder(newEncoderConfig()),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(errorWriter)),
		zap.ErrorLevel,
	)
	if opts.LineNum {
		errorLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))
	} else {
		errorLogger = zap.New(core)
	}

	// ====================== warn ==========================
	warnWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(opts.LogDir, "warn.log"),
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	core = zapcore.NewCore(
		zapcore.NewJSONEncoder(newEncoderConfig()),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(warnWriter)),
		zap.WarnLevel,
	)
	if opts.LineNum {
		warnLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))
	} else {
		warnLogger = zap.New(core)
	}

	// ====================== panic ==========================
	panicWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(opts.LogDir, "panic.log"),
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	core = zapcore.NewCore(
		zapcore.NewJSONEncoder(newEncoderConfig()),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(panicWriter)),
		zap.PanicLevel,
	)
	if opts.LineNum {
		panicLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2), zap.AddStacktrace(zapcore.PanicLevel))
	} else {
		panicLogger = zap.New(core, zap.AddStacktrace(zapcore.PanicLevel))
	}

}

func newEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "linenum",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeCaller:  zapcore.FullCallerEncoder,     // 全路径编码器
		EncodeName:    zapcore.FullNameEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		},
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	}
}

// func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
// 	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
// }

// Info Info
func Info(msg string, fields ...zap.Field) {

	if logger == nil {
		Configure(NewOptions())
	}
	logger.Info(msg, fields...)

}

// Debug Debug
func Debug(msg string, fields ...zap.Field) {

	if logger == nil {
		Configure(NewOptions())
	}
	logger.Debug(msg, fields...)

}

// Error Error
func Error(msg string, fields ...zap.Field) {


	if errorLogger == nil {
		Configure(NewOptions())
	}
	errorLogger.Error(msg, fields...)

}

func Fatal(msg string, fields ...zap.Field) {

	if panicLogger == nil {
		Configure(NewOptions())
	}
	panicLogger.Fatal(msg, fields...)
}
func Panic(msg string, fields ...zap.Field) {

	if panicLogger == nil {
		Configure(NewOptions())
	}
	panicLogger.Panic(msg, fields...)
}

// Warn Warn
func Warn(msg string, fields ...zap.Field) {

	if warnLogger == nil {
		Configure(NewOptions())
	}
	warnLogger.Warn(msg, fields...)
}

func Sync() error {
	err := panicLogger.Sync()
	if err != nil {
		fmt.Println("panicLogger sync error", err)
	}
	err = errorLogger.Sync()
	if err != nil {
		fmt.Println("errorLogger sync error", err)
	}
	err = warnLogger.Sync()
	if err != nil {
		fmt.Println("warnLogger sync error", err)
	}
	err = logger.Sync()
	if err != nil {
		fmt.Println("logger sync error", err)
	}
	return nil
}

// Log Log
type Log interface {
	Info(msg string, fields ...zap.Field)
	MessageTrace(msg string, clientMsgNo string, operationName string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
}

// WKLog TLog
type WKLog struct {
	prefix string // 日志前缀
}

// NewWKLog NewWKLog
func NewWKLog(prefix string) *WKLog {

	return &WKLog{prefix: prefix}
}

// Info Info
func (t *WKLog) Info(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("【")
	b.WriteString(t.prefix)
	b.WriteString("】")
	b.WriteString(msg)
	Info(b.String(), fields...)
}

func (t *WKLog) MessageTrace(msg string, clientMsgNo string, operationName string, fields ...zap.Field) {

	if !opts.TraceOn {
		return
	}

	var b strings.Builder
	b.WriteString("【")
	b.WriteString(t.prefix)
	b.WriteString("】")
	b.WriteString(msg)
	if len(fields) == 0 {
		Info(b.String(), zap.Int("msgTrace", 1), zap.Uint64("nodeId", opts.NodeId), zap.String("clientMsgNo", clientMsgNo), zap.String("operationName", operationName))
	} else {
		fields = append(fields, zap.Int("msgTrace", 1), zap.Uint64("nodeId", opts.NodeId), zap.String("clientMsgNo", clientMsgNo), zap.String("operationName", operationName))
		Info(b.String(), fields...)
	}

}

// Debug Debug
func (t *WKLog) Debug(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("【")
	b.WriteString(t.prefix)
	b.WriteString("】")
	b.WriteString(msg)
	Debug(b.String(), fields...)
}

// Error Error
func (t *WKLog) Error(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("【")
	b.WriteString(t.prefix)
	b.WriteString("】")
	b.WriteString(msg)
	Error(b.String(), fields...)
}

// Warn Warn
func (t *WKLog) Warn(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("【")
	b.WriteString(t.prefix)
	b.WriteString("】")
	b.WriteString(msg)
	Warn(b.String(), fields...)
}

func (t *WKLog) Fatal(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("【")
	b.WriteString(t.prefix)
	b.WriteString("】")
	b.WriteString(msg)
	Fatal(b.String(), fields...)
}
func (t *WKLog) Panic(msg string, fields ...zap.Field) {
	var b strings.Builder
	b.WriteString("【")
	b.WriteString(t.prefix)
	b.WriteString("】")
	b.WriteString(msg)
	Panic(b.String(), fields...)
}
