package store

import (
	"github.com/vijayvenkatj/kv-store/internal/server/wal"
)

type AppendEntriesRequest struct {
	Term     uint32 `json:"term"`
	LeaderId uint32 `json:"leader_id"`

	PrevLogIndex uint32 `json:"prev_log_index"`
	PrevLogTerm  uint32 `json:"prev_log_term"`

	Entries []*wal.LogEntry `json:"entries"`

	LeaderCommit uint32 `json:"leader_commit"`
}

type AppendEntriesResponse struct {
	Term    uint32 `json:"term"`
	Success bool   `json:"success"`
}

func (s *Store) AppendEntries(req AppendEntriesRequest) AppendEntriesResponse {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Term check
	if req.Term < s.CurrentTerm {
		return AppendEntriesResponse{Success: false}
	}

	s.CurrentTerm = req.Term

	// Check prev log
	if req.PrevLogIndex > 0 {
		entry, err := s.wal.Get(req.PrevLogIndex)
		if err != nil || entry.Term != req.PrevLogTerm {
			return AppendEntriesResponse{Success: false}
		}
	}

	// Conflict resolution + append
	for i, newEntry := range req.Entries {
		idx := req.PrevLogIndex + 1 + uint32(i)

		existing, err := s.wal.Get(idx)

		if err == nil {
			if existing.Term != newEntry.Term {
				if err := s.wal.TruncateFrom(idx); err != nil {
					return AppendEntriesResponse{Term: s.CurrentTerm, Success: false}
				}
			}
		} else {
			if err := s.wal.Append(newEntry); err != nil {
				return AppendEntriesResponse{Term: s.CurrentTerm, Success: false}
			}
		}
	}

	// Commit update
	if req.LeaderCommit > s.CommitIndex {
		s.CommitIndex = min(req.LeaderCommit, s.wal.LastIndex)
		s.cond.Broadcast()
	}

	return AppendEntriesResponse{Success: true, Term: s.CurrentTerm}
}
