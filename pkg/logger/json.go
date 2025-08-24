package logger

import (
	"time"
)

// appendJSON formats a log entry in JSON format and appends it to the buffer.
// It creates a JSON object with timestamp, level, message, and any additional fields.
// This method is optimized for minimal allocations using buffer operations.
func (l *Logger) appendJSON(buf []byte, level Level, msg string, fields ...Field) []byte {
	buf = append(buf, '{')

	buf = append(buf, `"timestamp":"`...)
	buf = append(buf, time.Now().Format(time.RFC3339Nano)...)
	buf = append(buf, '"')

	buf = append(buf, `,"level":"`...)
	buf = append(buf, level.String()...)
	buf = append(buf, '"')

	buf = append(buf, `,"message":"`...)
	buf = appendJSONString(buf, msg)
	buf = append(buf, '"')

	for _, field := range fields {
		buf = append(buf, ',')
		buf = append(buf, '"')
		buf = appendJSONString(buf, field.Key)
		buf = append(buf, '"', ':')
		buf = appendJSONValue(buf, field.Value)
	}

	buf = append(buf, '}')
	return buf
}

// appendJSONString escapes and appends a string value to the JSON buffer.
// It handles JSON string escaping for quotes, backslashes, and control characters.
// This function is optimized for performance with minimal allocations.
func appendJSONString(buf []byte, s string) []byte {
	for _, r := range []byte(s) {
		switch r {
		case '"':
			buf = append(buf, '\\', '"')
		case '\\':
			buf = append(buf, '\\', '\\')
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		default:
			buf = append(buf, r)
		}
	}
	return buf
}

// appendJSONValue appends a typed value to the JSON buffer with proper JSON formatting.
// It supports string, int, int64, float64, and bool types. Unknown types are
// represented as the string "unknown".
func appendJSONValue(buf []byte, value interface{}) []byte {
	switch v := value.(type) {
	case string:
		buf = append(buf, '"')
		buf = appendJSONString(buf, v)
		buf = append(buf, '"')
	case int:
		buf = appendInt(buf, int64(v))
	case int64:
		buf = appendInt(buf, v)
	case float64:
		buf = appendJSONFloat(buf, v)
	case bool:
		if v {
			buf = append(buf, "true"...)
		} else {
			buf = append(buf, "false"...)
		}
	default:
		buf = append(buf, '"')
		buf = appendJSONString(buf, "unknown")
		buf = append(buf, '"')
	}
	return buf
}

// appendJSONFloat appends a float64 value to the JSON buffer.
// It provides basic float formatting with 3 decimal places precision for the
// fractional part. This is optimized for performance over full precision.
func appendJSONFloat(buf []byte, f float64) []byte {
	if f == 0.0 {
		return append(buf, '0')
	}

	if f < 0 {
		buf = append(buf, '-')
		f = -f
	}

	integer := int64(f)
	fractional := f - float64(integer)

	buf = appendInt(buf, integer)

	if fractional > 0 {
		buf = append(buf, '.')
		fractional *= 1000
		buf = appendInt(buf, int64(fractional))
	}

	return buf
}
