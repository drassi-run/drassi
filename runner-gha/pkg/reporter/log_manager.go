package reporter

import (
	"context"
	"os"
	"sync"

	"drassi.run/gha-runner/pkg/reporter/service"
)

type LogManager struct {
	mu       sync.Mutex
	idx      int
	maxSize  int64
	currSize int64
	currLine int64

	file *os.File
	subs []service.Subscriber
}

func (lm *LogManager) Handle(ctx context.Context, line string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// newFile if needed
	if lm.file == nil {
		if err := lm.newFile(""); err != nil {
			return err
		}
	}

	// write data
	if err := lm.write(line); err != nil {
		return err
	}

	// rotate
	if lm.currSize >= lm.maxSize {
		if err := lm.rotate(); err != nil {
			return err
		}
	}

	// Notify subscribers
	for _, sub := range lm.subs {
		sub.OnLogRecord() // TODO
	}
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

func (lm *LogManager) newFile(path string) error {
	if f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); err != nil {
		return err
	} else {
		lm.file = f
		lm.currSize, lm.currLine = 0, 0
		return nil
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
	return nil
}

func (lm *LogManager) Close() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if f := lm.file; f != nil {
		return lm.rotate()
	}
	return nil
}
