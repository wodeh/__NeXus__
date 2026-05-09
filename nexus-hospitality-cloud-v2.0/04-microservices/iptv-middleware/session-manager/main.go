package main

// nhc-iptv-session-manager
// Production-grade session management for hospitality IPTV
// Handles guest authentication, stream entitlement, concurrent limits

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/nexushc/iptv-proto/gen/go/session/v1"
)

const (
	ServiceName          = "iptv-session-manager"
	ServiceVersion       = "2.3.1"
	MaxConcurrentStreams = 3
	SessionTTL           = 24 * time.Hour
	JWTSecretEnv         = "IPTV_JWT_SECRET"
)

var (
	tracer = otel.Tracer(ServiceName)

	sessionsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_sessions_created_total",
			Help: "Total number of IPTV sessions created",
		},
		[]string{"tenant_id", "property_id", "device_type"},
	)

	sessionsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "iptv_sessions_active",
			Help: "Number of active IPTV sessions",
		},
		[]string{"tenant_id", "property_id"},
	)

	streamStarts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_stream_starts_total",
			Help: "Total number of stream starts",
		},
		[]string{"tenant_id", "content_type", "quality"},
	)

	streamErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_stream_errors_total",
			Help: "Total number of stream errors",
		},
		[]string{"tenant_id", "error_type"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "iptv_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)
)

func init() {
	prometheus.MustRegister(sessionsCreated, sessionsActive, streamStarts, streamErrors, requestDuration)
}

type Session struct {
	SessionID        string                 `json:"session_id"`
	TenantID         string                 `json:"tenant_id"`
	PropertyID       string                 `json:"property_id"`
	RoomID           string                 `json:"room_id"`
	GuestID          string                 `json:"guest_id"`
	ReservationID    string                 `json:"reservation_id"`
	DeviceID         string                 `json:"device_id"`
	DeviceType       string                 `json:"device_type"`
	DeviceModel      string                 `json:"device_model"`
	IPAddress        string                 `json:"ip_address"`
	MACAddress       string                 `json:"mac_address"`
	Status           string                 `json:"status"`
	CurrentChannel   *string                `json:"current_channel,omitempty"`
	CurrentVOD       *string                `json:"current_vod,omitempty"`
	Quality          string                 `json:"quality"`
	ConcurrentCount  int                    `json:"concurrent_count"`
	MaxConcurrent    int                    `json:"max_concurrent"`
	Entitlements     []string               `json:"entitlements"`
	Preferences      map[string]interface{} `json:"preferences"`
	ParentalControls *ParentalControls      `json:"parental_controls,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	ExpiresAt        time.Time              `json:"expires_at"`
	LastActivity     time.Time              `json:"last_activity"`
}

type ParentalControls struct {
	Enabled         bool     `json:"enabled"`
	MaxRating       string   `json:"max_rating"`
	BlockedChannels []string `json:"blocked_channels"`
	TimeWindow      *struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"time_window,omitempty"`
}

type SessionStore struct {
	redis      *redis.Client
	localCache *sync.Map
	logger     *zap.Logger
}

func NewSessionStore(redisAddr string, logger *zap.Logger) *SessionStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           0,
		PoolSize:     50,
		MinIdleConns: 10,
		MaxRetries:   3,
	})
	return &SessionStore{redis: rdb, localCache: &sync.Map{}, logger: logger}
}

func (s *SessionStore) CreateSession(ctx context.Context, session *Session) error {
	ctx, span := tracer.Start(ctx, "SessionStore.CreateSession")
	defer span.End()
	span.SetAttributes(
		attribute.String("session.id", session.SessionID),
		attribute.String("tenant.id", session.TenantID),
	)

	activeCount, err := s.GetActiveSessionCount(ctx, session.TenantID, session.PropertyID, session.RoomID)
	if err != nil {
		return fmt.Errorf("concurrent check failed: %w", err)
	}
	if activeCount >= session.MaxConcurrent {
		return fmt.Errorf("concurrent stream limit exceeded: %d/%d", activeCount, session.MaxConcurrent)
	}

	sessionData, _ := json.Marshal(session)
	key := fmt.Sprintf("iptv:session:%s", session.SessionID)
	pipe := s.redis.Pipeline()
	pipe.Set(ctx, key, sessionData, SessionTTL)
	pipe.SAdd(ctx, fmt.Sprintf("iptv:tenant:%s:sessions", session.TenantID), session.SessionID)
	pipe.SAdd(ctx, fmt.Sprintf("iptv:property:%s:sessions", session.PropertyID), session.SessionID)
	pipe.SAdd(ctx, fmt.Sprintf("iptv:room:%s:sessions", session.RoomID), session.SessionID)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis pipeline failed: %w", err)
	}

	s.localCache.Store(session.SessionID, session)
	sessionsCreated.WithLabelValues(session.TenantID, session.PropertyID, session.DeviceType).Inc()
	sessionsActive.WithLabelValues(session.TenantID, session.PropertyID).Inc()
	return nil
}

