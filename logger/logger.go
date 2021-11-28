// Package logger contains an abstraction of logger methods and adapter for zap logger.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type vType uint8

const (
	tAny vType = iota
	tError
	tString
	tUint
	tInt
)

// Field is a general type to log typed values as log message attributes.
type Field struct {
	key   string
	vType vType
	value interface{}
}

// Any creates field with an attribute value of any type.
func Any(key string, value interface{}) Field {
	return Field{
		key:   key,
		vType: tAny,
		value: value,
	}
}

// Error creates field with an attribute of error type.
func Error(value error) Field {
	return Field{ //nolint:exhaustivestruct
		vType: tError,
		value: value,
	}
}

// String creates field with an attribute of string type.
func String(key string, value string) Field {
	return Field{
		key:   key,
		vType: tString,
		value: value,
	}
}

// Uint creates field with an attribute of uint type.
func Uint(key string, value uint) Field {
	return Field{
		key:   key,
		vType: tUint,
		value: value,
	}
}

// Int creates field with an attribute of int type.
func Int(key string, value int) Field {
	return Field{
		key:   key,
		vType: tInt,
		value: value,
	}
}

// Logger is an abstraction of typical logger methods.
type Logger interface {
	Debug(string, ...Field)
	Info(string, ...Field)
	Warn(string, ...Field)
	Error(string, ...Field)
}

var _ Logger = (*ZapAdapter)(nil)

// ZapAdapter implements Logger interface for zap logger.
type ZapAdapter struct {
	l *zap.Logger
}

// NewZapAdapter creates Logger adapter for zap logger.
func NewZapAdapter(l *zap.Logger) *ZapAdapter {
	return &ZapAdapter{l: l}
}

// Debug sends a message to logger at debug level.
func (l *ZapAdapter) Debug(message string, fields ...Field) {
	l.l.Debug(message, l.toZapFields(fields)...)
}

// Info sends a message to logger at info level.
func (l *ZapAdapter) Info(message string, fields ...Field) {
	l.l.Info(message, l.toZapFields(fields)...)
}

// Warn sends a message to logger at warn level.
func (l *ZapAdapter) Warn(message string, fields ...Field) {
	l.l.Warn(message, l.toZapFields(fields)...)
}

// Error sends a message to logger at error level.
func (l *ZapAdapter) Error(message string, fields ...Field) {
	l.l.Error(message, l.toZapFields(fields)...)
}

func (l *ZapAdapter) toZapFields(fields []Field) []zapcore.Field {
	zfs := make([]zapcore.Field, 0, len(fields))

	for _, field := range fields {
		switch field.vType {
		case tAny:
			zfs = append(zfs, zap.Any(field.key, field.value))
		case tError:
			zfs = append(zfs, zap.Error(field.value.(error)))
		case tString:
			zfs = append(zfs, zap.String(field.key, field.value.(string)))
		case tUint:
			zfs = append(zfs, zap.Uint(field.key, field.value.(uint)))
		case tInt:
			zfs = append(zfs, zap.Int(field.key, field.value.(int)))
		}
	}

	return zfs
}
