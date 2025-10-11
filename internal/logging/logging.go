package logging

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetLogger(logLevel string) *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""
	encoderConfig.CallerKey = ""

	logger, _ := zap.Config{
		Level: zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()
	return logger
}

func LogReceive(filename string, size int64, remoteAddr string) {
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    fmt.Printf("[%s] Upload: %s (%s) from %s\n", 
        timestamp, 
        filename, 
        formatBytes(size), 
        remoteAddr)
}

func LogServe(filename string, size int64, remoteAddr string) {
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    fmt.Printf("[%s] File Download: %s (%s) from %s\n", 
        timestamp, 
        filename, 
        formatBytes(size), 
        remoteAddr)
}

func formatBytes(bytes int64) string {
    if bytes < 1024 {
        return fmt.Sprintf("%d bytes", bytes)
    } else if bytes < 1024*1024 {
        return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
    } else {
        return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
    }
}