func (s *SessionStore) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	if cached, ok := s.localCache.Load(sessionID); ok {
		return cached.(*Session), nil
	}
	key := fmt.Sprintf("iptv:session:%s", sessionID)
	data, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	if err != nil {
		return nil, err
	}
	var session Session
	json.Unmarshal([]byte(data), &session)
	s.localCache.Store(sessionID, &session)
	return &session, nil
}

func (s *SessionStore) GetActiveSessionCount(ctx context.Context, tenantID, propertyID, roomID string) (int, error) {
	members, err := s.redis.SMembers(ctx, fmt.Sprintf("iptv:room:%s:sessions", roomID)).Result()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, sid := range members {
		session, err := s.GetSession(ctx, sid)
		if err != nil {
			continue
		}
		if session.Status == "active" {
			count++
		}
	}
	return count, nil
}

func (s *SessionStore) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("iptv:session:%s", sessionID)
	return s.redis.HSet(ctx, key, "last_activity", time.Now().UTC().Format(time.RFC3339)).Err()
}

func (s *SessionStore) TerminateSession(ctx context.Context, sessionID string, reason string) error {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	session.Status = "terminated"
	pipe := s.redis.Pipeline()
	pipe.Del(ctx, fmt.Sprintf("iptv:session:%s", sessionID))
	pipe.SRem(ctx, fmt.Sprintf("iptv:tenant:%s:sessions", session.TenantID), sessionID)
	pipe.SRem(ctx, fmt.Sprintf("iptv:property:%s:sessions", session.PropertyID), sessionID)
	pipe.SRem(ctx, fmt.Sprintf("iptv:room:%s:sessions", session.RoomID), sessionID)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}
	s.localCache.Delete(sessionID)
	sessionsActive.WithLabelValues(session.TenantID, session.PropertyID).Dec()
	return nil
}

type JWTClaims struct {
	SessionID    string   `json:"sid"`
	TenantID     string   `json:"tid"`
	PropertyID   string   `json:"pid"`
	RoomID       string   `json:"rid"`
	GuestID      string   `json:"gid"`
	Entitlements []string `json:"ent"`
	DeviceType   string   `json:"dev"`
	jwt.RegisteredClaims
}

