package manager

import (
	"context"
	"fmt"
	"os"
	"sync"

	"drassi.run/core/pkg/executor"
	"drassi.run/gha-runner/pkg/reporter/log"
)

type EventKind uint16

const (
	OnNewStep EventKind = iota
	OnCompleteStep
	OnLogRecord
)

type Event struct {
	Kind    EventKind
	StepRun executor.StepRun
	Data    *log.Update
}

type LogManager struct {
	basePath string
	stepRun  executor.StepRun
	idx      int

	mu       sync.Mutex
	maxSize  int64
	currSize int64
	currLine int64

	file *os.File
	subs []chan *Event
}

func (lm *LogManager) Subscribe() <-chan *Event {
	ch := make(chan *Event, 5)

	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.subs = append(lm.subs, ch)
	return ch
}

func (lm *LogManager) Handle(_ context.Context, line string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// newFile if needed
	if lm.file == nil {
		if err := lm.newFile(lm.currFile()); err != nil {
			return err
		}
	}

	// write data
	if err := lm.write(line); err != nil {
		return err
	}
	u := &log.Update{
		File:   lm.currFile(),
		Status: log.FileOpen,
		Line:   lm.currLine,
		Size:   lm.currSize,
	}

	// rotate
	if lm.currSize >= lm.maxSize {
		u.Status = log.FileClose
		if err := lm.rotate(); err != nil {
			return err
		}
	}

	// Notify subscribers
	lm.notify(u)
	return nil
}

func (lm *LogManager) write(line string) error {
	if l := len(line); l == 0 || line[l-1] != '\n' {
		line += "\n"
	}

	if n, err := lm.file.WriteString(line); err != nil {
		return err
	} else {
		lm.currSize += int64(n)
		lm.currLine++
		return nil
	}
}

func (lm *LogManager) currFile() string {
	return fmt.Sprintf("%s/%s.%d.log", lm.basePath, lm.stepRun.StepId(), lm.idx)
}

func (lm *LogManager) newFile(path string) error {
	if f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); err != nil {
		return err
	} else {
		lm.file = f
		lm.currSize, lm.currLine = 0, 0
		return nil
	}
}

func (lm *LogManager) notify(u *log.Update) {
	e := &Event{
		Kind:    OnLogRecord,
		StepRun: lm.stepRun,
		Data:    u,
	}

	for _, sub := range lm.subs {
		sub <- e
	}
}

func (lm *LogManager) rotate() error {
	// Chmod to RO
	if err := lm.file.Chmod(0400); err != nil {
		return err
	}
	if err := lm.file.Close(); err != nil {
		return err
	}

	lm.file, lm.currSize, lm.currLine = nil, 0, 0
	lm.idx++
	return nil
}

func (lm *LogManager) Close() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.file != nil {
		u := &log.Update{
			File:   lm.currFile(),
			Status: log.FileClose,
			Line:   lm.currLine,
			Size:   lm.currSize,
		}

		if err := lm.rotate(); err != nil {
			return err
		}

		lm.notify(u) // Notify subscribers
	}

	for _, sub := range lm.subs {
		close(sub)
	}
	return nil
}
