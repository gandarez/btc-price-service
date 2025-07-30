package log

import (
	"fmt"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the log entry.
type Logger struct {
	entry         *zap.Logger
	atomicLevel   zap.AtomicLevel
	currentOutput io.Writer
	verbose       bool
}

// New creates a new Logger that writes to dest.
func New(dest io.Writer) *Logger {
	atom := zap.NewAtomicLevel()

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "now"
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderCfg.MessageKey = "message"
	encoderCfg.FunctionKey = "func"

	l := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(dest),
		atom,
	),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.FatalLevel),
	)

	logger := &Logger{
		entry:         l,
		atomicLevel:   atom,
		currentOutput: dest,
	}

	return logger
}

// IsVerboseEnabled returns true if debug is enabled.
func (l *Logger) IsVerboseEnabled() bool {
	return l.verbose
}

// Output returns the current log output.
func (l *Logger) Output() io.Writer {
	return l.currentOutput
}

// SetVerbose sets log level to debug if enabled.
func (l *Logger) SetVerbose(verbose bool) {
	l.verbose = verbose

	if verbose {
		l.atomicLevel.SetLevel(zap.DebugLevel)
	} else {
		l.atomicLevel.SetLevel(zap.InfoLevel)
	}
}

// Flush flushes the log output and closes the file.
func (l *Logger) Flush() {
	if err := l.entry.Sync(); err != nil {
		l.Debugf("failed to flush log file: %s", err)
	}

	if closer, ok := l.currentOutput.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			l.Debugf("failed to close log file: %s", err)
		}
	}
}

// Log logs a message at the given level.
func (l Logger) Log(level zapcore.Level, msg string) {
	l.entry.Log(level, msg)
}

// Logf logs a message at the given level.
func (l Logger) Logf(level zapcore.Level, format string, args ...any) {
	l.entry.Log(level, fmt.Sprintf(format, args...))
}

// Debugf logs a message at level Debug.
func (l *Logger) Debugf(format string, args ...any) {
	l.entry.Log(zapcore.DebugLevel, fmt.Sprintf(format, args...))
}

// Infof logs a message at level Info.
func (l *Logger) Infof(format string, args ...any) {
	l.entry.Log(zapcore.InfoLevel, fmt.Sprintf(format, args...))
}

// Warnf logs a message at level Warn.
func (l *Logger) Warnf(format string, args ...any) {
	l.entry.Log(zapcore.WarnLevel, fmt.Sprintf(format, args...))
}

// Errorf logs a message at level Error.
func (l *Logger) Errorf(format string, args ...any) {
	l.entry.Log(zapcore.ErrorLevel, fmt.Sprintf(format, args...))
}

// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
func (l *Logger) Fatalf(format string, args ...any) {
	l.entry.Log(zapcore.FatalLevel, fmt.Sprintf(format, args...))
}

// Debugln logs a message at level Debug.
func (l *Logger) Debugln(msg string) {
	l.entry.Log(zapcore.DebugLevel, msg)
}

// Infoln logs a message at level Info.
func (l *Logger) Infoln(msg string) {
	l.entry.Log(zapcore.InfoLevel, msg)
}

// Warnln logs a message at level Warn.
func (l *Logger) Warnln(msg string) {
	l.entry.Log(zapcore.WarnLevel, msg)
}

// Errorln logs a message at level Error.
func (l *Logger) Errorln(msg string) {
	l.entry.Log(zapcore.ErrorLevel, msg)
}

// Fatalln logs a message at level Fatal then the process will exit with status set to 1.
func (l *Logger) Fatalln(msg string) {
	l.entry.Log(zapcore.FatalLevel, msg)
}

// WithFields adds fields to the Logger.
func (l *Logger) WithFields(fields ...zap.Field) {
	l.entry = l.entry.With(fields...)
}