func generateSessionToken(session *Session, secret []byte) (string, error) {
	claims := JWTClaims{
		SessionID:    session.SessionID,
		TenantID:     session.TenantID,
		PropertyID:   session.PropertyID,
		RoomID:       session.RoomID,
		GuestID:      session.GuestID,
		Entitlements: session.Entitlements,
		DeviceType:   session.DeviceType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(session.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    ServiceName,
			Subject:   session.SessionID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func validateSessionToken(tokenString string, secret []byte) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

type TenantRateLimiter struct {
	limiters *sync.Map
	rate     rate.Limit
	burst    int
}

func NewTenantRateLimiter(r rate.Limit, burst int) *TenantRateLimiter {
	return &TenantRateLimiter{limiters: &sync.Map{}, rate: r, burst: burst}
}

func (rl *TenantRateLimiter) GetLimiter(tenantID string) *rate.Limiter {
	if limiter, ok := rl.limiters.Load(tenantID); ok {
		return limiter.(*rate.Limiter)
	}
	newLimiter := rate.NewLimiter(rl.rate, rl.burst)
	actual, _ := rl.limiters.LoadOrStore(tenantID, newLimiter)
	return actual.(*rate.Limiter)
}

type Server struct {
	store       *SessionStore
	logger      *zap.Logger
	jwtSecret   []byte
	rateLimiter *TenantRateLimiter
	grpcClient  pb.StreamControllerClient
}

func NewServer(store *SessionStore, logger *zap.Logger) *Server {
	secret := []byte(os.Getenv(JWTSecretEnv))
	if len(secret) == 0 {
		secret = make([]byte, 32)
		rand.Read(secret)
	}
	conn, _ := grpc.Dial(os.Getenv("STREAM_CONTROLLER_ADDR"),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	var streamClient pb.StreamControllerClient
	if conn != nil {
		streamClient = pb.NewStreamControllerClient(conn)
	}
	return &Server{
		store: store, logger: logger, jwtSecret: secret,
		rateLimiter: NewTenantRateLimiter(100, 200),
		grpcClient:  streamClient,
	}
}

func (s *Server) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", s.healthCheck)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	v1 := router.Group("/api/v1/iptv")
	v1.Use(s.tenantMiddleware())
	v1.Use(s.rateLimitMiddleware())
	v1.Use(s.authMiddleware())
	v1.Use(s.observabilityMiddleware())
	v1.POST("/sessions", s.createSession)
	v1.GET("/sessions/:session_id", s.getSession)
	v1.DELETE("/sessions/:session_id", s.terminateSession)
	v1.POST("/sessions/:session_id/heartbeat", s.sessionHeartbeat)
	v1.POST("/sessions/:session_id/stream", s.startStream)
	v1.DELETE("/sessions/:session_id/stream", s.stopStream)
	v1.GET("/sessions/:session_id/entitlements", s.getEntitlements)
	v1.POST("/sessions/:session_id/cast", s.startCasting)
	v1.GET("/channels", s.listChannels)
	v1.GET("/channels/:channel_id/stream", s.getChannelStream)
	v1.GET("/vod", s.listVOD)
	v1.GET("/vod/:content_id/stream", s.getVODStream)
	v1.GET("/epg", s.getEPG)
	v1.POST("/emergency/broadcast", s.emergencyBroadcast)
}

func (s *Server) tenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID required"})
			return
		}
		c.Set("tenant_id", tenantID)
		c.Set("property_id", c.GetHeader("X-Property-ID"))
		c.Next()
	}
}

func (s *Server) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.rateLimiter.GetLimiter(c.GetString("tenant_id")).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := validateSessionToken(tokenString, s.jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("session_id", claims.SessionID)
		c.Set("guest_id", claims.GuestID)
		c.Set("room_id", claims.RoomID)
		c.Set("entitlements", claims.Entitlements)
		c.Next()
	}
}

func (s *Server) observabilityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		requestDuration.WithLabelValues(c.Request.Method, c.FullPath(), fmt.Sprintf("%d", c.Writer.Status())).Observe(time.Since(start).Seconds())
	}
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"service": ServiceName, "version": ServiceVersion, "status": "healthy"})
}

type CreateSessionRequest struct {
	GuestID       string                 `json:"guest_id" binding:"required"`
	ReservationID string                 `json:"reservation_id" binding:"required"`
	RoomID        string                 `json:"room_id" binding:"required"`
	DeviceID      string                 `json:"device_id" binding:"required"`
	DeviceType    string                 `json:"device_type" binding:"required"`
	DeviceModel   string                 `json:"device_model"`
	MACAddress    string                 `json:"mac_address"`
	Preferences   map[string]interface{} `json:"preferences,omitempty"`
}

func (s *Server) createSession(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "SessionManager.CreateSession")
	defer span.End()
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID := c.GetString("tenant_id")
	propertyID := c.GetString("property_id")

	session := &Session{
		SessionID:     uuid.New().String(),
		TenantID:      tenantID,
		PropertyID:    propertyID,
		RoomID:        req.RoomID,
		GuestID:       req.GuestID,
		ReservationID: req.ReservationID,
		DeviceID:      req.DeviceID,
		DeviceType:    req.DeviceType,
		DeviceModel:   req.DeviceModel,
		IPAddress:     c.ClientIP(),
		MACAddress:    req.MACAddress,
		Status:        "active",
		Quality:       "1080p",
		MaxConcurrent: MaxConcurrentStreams,
		Entitlements:  []string{"basic", "live_tv", "vod_standard"},
		Preferences:   req.Preferences,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		ExpiresAt:     time.Now().UTC().Add(SessionTTL),
		LastActivity:  time.Now().UTC(),
	}

	if err := s.store.CreateSession(ctx, session); err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	token, _ := generateSessionToken(session, s.jwtSecret)
	c.JSON(http.StatusCreated, gin.H{
		"session_id":     session.SessionID,
		"token":          token,
		"expires_at":     session.ExpiresAt.Format(time.RFC3339),
		"entitlements":   session.Entitlements,
		"max_concurrent": session.MaxConcurrent,
	})
}

