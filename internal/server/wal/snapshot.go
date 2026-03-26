package wal

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type SnapshotData struct {
	Data      map[string]string
	SnapIndex uint32
}

type Snapshot struct {
	mu sync.Mutex

	Path      string
	SnapIndex uint32
}

func NewSnapshot(path string) *Snapshot {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &Snapshot{Path: path, SnapIndex: 0}
}

func (snapshot *Snapshot) Save(data map[string]string, lastIndex uint32) error {

	snapshot.mu.Lock()
	defer snapshot.mu.Unlock()

	tmpName := filepath.Join(snapshot.Path, "snapshot.tmp")
	fileName := filepath.Join(snapshot.Path, "snapshot.dat")

	// Create a temp file to prevent corruption
	file, err := os.OpenFile(tmpName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	// Encode data in GOB format
	encoder := gob.NewEncoder(file)

	snapShot := SnapshotData{
		Data:      data,
		SnapIndex: lastIndex,
	}
	if err := encoder.Encode(snapShot); err != nil {
		_ = file.Close()
		return err
	}

	// Fsync the file for persistence
	err = file.Sync()
	if err != nil {
		_ = file.Close()
		return err
	}
	_ = file.Close()

	// Rename the file to .dat
	err = os.Rename(tmpName, fileName)
	if err != nil {
		return err
	}

	dir, err := os.Open(snapshot.Path)
	if err == nil {
		dir.Sync()
		dir.Close()
	}

	return nil
}

func (snapshot *Snapshot) Read() (*SnapshotData, error) {

	snapshot.mu.Lock()
	defer snapshot.mu.Unlock()

	fileName := filepath.Join(snapshot.Path, "snapshot.dat")

	// Open the snapshot with ReadOnly mode
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return &SnapshotData{
				Data:      make(map[string]string),
				SnapIndex: 0,
			}, nil
		}
		return nil, err
	}
	defer file.Close()

	// Decode the snapshot with GOB format.
	decoder := gob.NewDecoder(file)
	snapShot := SnapshotData{}
	if err := decoder.Decode(&snapShot); err != nil {
		return nil, err
	}

	return &snapShot, nil
}
