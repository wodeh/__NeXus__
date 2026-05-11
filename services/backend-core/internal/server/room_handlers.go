package server

import (
	"encoding/json"
	"net/http"

	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerRoomHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/rooms", s.withTenant(s.handleRooms))
	mux.HandleFunc("/v1/rooms/", s.withTenant(s.handleRoomDetail))
}

func (s *Server) handleRooms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	
	repo := repository.NewRoomRepository(s.repo.Pool())
	
	if r.Method == http.MethodGet {
		rooms, err := repo.List(ctx, tenantID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"rooms": rooms})
		return
	}
	
	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func (s *Server) handleRoomDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	
	path := r.URL.Path[len("/v1/rooms/"):]
	parts := splitPath(path)
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing room number"}`, http.StatusBadRequest)
		return
	}
	
	number := parts[0]
	repo := repository.NewRoomRepository(s.repo.Pool())
	
	if len(parts) > 1 && parts[1] == "status" {
		if r.Method != http.MethodPatch {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req domain.RoomStatusUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if err := repo.UpdateStatus(ctx, tenantID, number, req.Status); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": req.Status})
		return
	}
	
	http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
}
