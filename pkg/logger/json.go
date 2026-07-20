package logger

import (
	"time"
)

// appendJSON formats a log entry in JSON format and appends it to the buffer.
// It creates a JSON object with timestamp, level, message, and any additional fields.
// This method is optimized for minimal allocations using buffer operations.
func (l *Logger) appendJSON(buf []byte, level Level, msg string, fields ...Field) []byte {
	buf = append(buf, '{')

	now := time.Now()
	if l.config.UseUTC {
		now = now.UTC()
	}

	buf = append(buf, `"timestamp":"`...)
	buf = append(buf, now.Format(DefaultTimeFormat)...)
	buf = append(buf, '"')

	buf = append(buf, `,"level":"`...)
	buf = append(buf, level.String()...)
	buf = append(buf, '"')

	buf = append(buf, `,"message":"`...)
	buf = appendJSONString(buf, msg)
	buf = append(buf, '"')

	for _, field := range fields {
		buf = append(buf, ',', '"')
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
// It mirrors the type coverage of appendValue so that text and JSON output agree:
// all signed and unsigned integer widths, floats, bool, string, []byte, nil,
// time.Duration (as its string form, e.g. "1.5s"), time.Time (RFC3339Nano), and
// error (as its message). Named types over those kinds are handled by
// appendJSONReflect. Anything else is rendered as the string "<unsupported>".
//
//nolint:gocyclo
func appendJSONValue(buf []byte, value any) []byte {
	switch v := value.(type) {
	case nil:
		return append(buf, "null"...)
	case string:
		return appendJSONQuoted(buf, v)
	case []byte:
		return appendJSONQuoted(buf, string(v))
	case int:
		return appendInt(buf, int64(v))
	case int64:
		return appendInt(buf, v)
	case int32:
		return appendInt(buf, int64(v))
	case int16:
		return appendInt(buf, int64(v))
	case int8:
		return appendInt(buf, int64(v))
	case uint:
		return appendUint(buf, uint64(v))
	case uint64:
		return appendUint(buf, v)
	case uint32:
		return appendUint(buf, uint64(v))
	case uint16:
		return appendUint(buf, uint64(v))
	case uint8:
		return appendUint(buf, uint64(v))
	case uintptr:
		return appendUint(buf, uint64(v))
	case bool:
		if v {
			return append(buf, "true"...)
		}
		return append(buf, "false"...)
	case float64:
		return appendFloat(buf, v, 64)
	case float32:
		return appendFloat(buf, float64(v), 32)
	case time.Duration:
		return appendJSONQuoted(buf, v.String())
	case time.Time:
		buf = append(buf, '"')
		buf = v.AppendFormat(buf, time.RFC3339Nano)
		return append(buf, '"')
	case error:
		return appendJSONQuoted(buf, v.Error())
	default:
		return appendJSONReflect(buf, value)
	}
}

// appendJSONQuoted appends s as a quoted, escaped JSON string.
func appendJSONQuoted(buf []byte, s string) []byte {
	buf = append(buf, '"')
	buf = appendJSONString(buf, s)

	return append(buf, '"')
}

// appendJSONReflect handles named types whose underlying kind is representable
// in JSON (e.g. `type Status string`), which the type switch in appendJSONValue
// cannot match directly.
func appendJSONReflect(buf []byte, value any) []byte {
	return appendReflectValue(buf, value, appendJSONQuoted, "null")
}
