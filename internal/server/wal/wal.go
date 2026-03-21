package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type LogEntry struct {
	LogIndex  uint32
	Operation string
	Key       string
	Value     string
}

type WAL struct {
	mu   sync.Mutex
	file *os.File
}

func NewWAL(filePath string) (*WAL, error) {

	err := os.MkdirAll(filePath, 0755)
	if err != nil {
		return nil, err
	}

	fileName := filepath.Join(filePath, "wal.log")
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{
		file: file,
	}, nil
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

func (wal *WAL) ReadAll() ([]*LogEntry, error) {
	wal.mu.Lock()
	defer wal.mu.Unlock()

	if wal.file == nil {
		return nil, nil
	}

	if _, err := wal.file.Seek(0, 0); err != nil {
		return nil, err
	}

	entries := make([]*LogEntry, 0)

	for {
		header := make([]byte, 16)

		_, err := io.ReadFull(wal.file, header)
		if err == io.EOF {
			break
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}

		opLen := binary.LittleEndian.Uint32(header[4:8])
		keyLen := binary.LittleEndian.Uint32(header[8:12])
		valLen := binary.LittleEndian.Uint32(header[12:16])

		totalBody := int(opLen + keyLen + valLen)

		body := make([]byte, totalBody)

		_, err = io.ReadFull(wal.file, body)
		if err != nil {
			fmt.Println("Read Error:", err)
			break
		}

		data := make([]byte, 16+totalBody)
		copy(data[:16], header)
		copy(data[16:], body)

		entry, err := decode(data)

		entries = append(entries, entry)
	}

	_, err := wal.file.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

/*
	HELPER FUNCTIONS
*/

// Name returns the file name of the wal.
func (wal *WAL) Name() string {
	return wal.file.Name()
}

// Reset truncates the wal file
func (wal *WAL) Reset() error {
	wal.mu.Lock()
	defer wal.mu.Unlock()

	if err := wal.file.Truncate(0); err != nil {
		return err
	}
	if _, err := wal.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := wal.file.Sync(); err != nil {
		return err
	}

	return nil
}

// Close closes the wal file.
func (wal *WAL) Close() error {
	wal.mu.Lock()
	defer wal.mu.Unlock()

	if wal.file != nil {
		return wal.file.Close()
	}

	return nil
}
