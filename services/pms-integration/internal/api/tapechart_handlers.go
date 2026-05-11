package api

import (
	"net/http"
	"time"
)

// ==================== TAPECHART ====================

func (h *Handler) getTapechart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)

	propertyID := r.URL.Query().Get("propertyId")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if propertyID == "" || startStr == "" || endStr == "" {
		respondError(w, http.StatusBadRequest, "propertyId, start, end required")
		return
	}

	start, err := parseDate(startStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid start date")
		return
	}
	end, err := parseDate(endStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid end date")
		return
	}

	entries, err := h.store.Tapechart.GetTapechart(ctx, tenant, propertyID, start, end)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Calculate daily stats
	days := int(end.Sub(start).Hours()/24) + 1
	dailyStats := make([]map[string]interface{}, days)
	for i := 0; i < days; i++ {
		day := start.Add(time.Duration(i) * 24 * time.Hour)
		totalRooms := len(entries)
		occupied := 0
		for _, e := range entries {
			for _, res := range e.Reservations {
				if day.Equal(res.CheckInDate) || day.Equal(res.CheckOutDate) || (day.After(res.CheckInDate) && day.Before(res.CheckOutDate)) {
					occupied++
					break
				}
			}
		}
		dailyStats[i] = map[string]interface{}{
			"date":            day.Format("2006-01-02"),
			"occupancy_pct":   0.0,
			"available_rooms": totalRooms - occupied,
			"occupied_rooms":  occupied,
			"total_rooms":     totalRooms,
		}
		if totalRooms > 0 {
			dailyStats[i]["occupancy_pct"] = float64(occupied) / float64(totalRooms) * 100
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"property_id":  propertyID,
		"start_date":   startStr,
		"end_date":     endStr,
		"days":         days,
		"entries":      entries,
		"daily_stats":  dailyStats,
	})
}
