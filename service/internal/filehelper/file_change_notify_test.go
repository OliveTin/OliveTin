package filehelper

import (
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/require"
)

func TestProcessDebounceCapturesEventName(t *testing.T) {
	debounceWriteLogMutex.Lock()
	debounceWriteLog = make(map[string]*FsNotifyLogEntry)
	debounceWriteLogMutex.Unlock()

	callbackNames := make(chan string, 1)
	firstEvent := fsnotify.Event{Name: "first"}
	ctx := &watchContext{
		callback: func(filename string) {
			callbackNames <- filename
		},
		event:    &firstEvent,
		filename: t.Name(),
	}

	processDebounce(ctx)

	secondEvent := fsnotify.Event{Name: "second"}
	ctx.event = &secondEvent

	select {
	case callbackName := <-callbackNames:
		require.Equal(t, firstEvent.Name, callbackName)
	case <-time.After(time.Second):
		t.Fatal("debounced callback did not run")
	}
}
