package sharedlog

import (
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/barnowlsnest/go-logslib/v2/pkg/logger"
)

// resetSingleton resets the package-level singleton only in testing purposes, so each test gets a fresh logger
// that captures the current os.Stdout at init time.
func resetSingleton() {
	once = sync.Once{}
	sharedLogger = nil
}

// captureStdout redirects stdout to a pipe, resets the singleton, runs fn, then
// returns what was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w
	resetSingleton()

	defer func() {
		os.Stdout = origStdout
		resetSingleton()
	}()

	fn()

	require.NoError(t, w.Close())
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	_ = r.Close()
	return string(buf[:n])
}

func TestInfo(t *testing.T) {
	out := captureStdout(t, func() {
		Info("smoke test info")
	})

	assert.Contains(t, out, `"level":"INFO"`)
	assert.Contains(t, out, `"message":"smoke test info"`)
}

func TestDebug(t *testing.T) {
	out := captureStdout(t, func() {
		Debug("debug trace", logger.StringField("component", "cache"))
	})

	assert.Contains(t, out, `"level":"DEBUG"`)
	assert.Contains(t, out, `"message":"debug trace"`)
	assert.Contains(t, out, `"component":"cache"`)
}

func TestError(t *testing.T) {
	out := captureStdout(t, func() {
		Error(errors.New("something broke"), logger.IntField("code", 500))
	})

	assert.Contains(t, out, `"level":"ERROR"`)
	assert.Contains(t, out, `"message":"something broke"`)
	assert.Contains(t, out, `"code":500`)
}

func TestF(t *testing.T) {
	f := logger.StringField("key", "value")
	assert.Equal(t, "key", f.Key)
	assert.Equal(t, "value", f.Value)
}

func TestLogger(t *testing.T) {
	l := Logger()
	assert.NotNil(t, l, "Logger() should return a non-nil logger")

	// Calling Logger() again should return the same singleton.
	assert.Same(t, l, Logger())
}
