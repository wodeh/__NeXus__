package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerHousekeepingHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/housekeeping/tasks", s.withTenant(s.handleHousekeepingTasks))
	mux.HandleFunc("/v1/housekeeping/tasks/", s.withTenant(s.handleHousekeepingTaskDetail))
	mux.HandleFunc("/v1/housekeeping/staff", s.withTenant(s.handleHousekeepingStaff))
}

func (s *Server) handleHousekeepingTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	
	repo := repository.NewHousekeepingRepository(s.repo.Pool())
	
	switch r.Method {
	case http.MethodGet:
		tasks, err := repo.ListTasks(ctx, tenantID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
		
	case http.MethodPost:
		var req domain.HousekeepingTaskCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		task, err := repo.CreateTask(ctx, tenantID, &req)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, task)
		
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleHousekeepingTaskDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	
	path := r.URL.Path[len("/v1/housekeeping/tasks/"):]
	parts := splitPath(path)
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing task id"}`, http.StatusBadRequest)
		return
	}
	
	id, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, `{"error":"invalid task id"}`, http.StatusBadRequest)
		return
	}
	
	repo := repository.NewHousekeepingRepository(s.repo.Pool())
	
	if r.Method != http.MethodPatch {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	
	var req domain.HousekeepingTaskUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	
	if err := repo.UpdateTask(ctx, tenantID, id, &req); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) handleHousekeepingStaff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	
	repo := repository.NewHousekeepingRepository(s.repo.Pool())
	
	if r.Method == http.MethodGet {
		staff, err := repo.ListStaff(ctx, tenantID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"staff": staff})
		return
	}
	
	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}
