package reporter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"drassi.run/gha-runner/pkg/reporter/transport"
)

func NewActor(courier []transport.Courier, dir, base string, limit int) (*Actor, error) {
	conveyor := transport.NewConveyor(courier)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	a := &Actor{
		conveyor: conveyor,
		dir:      dir,
		base:     base,
		limit:    limit,
	}

	return a, nil
}

type Actor struct {
	conveyor  *transport.Conveyor
	dir, base string
	limit     int

	mu sync.Mutex

	idx, n int
	file   *os.File
}

func (a *Actor) Handle(_ context.Context, line string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	n, err := a.file.WriteString(line)
	if n > 0 {
		a.n += n
	}
	if a.n >= a.limit {
		if err2 := a.rotate(); err2 != nil {
			err = errors.Join(err, err2)
		}
	}
	return err
}

func (a *Actor) rotate() (err error) {
	if err = a.file.Close(); err != nil {
		return err
	}

	a.idx++
	file := fmt.Sprintf("%s-%d.txt", a.base, a.idx)
	file = filepath.Join(a.dir, file)

	if a.file, err = os.Create(file); err != nil {
		return err
	}

	a.n = 0
	a.conveyor.AddFile(file)
	return nil
}

func (a *Actor) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file == nil {
		return nil
	}

	if err := a.file.Close(); err != nil {
		return err
	}

	a.conveyor.Complete()
	a.file = nil
	return nil
}

func (a *Actor) Wait() {
	a.conveyor.Wait()
}
