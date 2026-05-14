package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/repository"
)

// ── WiFi Handlers ──

// registerWiFiHandlers mounts WiFi monitoring routes.
func (s *Server) registerWiFiHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/wifi/floors", s.withTenant(s.handleWiFiFloors))
	mux.HandleFunc("/v1/wifi/floors/", s.withTenant(s.handleWiFiFloorDetail))
	mux.HandleFunc("/v1/wifi/aps", s.withTenant(s.handleWiFiAPs))
	mux.HandleFunc("/v1/wifi/aps/", s.withTenant(s.handleWiFiAPDetail))
	mux.HandleFunc("/v1/wifi/metrics", s.withTenant(s.handleWiFiMetrics))
	mux.HandleFunc("/v1/wifi/alerts", s.withTenant(s.handleWiFiAlerts))
	mux.HandleFunc("/v1/wifi/alerts/", s.withTenant(s.handleWiFiAlertDetail))
	mux.HandleFunc("/v1/wifi/heatmap", s.withTenant(s.handleWiFiHeatmap))
	mux.HandleFunc("/v1/wifi/scan", s.withTenant(s.handleWiFiScan))
	mux.HandleFunc("/v1/wifi/dashboard", s.withTenant(s.handleWiFiDashboard))
}

// handleWiFiFloors returns all floor summaries.
func (s *Server) handleWiFiFloors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tenantID, ok := getTenantID(r)
	if !ok || tenantID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing tenant context")
		return
	}

	if s.repo == nil || s.repo.WiFi == nil {
		// Return demo data when DB is not available
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"floors": []map[string]interface{}{
				{"floor": "Lobby", "ap_count": 4, "online_aps": 4, "avg_signal_dbm": -45, "total_clients": 87, "active_alerts": 0, "overall_health": "excellent"},
				{"floor": "Floor 1", "ap_count": 6, "online_aps": 6, "avg_signal_dbm": -52, "total_clients": 43, "active_alerts": 0, "overall_health": "good"},
				{"floor": "Floor 2", "ap_count": 6, "online_aps": 5, "avg_signal_dbm": -58, "total_clients": 38, "active_alerts": 1, "overall_health": "fair"},
				{"floor": "Floor 3", "ap_count": 6, "online_aps": 6, "avg_signal_dbm": -48, "total_clients": 41, "active_alerts": 0, "overall_health": "excellent"},
				{"floor": "Floor 4", "ap_count": 6, "online_aps": 6, "avg_signal_dbm": -55, "total_clients": 35, "active_alerts": 0, "overall_health": "good"},
				{"floor": "Floor 5", "ap_count": 6, "online_aps": 4, "avg_signal_dbm": -67, "total_clients": 29, "active_alerts": 3, "overall_health": "poor"},
			},
		})
		return
	}

	tid, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	summaries, err := s.repo.WiFi.ListFloorSummaries(r.Context(), tid)
	if err != nil {
		slog.Error("failed to list floor summaries", slog.String("error", err.Error()), slog.String("tenant_id", tenantID))
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"floors": summaries})
}

// handleWiFiFloorDetail returns detailed info for a specific floor.
func (s *Server) handleWiFiFloorDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	floor := extractPathParam(r.URL.Path, "/v1/wifi/floors/")
	if floor == "" {
		writeJSONError(w, http.StatusBadRequest, "missing floor parameter")
		return
	}

	tenantID, ok := getTenantID(r)
	if !ok || tenantID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing tenant context")
		return
	}

	if s.repo == nil || s.repo.WiFi == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"floor": floor, "aps": []interface{}{}, "metrics": []interface{}{}, "alerts": []interface{}{}})
		return
	}

	tid, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	summary, _ := s.repo.WiFi.GetFloorSummary(r.Context(), tid, floor)
	aps, _ := s.repo.WiFi.ListAccessPoints(r.Context(), tid)
	var floorAPs []repository.WiFiAccessPoint
	for _, ap := range aps {
		if ap.Floor == floor {
			floorAPs = append(floorAPs, ap)
		}
	}
	metrics, _ := s.repo.WiFi.GetFloorMetrics(r.Context(), tid, floor)
	alerts, _ := s.repo.WiFi.ListAlerts(r.Context(), tid, floor, true)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"floor":   floor,
		"summary": summary,
		"aps":     floorAPs,
		"metrics": metrics,
		"alerts":  alerts,
	})
}

