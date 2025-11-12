package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/database"
)

type DebugHandler struct {
	db     *database.Database
	logger *slog.Logger
}

func NewDebugHandler(db *database.Database, logger *slog.Logger) *DebugHandler {
	return &DebugHandler{
		db:     db,
		logger: logger.With("component", "debug_handler"),
	}
}

func (h *DebugHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("http request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.UserAgent())

	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/debug/health":
		h.handleHealth(w, r)
	case "/debug/schema":
		h.handleSchema(w, r)
	case "/debug/ping":
		h.handlePing(w, r)
	case "/debug/users":
		h.handleUsers(w, r)
	default:
		h.handleIndex(w, r)
	}
}

func (h *DebugHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service": "analyzer",
		"status":  "running",
		"endpoints": []string{
			"/debug/health",
			"/debug/schema",
			"/debug/ping",
			"/debug/users",
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (h *DebugHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.db.Pool().Ping(ctx)
	if err != nil {
		h.logger.Error("health check failed", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	h.logger.Info("health check ok")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"db":     "connected",
	})
}

func (h *DebugHandler) handleSchema(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tables, err := h.db.GetTablesSchema(ctx)
	if err != nil {
		h.logger.Error("failed to get schema", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("schema retrieved", "tables_count", len(tables))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tables": tables,
		"count":  len(tables),
	})
}

func (h *DebugHandler) handlePing(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var result int
	err := h.db.Pool().QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		h.logger.Error("database ping failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("database query failed: %v", err),
		})
		return
	}

	h.logger.Info("database ping ok", "result", result)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ping":   "ok",
		"result": result,
	})
}

func (h *DebugHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := `SELECT id, login, created_at FROM users ORDER BY created_at DESC LIMIT 100`

	rows, err := h.db.Pool().Query(ctx, query)
	if err != nil {
		h.logger.Error("failed to query users", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("query failed: %v", err),
		})
		return
	}
	defer rows.Close()

	type User struct {
		ID        string `json:"id"`
		Login     string `json:"login"`
		CreatedAt string `json:"created_at"`
	}

	var users []User

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Login, &user.CreatedAt); err != nil {
			h.logger.Error("failed to scan user row", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": fmt.Sprintf("scan failed: %v", err),
			})
			return
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		h.logger.Error("error iterating user rows", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("iteration failed: %v", err),
		})
		return
	}

	h.logger.Info("users retrieved", "count", len(users))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}
