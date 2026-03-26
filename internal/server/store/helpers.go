package store

import (
	"fmt"

	"github.com/vijayvenkatj/kv-store/internal/server/wal"
)

/*
applyToMemory applies a log entry to the in-memory data. Decouples the in-memory entry logic from Apply
*/
func (s *Store) applyToMemory(entry *wal.LogEntry) error {
	switch entry.Operation {
	case "put":
		s.data[entry.Key] = entry.Value
	case "delete":
		delete(s.data, entry.Key)
	default:
		return IllegalOperationErr
	}
	return nil
}

/*
ApplyLoop continuously applies log entries from the WAL to the in-memory data until it reaches the commit index.
*/
func (s *Store) ApplyLoop() {
	for {
		s.mu.Lock()

		for s.LastApplied >= s.CommitIndex {
			s.cond.Wait()
		}

		for s.LastApplied < s.CommitIndex {
			nextIndex := s.LastApplied + 1

			entry, err := s.wal.Get(nextIndex)
			if err != nil {
				fmt.Printf("Error getting entry at index %d: %v\n", nextIndex, err)
				break
			}

			if err := s.applyToMemory(entry); err != nil {
				fmt.Printf("Error applying entry at index %d: %v\n", nextIndex, err)
				break
			}

			s.LastApplied = nextIndex
		}

		s.mu.Unlock()
	}
}

/*
ApplyTill applies log entries from the WAL to the in-memory data until it reaches the specified commit index.
*/
func (s *Store) ApplyTill(commitIdx uint32) error {
	entries, err := s.wal.ReadSince(s.LastApplied)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.LogIndex > commitIdx {
			break
		}
		if err := s.applyToMemory(entry); err != nil {
			return err
		}
	}

	return nil
}

/*
storeToSnapshot creates a snapshot of the current state of the store and saves it to disk. It also truncates the WAL to free up space.
The snapshot includes the current data and the last applied index, which can be used for recovery in case of a crash.
*/
func (s *Store) storeToSnapshot() error {

	// Copy the data
	oldData := make(map[string]string)
	for k, v := range s.data {
		oldData[k] = v
	}
	lastAppliedIndex := s.LastApplied

	// Snapshot
	err := s.snap.Save(oldData, lastAppliedIndex)
	if err != nil {
		fmt.Println("Error writing snapshot:", err)
		return err
	}

	// Truncate WAL
	err = s.wal.Reset()
	if err != nil {
		fmt.Println("Error resetting wal:", err)
	}
	s.snap.SnapIndex = lastAppliedIndex

	return nil
}
