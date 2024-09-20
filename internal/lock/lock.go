package lock

import (
	"github.com/alexflint/go-filemutex"
)

type Lock struct {
	mtx *filemutex.FileMutex
}

var ErrAlreadyLocked = filemutex.AlreadyLocked

func NewLock() (*Lock, error) {
	mtx, err := filemutex.New("/tmp/ukcp-srd-import.lock")
	if err != nil {
		return nil, err
	}

	err = mtx.TryLock()
	if err != nil {
		return nil, err
	}

	return &Lock{mtx: mtx}, nil
}

func (l *Lock) Unlock() error {
	return l.mtx.Close()
}
