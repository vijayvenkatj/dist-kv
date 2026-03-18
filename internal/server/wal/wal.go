package wal

import (
	"io"
	"os"
	"sync"
)

type LogEntry struct {
	Operation string
	Key       string
	Value     string
}

type WAL struct {
	mu   sync.Mutex
	file *os.File
}

func NewWAL(filePath string) (*WAL, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{
		file: file,
	}, nil
}

func Close(wal *WAL) error {
	wal.mu.Lock()
	defer wal.mu.Unlock()

	if wal.file != nil {
		return wal.file.Close()
	}

	return nil
}

func (wal *WAL) Append(entry *LogEntry) error {
	wal.mu.Lock()
	defer wal.mu.Unlock()

	data := encode(entry)

	n, err := wal.file.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}

	if err := wal.file.Sync(); err != nil {
		return err
	}

	return nil
}

func (wal *WAL) Restore() (*LogEntry, error) {
	wal.mu.Lock()
	defer wal.mu.Unlock()

	if wal.file == nil {
		return nil, nil
	}

}