func (s *Server) getSession(c *gin.Context) {
	session, err := s.store.GetSession(c.Request.Context(), c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (s *Server) terminateSession(c *gin.Context) {
	if err := s.store.TerminateSession(c.Request.Context(), c.Param("session_id"), c.Query("reason")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "terminated"})
}

func (s *Server) sessionHeartbeat(c *gin.Context) {
	if err := s.store.UpdateSessionActivity(c.Request.Context(), c.Param("session_id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "active"})
}

type StartStreamRequest struct {
	ContentID   string `json:"content_id" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	Quality     string `json:"quality"`
}

func (s *Server) startStream(c *gin.Context) {
	ctx := c.Request.Context()
	var req StartStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session, err := s.store.GetSession(ctx, c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if !hasEntitlement(session.Entitlements, req.ContentType) {
		streamErrors.WithLabelValues(session.TenantID, "entitlement_denied").Inc()
		c.JSON(http.StatusForbidden, gin.H{"error": "content not entitled"})
		return
	}
	if s.grpcClient != nil {
		streamResp, err := s.grpcClient.GetStreamURL(ctx, &pb.StreamRequest{
			SessionId:   session.SessionID,
			ContentId:   req.ContentID,
			ContentType: req.ContentType,
			Quality:     req.Quality,
			TenantId:    session.TenantID,
			PropertyId:  session.PropertyID,
			DeviceType:  session.DeviceType,
		})
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stream service unavailable"})
			return
		}
		streamStarts.WithLabelValues(session.TenantID, req.ContentType, req.Quality).Inc()
		c.JSON(http.StatusOK, gin.H{
			"stream_url":      streamResp.StreamUrl,
			"manifest_url":    streamResp.ManifestUrl,
			"drm_license_url": streamResp.DrmLicenseUrl,
		})
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "stream controller unavailable"})
}

func (s *Server) stopStream(c *gin.Context) {
	if s.grpcClient != nil {
		s.grpcClient.ReleaseStream(c.Request.Context(), &pb.ReleaseRequest{SessionId: c.Param("session_id")})
	}
	c.JSON(http.StatusOK, gin.H{"status": "stream_released"})
}

func (s *Server) getEntitlements(c *gin.Context) {
	session, err := s.store.GetSession(c.Request.Context(), c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entitlements": session.Entitlements, "max_concurrent": session.MaxConcurrent})
}

func (s *Server) startCasting(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "casting_initiated", "cast_session_id": uuid.New().String()})
}
func (s *Server) listChannels(c *gin.Context)      { c.JSON(http.StatusOK, gin.H{"channels": []interface{}{}}) }
func (s *Server) getChannelStream(c *gin.Context)  { c.JSON(http.StatusOK, gin.H{"stream_url": ""}) }
func (s *Server) listVOD(c *gin.Context)           { c.JSON(http.StatusOK, gin.H{"content": []interface{}{}}) }
func (s *Server) getVODStream(c *gin.Context)      { c.JSON(http.StatusOK, gin.H{"stream_url": ""}) }
func (s *Server) getEPG(c *gin.Context)            { c.JSON(http.StatusOK, gin.H{"epg": []interface{}{}}) }

type EmergencyBroadcastRequest struct {
	Level       string   `json:"level" binding:"required"`
	Message     string   `json:"message" binding:"required"`
	PropertyIDs []string `json:"property_ids" binding:"required"`
}

func (s *Server) emergencyBroadcast(c *gin.Context) {
	var req EmergencyBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.logger.Warn("EMERGENCY BROADCAST", zap.String("level", req.Level), zap.Strings("properties", req.PropertyIDs))
	c.JSON(http.StatusAccepted, gin.H{"status": "broadcast_initiated", "affected": len(req.PropertyIDs)})
}

func hasEntitlement(entitlements []string, required string) bool {
	for _, e := range entitlements {
		if e == required || e == "all" {
			return true
		}
	}
	return false
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	store := NewSessionStore(os.Getenv("REDIS_ADDR"), logger)
	server := NewServer(store, logger)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	server.RegisterRoutes(router)
	srv := &http.Server{Addr: ":8081", Handler: router}
	go srv.ListenAndServe()
	logger.Info("IPTV Session Manager started", zap.String("addr", ":8081"))
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	logger.Info("server exited")
}