// handleWiFiAPs lists all access points.
func (s *Server) handleWiFiAPs(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok || tenantID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing tenant context")
		return
	}

	if s.repo == nil || s.repo.WiFi == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"aps": []interface{}{}})
		return
	}

	tid, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	aps, err := s.repo.WiFi.ListAccessPoints(r.Context(), tid)
	if err != nil {
		slog.Error("failed to list APs", slog.String("error", err.Error()))
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"aps": aps})
}

// handleWiFiAPDetail handles single AP CRUD.
func (s *Server) handleWiFiAPDetail(w http.ResponseWriter, r *http.Request) {
	apIDStr := extractPathParam(r.URL.Path, "/v1/wifi/aps/")
	apID, err := uuid.Parse(apIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid AP ID")
		return
	}

	tenantID, ok := getTenantID(r)
	if !ok || tenantID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing tenant context")
		return
	}
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		if s.repo == nil || s.repo.WiFi == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "database not configured")
			return
		}
		ap, err := s.repo.WiFi.GetAccessPoint(r.Context(), tid, apID)
		if err != nil {
			writeJSONError(w, http.StatusNotFound, "AP not found")
			return
		}

		// Get last 24h of metrics
		since := time.Now().Add(-24 * time.Hour)
		history, _ := s.repo.WiFi.GetAPMetricsHistory(r.Context(), tid, apID, since)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ap":      ap,
			"history": history,
		})

	case http.MethodPatch:
		if s.repo == nil || s.repo.WiFi == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "database not configured")
			return
		}
		var req repository.WiFiAccessPoint
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		req.ID = apID
		req.TenantID = tid
		if err := s.repo.WiFi.UpdateAccessPoint(r.Context(), &req); err != nil {
			slog.Error("failed to update AP", slog.String("error", err.Error()))
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("update failed: %v", err))
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})

	case http.MethodDelete:
		if s.repo == nil || s.repo.WiFi == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "database not configured")
			return
		}
		if err := s.repo.WiFi.DeleteAccessPoint(r.Context(), tid, apID); err != nil {
			slog.Error("failed to delete AP", slog.String("error", err.Error()))
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("delete failed: %v", err))
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleWiFiMetrics returns metrics for a floor or AP.
func (s *Server) handleWiFiMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tenantID, ok := getTenantID(r)
	if !ok || tenantID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing tenant context")
		return
	}

	floor := r.URL.Query().Get("floor")
	apIDStr := r.URL.Query().Get("ap_id")

	if s.repo == nil || s.repo.WiFi == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"metrics": []interface{}{}})
		return
	}

	tid, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	var metrics []repository.WiFiMetric
	if floor != "" {
		metrics, err = s.repo.WiFi.GetFloorMetrics(r.Context(), tid, floor)
	} else if apIDStr != "" {
		apID, err := uuid.Parse(apIDStr)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid AP ID")
			return
		}
		metrics, err = s.repo.WiFi.GetAPMetricsHistory(r.Context(), tid, apID, time.Now().Add(-24*time.Hour))
	} else {
		writeJSONError(w, http.StatusBadRequest, "specify floor or ap_id")
		return
	}

	if err != nil {
		slog.Error("failed to get metrics", slog.String("error", err.Error()))
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"metrics": metrics})
}

// handleWiFiAlerts lists or creates alerts.
func (s *Server) handleWiFiAlerts(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := getTenantID(r)
	if !ok || tenantID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing tenant context")
		return
	}

	tid, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		floor := r.URL.Query().Get("floor")
		unresolvedOnly := r.URL.Query().Get("unresolved") == "true"

		if s.repo == nil || s.repo.WiFi == nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"alerts": []interface{}{}})
			return
		}

		alerts, err := s.repo.WiFi.ListAlerts(r.Context(), tid, floor, unresolvedOnly)
		if err != nil {
			slog.Error("failed to list alerts", slog.String("error", err.Error()))
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"alerts": alerts})

	case http.MethodPost:
		if s.repo == nil || s.repo.WiFi == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "database not configured")
			return
		}
		var alert repository.WiFiAlert
		if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		alert.TenantID = tid
		if err := s.repo.WiFi.CreateAlert(r.Context(), &alert); err != nil {
			slog.Error("failed to create alert", slog.String("error", err.Error()))
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("create failed: %v", err))
			return
		}
		writeJSON(w, http.StatusCreated, alert)

	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleWiFiAlertDetail resolves or gets a single alert.
