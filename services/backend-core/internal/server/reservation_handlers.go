package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerReservationHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/reservations", s.withTenant(s.handleReservations))
}

func (s *Server) handleReservations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	
	repo := repository.NewReservationRepository(s.repo.Pool())
	
	switch r.Method {
	case http.MethodGet:
		reservations, err := repo.List(ctx, tenantID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"reservations": reservations})
		
	case http.MethodPost:
		var req domain.ReservationCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		res, err := repo.Create(ctx, nil, tenantID, &req)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, res)
		
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleReservationDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	
	// Extract ID from path: /v1/reservations/{id}/action
	path := r.URL.Path[len("/v1/reservations/"):]
	parts := splitPath(path)
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing reservation id"}`, http.StatusBadRequest)
		return
	}
	
	id, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, `{"error":"invalid reservation id"}`, http.StatusBadRequest)
		return
	}
	
	repo := repository.NewReservationRepository(s.repo.Pool())
	
	// Check for action suffix
	if len(parts) > 1 {
		action := parts[1]
		switch action {
		case "move":
			if r.Method != http.MethodPatch {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			var req domain.ReservationMoveRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
				return
			}
			updated, err := repo.Move(ctx, tenantID, id, &req)
			if err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, updated)
			return
			
		case "checkin":
			if r.Method != http.MethodPatch {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			if err := repo.CheckIn(ctx, tenantID, id); err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "checked_in"})
			return
			
		case "checkout":
			if r.Method != http.MethodPatch {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			if err := repo.CheckOut(ctx, tenantID, id); err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "checked_out"})
			return
			
		case "cancel":
			if r.Method != http.MethodPatch {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			if err := repo.Cancel(ctx, tenantID, id); err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
			return
			
		case "assign-room":
			if r.Method != http.MethodPatch {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			var req domain.ReservationAssignRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
				return
			}
			if err := repo.AssignRoom(ctx, tenantID, id, req.RoomNumber); err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"room_number": req.RoomNumber})
			return
		}
	}
	
	// GET single reservation
	if r.Method == http.MethodGet {
		res, err := repo.Get(ctx, tenantID, id)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, res)
		return
	}
	
	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func splitPath(path string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}
