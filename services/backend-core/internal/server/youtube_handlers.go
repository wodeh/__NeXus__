package server

import (
	"net/http"
	"strings"
)

func (s *Server) registerYouTubeHandlers(mux *http.ServeMux) {
	// YouTube integration endpoints
	mux.HandleFunc("/v1/iptv/youtube/search", s.withTenant(s.handleYouTubeSearch))
	mux.HandleFunc("/v1/iptv/youtube/play/", s.withTenant(s.handleYouTubePlay))
	mux.HandleFunc("/v1/iptv/youtube/trending", s.withTenant(s.handleYouTubeTrending))
}

// handleYouTubeSearch searches YouTube videos
func (s *Server) handleYouTubeSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		query = "hotel welcome"
	}

	// For demo, return sample results
	// In production, integrate with YouTube Data API v3
	results := []map[string]interface{}{
		{
			"id":          "demo1",
			"title":       "Welcome to Grand Plaza Hotel",
			"description": "Experience luxury at Grand Plaza Hotel",
			"thumbnail":   "https://i.ytimg.com/vi/demo1/hqdefault.jpg",
			"channel":     "Grand Plaza Official",
			"duration":    "2:30",
			"video_id":    "demo1",
		},
		{
			"id":          "demo2",
			"title":       "Local Attractions Guide",
			"description": "Discover the best places near our hotel",
			"thumbnail":   "https://i.ytimg.com/vi/demo2/hqdefault.jpg",
			"channel":     "Travel Guide",
			"duration":    "5:45",
			"video_id":    "demo2",
		},
		{
			"id":          "demo3",
			"title":       "Spa & Wellness Introduction",
			"description": "Relax and rejuvenate at our spa",
			"thumbnail":   "https://i.ytimg.com/vi/demo3/hqdefault.jpg",
			"channel":     "Grand Plaza Spa",
			"duration":    "3:15",
			"video_id":    "demo3",
		},
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"query":   query,
	})
}

// handleYouTubePlay returns playback info for a video
func (s *Server) handleYouTubePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	videoID := extractID(r.URL.Path, "/v1/iptv/youtube/play/")
	if videoID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing video id"})
		return
	}

	// Build embed URL
	embedURL := "https://www.youtube.com/embed/" + videoID + "?autoplay=1&rel=0&modestbranding=1"
	
	// For demo videos, provide local URLs
	if strings.HasPrefix(videoID, "demo") {
		embedURL = "/iptv/assets/demo-video-" + videoID + ".mp4"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"video_id":  videoID,
		"embed_url": embedURL,
		"type":      "youtube",
		"quality":   "hd1080",
	})
}

// handleYouTubeTrending returns trending videos
func (s *Server) handleYouTubeTrending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	// Demo trending content for hospitality
	videos := []map[string]interface{}{
		{
			"id":        "trend1",
			"title":     "Top 10 Luxury Hotels 2026",
			"thumbnail": "https://i.ytimg.com/vi/trend1/hqdefault.jpg",
			"channel":   "Travel + Leisure",
			"views":     "1.2M",
			"video_id":  "trend1",
		},
		{
			"id":        "trend2", 
			"title":     "Best Room Service Experiences",
			"thumbnail": "https://i.ytimg.com/vi/trend2/hqdefault.jpg",
			"channel":   "Hotel Management",
			"views":     "856K",
			"video_id":  "trend2",
		},
		{
			"id":        "trend3",
			"title":     "Hotel Technology Trends",
			"thumbnail": "https://i.ytimg.com/vi/trend3/hqdefault.jpg",
			"channel":   "Hospitality Tech",
			"views":     "430K",
			"video_id":  "trend3",
		},
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"videos": videos,
	})
}
