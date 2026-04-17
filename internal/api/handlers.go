package api

import (
	"encoding/json"
	"net/http"

	"github.com/vijayvenkatj/kv-store/internal/server/wal"
)

type Handler struct {
	NodeID uint32
	SM     *ShardManager
}

func NewHandler(sm *ShardManager, config Config) *Handler {
	return &Handler{
		NodeID: config.NodeID,
		SM:     sm,
	}
}

func (handler *Handler) GetValueHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "missing key"})
		return
	}

	shardID := handler.SM.getShardID(key)
	store, err := handler.SM.GetLocalShard(shardID)
	if err != nil {
		shardAddr := handler.SM.ShardLocation(shardID).Address
		url := "http://" + shardAddr + "/api/v1/value"
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	val, err := store.Get(key)
	if err != nil {
		writeJSON(w, http.StatusNotFound, Response{Error: "not found"})
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Data: map[string]string{
			"key":   key,
			"value": val,
		},
	})
}

func (handler *Handler) PutValueHandler(w http.ResponseWriter, r *http.Request) {
	var req putRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Key == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "invalid request"})
		return
	}

	shardID := handler.SM.getShardID(req.Key)
	store, err := handler.SM.GetLocalShard(shardID)
	if err != nil {
		shardAddr := handler.SM.ShardLocation(shardID).Address
		url := "http://" + shardAddr + "/api/v1/value"
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	if !store.IsLeader() {
		leaderID := store.LeaderId

		if leaderID == 0 {
			writeJSON(w, 503, Response{Error: "leader unknown"})
			return
		}

		url := "http://" + handler.SM.NodeMap[leaderID] + "/api/v1/value"

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	logEntry := &wal.LogEntry{
		Operation: "put",
		Key:       req.Key,
		Value:     req.Value,
	}
	if err := store.Apply(logEntry); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Data: "ok",
	})
}

func (handler *Handler) DeleteValueHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		writeJSON(w, http.StatusBadRequest, Response{Error: "missing key"})
		return
	}

	shardID := handler.SM.getShardID(key)
	store, err := handler.SM.GetLocalShard(shardID)
	if err != nil {
		shardAddr := handler.SM.ShardLocation(shardID).Address
		url := "http://" + shardAddr + "/api/v1/value"
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	if !store.IsLeader() {
		leaderID := store.LeaderId
		if leaderID == 0 {
			writeJSON(w, 503, Response{Error: "leader unknown"})
			return
		}
		url := "http://" + handler.SM.NodeMap[leaderID] + "/api/v1/value"
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	logEntry := &wal.LogEntry{
		Operation: "delete",
		Key:       key,
	}

	if err := store.Apply(logEntry); err != nil {
		writeJSON(w, http.StatusNotFound, Response{Error: "not found"})
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Data: "deleted",
	})
}