func (s *Server) handleWiFiAlertDetail(w http.ResponseWriter, r *http.Request) {
	alertIDStr := extractPathParam(r.URL.Path, "/v1/wifi/alerts/")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid alert ID")
		return
	}

	tenantID, ok := getTenantID(r)
	if !ok || tenantID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing tenant context")
		return
	}
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	if r.Method != http.MethodPatch {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if s.repo == nil || s.repo.WiFi == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	var req struct {
		ResolvedBy string `json:"resolved_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.ResolvedBy = "system"
	}
	if req.ResolvedBy == "" {
		req.ResolvedBy = "system"
	}

	if err := s.repo.WiFi.ResolveAlert(r.Context(), tid, alertID, req.ResolvedBy); err != nil {
		slog.Error("failed to resolve alert", slog.String("error", err.Error()))
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("resolve failed: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

// handleWiFiHeatmap returns synthetic grid data for a floor heatmap.
func (s *Server) handleWiFiHeatmap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	floor := r.URL.Query().Get("floor")
	if floor == "" {
		writeJSONError(w, http.StatusBadRequest, "missing floor parameter")
		return
	}

	// Generate a 20x10 grid of signal readings
	// In production this would be based on actual AP locations and propagation models
	grid := make([][]int, 10)
	for y := 0; y < 10; y++ {
		grid[y] = make([]int, 20)
		for x := 0; x < 20; x++ {
			// Simulate signal with AP positions at certain coordinates
			// This is synthetic data for demo
			distances := []float64{
				math.Sqrt(float64((x-3)*(x-3) + (y-2)*(y-2))),   // AP 1
				math.Sqrt(float64((x-15)*(x-15) + (y-3)*(y-3))),  // AP 2
				math.Sqrt(float64((x-8)*(x-8) + (y-7)*(y-7))),    // AP 3
			}
			// Best signal from nearest AP
			minDist := distances[0]
			for _, d := range distances[1:] {
				if d < minDist {
					minDist = d
				}
			}
			// Convert distance to dBm (closer = better)
			signal := -30 - int(minDist*8)
			if signal < -90 {
				signal = -90
			}
			if signal > -30 {
				signal = -30
			}
			grid[y][x] = signal
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"floor":       floor,
		"grid_width":  20,
		"grid_height": 10,
		"cell_size_m": 5,
		"unit":        "dBm",
		"grid":        grid,
		"legend": map[string]interface{}{
			"excellent": "-30 to -50 dBm",
			"good":      "-50 to -60 dBm",
			"fair":      "-60 to -70 dBm",
			"poor":      "-70+ dBm",
		},
	})
}

// handleWiFiScan triggers a manual SNMP scan (simulated for now).
func (s *Server) handleWiFiScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tenantID, ok := getTenantID(r)
	if !ok || tenantID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing tenant context")
		return
	}

	floor := r.URL.Query().Get("floor")

	// Simulate scan results
	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":     "scanning",
		"floor":      floor,
		"message":    "SNMP scan initiated",
		"estimated":  "30 seconds",
		"scan_id":    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
	})
}

// handleWiFiDashboard returns a consolidated dashboard view.
func (s *Server) handleWiFiDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tenantID, ok := getTenantID(r)
	if !ok || tenantID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing tenant context")
		return
	}

	if s.repo == nil || s.repo.WiFi == nil {
		// Demo dashboard
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"overview": map[string]interface{}{
				"total_aps":        34,
				"online_aps":       31,
				"total_clients":    273,
				"active_alerts":    4,
				"overall_health":   "good",
				"avg_signal_dbm":   -53,
			},
			"floors": []map[string]interface{}{
				{"floor": "Lobby", "health": "excellent", "signal": -45, "clients": 87, "alerts": 0, "online_aps": 4, "total_aps": 4},
				{"floor": "Floor 1", "health": "good", "signal": -52, "clients": 43, "alerts": 0, "online_aps": 6, "total_aps": 6},
				{"floor": "Floor 2", "health": "fair", "signal": -58, "clients": 38, "alerts": 1, "online_aps": 5, "total_aps": 6},
				{"floor": "Floor 3", "health": "excellent", "signal": -48, "clients": 41, "alerts": 0, "online_aps": 6, "total_aps": 6},
				{"floor": "Floor 4", "health": "good", "signal": -55, "clients": 35, "alerts": 0, "online_aps": 6, "total_aps": 6},
				{"floor": "Floor 5", "health": "poor", "signal": -67, "clients": 29, "alerts": 3, "online_aps": 4, "total_aps": 6},
			},
			"alerts": []map[string]interface{}{
				{"id": "1", "floor": "Floor 2", "type": "ap_offline", "severity": "critical", "message": "AP F2-AP-01 offline for 15 min", "suggested_fix": "Check power and Ethernet"},
				{"id": "2", "floor": "Floor 5", "type": "poor_signal", "severity": "warning", "message": "F5-AP-01 signal at -68 dBm", "suggested_fix": "Check placement, add AP nearby"},
				{"id": "3", "floor": "Floor 5", "type": "channel_conflict", "severity": "warning", "message": "Channel overlap on 5GHz ch 48", "suggested_fix": "Switch to channel 52 or 56"},
				{"id": "4", "floor": "Floor 5", "type": "overloaded", "severity": "warning", "message": "F5-AP-03 utilization at 58%", "suggested_fix": "Enable band steering"},
			},
			"trends": map[string]interface{}{
				"24h_signal_avg":   []int{-50, -51, -52, -53, -54, -55, -54, -53, -52, -51, -50, -52, -53, -55, -56, -57, -58, -57, -56, -55, -54, -53, -52, -51},
				"24h_client_count": []int{245, 250, 260, 270, 280, 275, 265, 255, 250, 248, 252, 258, 265, 270, 272, 268, 265, 260, 255, 250, 248, 245, 243, 240},
			},
		})
		return
	}

	tid, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	floors, err := s.repo.WiFi.ListFloorSummaries(r.Context(), tid)
	if err != nil {
		slog.Error("failed to get floor summaries", slog.String("error", err.Error()))
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		return
	}

	alerts, err := s.repo.WiFi.ListAlerts(r.Context(), tid, "", true)
	if err != nil {
		slog.Error("failed to get alerts", slog.String("error", err.Error()))
	}

	// Calculate totals
	var totalAPs, onlineAPs, totalClients, activeAlerts int32
	var totalSignal int32
	var signalCount int32
	for _, f := range floors {
		totalAPs += f.APCount
		onlineAPs += f.OnlineAPs
		totalClients += f.TotalClients
		activeAlerts += f.ActiveAlerts
		if f.AvgSignalDbm != nil {
			totalSignal += *f.AvgSignalDbm
			signalCount++
		}
	}

	avgSignal := -60
	if signalCount > 0 {
		avgSignal = int(totalSignal / signalCount)
	}

	health := "good"
	if avgSignal >= -50 {
		health = "excellent"
	} else if avgSignal >= -60 {
		health = "good"
	} else if avgSignal >= -70 {
		health = "fair"
	} else {
		health = "poor"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"overview": map[string]interface{}{
			"total_aps":      totalAPs,
			"online_aps":     onlineAPs,
			"total_clients":  totalClients,
			"active_alerts":  activeAlerts,
			"overall_health": health,
			"avg_signal_dbm": avgSignal,
		},
		"floors": floors,
		"alerts": alerts,
	})
}

// ── SNMP Polling Worker ──

// StartWiFiPoller starts a background goroutine that simulates SNMP polling.
// In production this would use gosnmp to actually poll the APs.
func (s *Server) StartWiFiPoller() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			if s.repo == nil || s.repo.WiFi == nil {
				continue
			}
			// In production: iterate all APs, SNMP poll each one, store metrics
			// For now: simulate metric updates for demo APs
			s.simulateWiFiMetrics()
		}
	}()
}

// simulateWiFiMetrics generates synthetic metrics for demonstration.
func (s *Server) simulateWiFiMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get all tenants with WiFi APs
	// For demo, just use DOWNTOWN tenant
	tid := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	aps, err := s.repo.WiFi.ListAccessPoints(ctx, tid)
	if err != nil {
		slog.Error("wifi poller: failed to list APs", slog.String("error", err.Error()))
		return
	}

	for _, ap := range aps {
		// Skip offline APs
		if ap.Status == "offline" {
			continue
		}

		// Simulate realistic signal based on floor
		baseRSSI := int32(-45)
		switch ap.Floor {
		case "Floor 1":
			baseRSSI = -52
		case "Floor 2":
			baseRSSI = -58
		case "Floor 3":
			baseRSSI = -50
		case "Floor 4":
			baseRSSI = -55
		case "Floor 5":
			baseRSSI = -68
		}

		// Add random variance
		variance := rand.Int31n(8) - 4
		rssi := baseRSSI + variance
		noise := int32(-90 - rand.Int31n(8))
		snr := rssi - noise
		if snr < 0 {
			snr = 0
		}

		// Quality score 0-100
		quality := int32(100)
		if rssi >= -50 {
			quality = 90 + rand.Int31n(10)
		} else if rssi >= -60 {
			quality = 75 + rand.Int31n(15)
		} else if rssi >= -70 {
			quality = 50 + rand.Int31n(20)
		} else {
			quality = 20 + rand.Int31n(30)
		}

		clientCount := int32(5 + rand.Int31n(20))
		bandwidth := 100.0 + rand.Float64()*400.0
		if ap.Floor == "Floor 5" {
			bandwidth = 50.0 + rand.Float64()*150.0
		}

		m := &repository.WiFiMetric{
			TenantID:      tid,
			APID:          ap.ID,
			Floor:         ap.Floor,
			RSSIDbm:       &rssi,
			NoiseDbm:      &noise,
			SNRDb:         &snr,
			QualityScore:  &quality,
			ClientCount:   clientCount,
			BandwidthMbps: &bandwidth,
		}

		if err := s.repo.WiFi.InsertMetric(ctx, m); err != nil {
			slog.Error("wifi poller: failed to insert metric", slog.String("ap", ap.Name), slog.String("error", err.Error()))
		}
	}

	// Check for alert conditions
	s.checkWiFiAlerts(ctx, tid)
}

// checkWiFiAlerts evaluates metrics and creates alerts.
func (s *Server) checkWiFiAlerts(ctx context.Context, tenantID uuid.UUID) {
	floors, err := s.repo.WiFi.ListFloorSummaries(ctx, tenantID)
	if err != nil {
		slog.Error("wifi alerts: failed to get floors", slog.String("error", err.Error()))
		return
	}

	for _, floor := range floors {
		if floor.OverallHealth == "poor" {
			// Check if alert already exists
			alerts, _ := s.repo.WiFi.ListAlerts(ctx, tenantID, floor.Floor, true)
			var hasPoorSignal bool
			for _, a := range alerts {
				if a.Type == "poor_signal" {
					hasPoorSignal = true
					break
				}
			}
			if !hasPoorSignal && floor.AvgSignalDbm != nil && *floor.AvgSignalDbm < -65 {
				alert := &repository.WiFiAlert{
					TenantID:     tenantID,
					Floor:        floor.Floor,
					Type:         "poor_signal",
					Severity:     "warning",
					Message:      fmt.Sprintf("%s average signal degraded to %d dBm", floor.Floor, *floor.AvgSignalDbm),
					SuggestedFix: func(s string) *string { return &s }("Check AP placement and consider adding an access point"),
				}
				_ = s.repo.WiFi.CreateAlert(ctx, alert)
			}
		}
	}
}

// getTenantID extracts tenant_id from request context (set by withTenant middleware).
func getTenantID(r *http.Request) (string, bool) {
	tenantID, ok := r.Context().Value("tenant_id").(string)
	return tenantID, ok
}

// extractPathParam extracts a trailing path parameter.
func extractPathParam(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	param := strings.TrimPrefix(path, prefix)
	// Remove trailing slash if present
	param = strings.TrimSuffix(param, "/")
	return param
}
