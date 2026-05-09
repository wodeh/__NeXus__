package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	pb "github.com/nexushc/iptv-proto/gen/go/stream/v1"
)

const ServiceName = "stream-controller"
var tracer = otel.Tracer(ServiceName)

var (
	streamURLsGenerated = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "stream_urls_generated_total", Help: "Total stream URLs generated"},
		[]string{"tenant_id", "content_type", "cdn"},
	)
	drmLicensesIssued = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "drm_licenses_issued_total", Help: "Total DRM licenses issued"},
		[]string{"tenant_id", "drm_type"},
	)
	cdnLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "cdn_latency_seconds", Help: "CDN selection latency"},
		[]string{"cdn", "region"},
	)
)

func init() {
	prometheus.MustRegister(streamURLsGenerated, drmLicensesIssued, cdnLatency)
}

type StreamController struct {
	pb.UnimplementedStreamControllerServer
	redis       *redis.Client
	logger      *zap.Logger
	cdnSelector *CDNSelector
	drmService  *DRMService
}

func NewStreamController(redisAddr string, logger *zap.Logger) *StreamController {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr, Password: os.Getenv("REDIS_PASSWORD"),
		DB: 0, PoolSize: 50, MinIdleConns: 10,
	})
	return &StreamController{
		redis:       rdb,
		logger:      logger,
		cdnSelector: NewCDNSelector(logger),
		drmService:  NewDRMService(logger),
	}
}

func (sc *StreamController) GetStreamURL(ctx context.Context, req *pb.StreamRequest) (*pb.StreamResponse, error) {
	ctx, span := tracer.Start(ctx, "StreamController.GetStreamURL")
	defer span.End()
	span.SetAttributes(
		attribute.String("content.id", req.ContentId),
		attribute.String("content.type", req.ContentType),
	)

	cdn, latency := sc.cdnSelector.Select(req.TenantId, req.PropertyId, req.DeviceType)
	cdnLatency.WithLabelValues(cdn.Name, cdn.Region).Observe(latency.Seconds())

	signedURL, err := sc.generateSignedURL(ctx, req, cdn)
	if err != nil {
		return nil, fmt.Errorf("url generation failed: %w", err)
	}

	var drmLicenseURL string
	if req.ContentType != "hotel_content" {
		drmLicenseURL = fmt.Sprintf("%s/drm/license?session=%s&content=%s",
			os.Getenv("DRM_SERVICE_URL"), req.SessionId, req.ContentId)
	}

	streamURLsGenerated.WithLabelValues(req.TenantId, req.ContentType, cdn.Name).Inc()

	return &pb.StreamResponse{
		StreamUrl:      signedURL,
		ManifestUrl:    signedURL + "/manifest.m3u8",
		DrmLicenseUrl:  drmLicenseURL,
		QualityOptions: []string{"4k", "1080p", "720p", "480p", "360p"},
		ExpiresIn:      3600,
	}, nil
}

func (sc *StreamController) ReleaseStream(ctx context.Context, req *pb.ReleaseRequest) (*pb.ReleaseResponse, error) {
	sc.redis.Del(ctx, fmt.Sprintf("stream:url:%s", req.SessionId))
	return &pb.ReleaseResponse{Success: true}, nil
}

func (sc *StreamController) generateSignedURL(ctx context.Context, req *pb.StreamRequest, cdn *CDN) (string, error) {
	var contentPath string
	switch req.ContentType {
	case "live_channel":
		contentPath = fmt.Sprintf("/live/%s", req.ContentId)
	case "vod":
		contentPath = fmt.Sprintf("/vod/%s", req.ContentId)
	case "catchup":
		contentPath = fmt.Sprintf("/catchup/%s", req.ContentId)
	default:
		contentPath = fmt.Sprintf("/content/%s", req.ContentId)
	}

	token := uuid.New().String()
	sc.redis.Set(ctx, fmt.Sprintf("stream:token:%s", token), req.SessionId, 2*time.Hour)

	return fmt.Sprintf("%s%s?token=%s&quality=%s&device=%s",
		cdn.BaseURL, contentPath, token, req.Quality, req.DeviceType), nil
}

type CDN struct {
	Name     string
	BaseURL  string
	Region   string
	Latency  time.Duration
	Capacity float64
	Healthy  bool
	Priority int
}

type CDNSelector struct {
	cdns   []*CDN
	logger *zap.Logger
	mu     sync.RWMutex
}

func NewCDNSelector(logger *zap.Logger) *CDNSelector {
	cs := &CDNSelector{
		cdns: []*CDN{
			{Name: "cloudflare", BaseURL: "https://cf.nhc-cdn.com", Region: "global", Priority: 1},
			{Name: "fastly", BaseURL: "https://fastly.nhc-cdn.com", Region: "global", Priority: 2},
			{Name: "aws-cloudfront", BaseURL: "https://cf.nhc-aws.com", Region: "global", Priority: 3},
		},
		logger: logger,
	}
	go cs.healthCheckLoop()
	return cs
}

func (cs *CDNSelector) Select(tenantID, propertyID, deviceType string) (*CDN, time.Duration) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	var best *CDN
	bestScore := -1.0
	for _, cdn := range cs.cdns {
		if !cdn.Healthy {
			continue
		}
		score := (1 - cdn.Capacity) * 100 / float64(cdn.Priority)
		if score > bestScore {
			bestScore = score
			best = cdn
		}
	}
	if best == nil {
		best = cs.cdns[0]
	}
	return best, best.Latency
}

func (cs *CDNSelector) healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		for _, cdn := range cs.cdns {
			start := time.Now()
			resp, err := http.Get(cdn.BaseURL + "/health")
			latency := time.Since(start)
			cs.mu.Lock()
			if err != nil || resp.StatusCode != 200 {
				cdn.Healthy = false
			} else {
				cdn.Healthy = true
				cdn.Latency = latency
			}
			cs.mu.Unlock()
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}

type DRMService struct {
	logger       *zap.Logger
	widevineKey  []byte
	fairplayKey  []byte
	playreadyKey []byte
}

func NewDRMService(logger *zap.Logger) *DRMService {
	return &DRMService{
		logger:       logger,
		widevineKey:  []byte(os.Getenv("WIDEVINE_KEY")),
		fairplayKey:  []byte(os.Getenv("FAIRPLAY_KEY")),
		playreadyKey: []byte(os.Getenv("PLAYREADY_KEY")),
	}
}

func (ds *DRMService) IssueLicense(ctx context.Context, req *pb.DRMLicenseRequest) (*pb.DRMLicenseResponse, error) {
	var license []byte
	var drmType string
	switch req.DrmType {
	case "widevine":
		license = ds.generateWidevineLicense(req)
		drmType = "widevine"
	case "fairplay":
		license = ds.generateFairPlayLicense(req)
		drmType = "fairplay"
	case "playready":
		license = ds.generatePlayReadyLicense(req)
		drmType = "playready"
	default:
		return nil, fmt.Errorf("unsupported DRM type: %s", req.DrmType)
	}
	drmLicensesIssued.WithLabelValues(req.TenantId, drmType).Inc()
	return &pb.DRMLicenseResponse{
		License:   license,
		DrmType:   drmType,
		ExpiresAt: time.Now().Add(4 * time.Hour).Unix(),
	}, nil
}

func (ds *DRMService) generateWidevineLicense(req *pb.DRMLicenseRequest) []byte {
	data := fmt.Sprintf("widevine:%s:%s:%d", req.ContentId, req.SessionId, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

func (ds *DRMService) generateFairPlayLicense(req *pb.DRMLicenseRequest) []byte {
	data := fmt.Sprintf("fairplay:%s:%s:%d", req.ContentId, req.SessionId, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

func (ds *DRMService) generatePlayReadyLicense(req *pb.DRMLicenseRequest) []byte {
	data := fmt.Sprintf("playready:%s:%s:%d", req.ContentId, req.SessionId, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

func (ds *DRMService) RegisterRoutes(router *gin.Engine) {
	router.POST("/drm/license", ds.handleLicenseRequest)
}

func (ds *DRMService) handleLicenseRequest(c *gin.Context) {
	var req pb.DRMLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := ds.IssueLicense(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/octet-stream", resp.License)
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	sc := NewStreamController(os.Getenv("REDIS_ADDR"), logger)

	grpcServer := grpc.NewServer()
	pb.RegisterStreamControllerServer(grpcServer, sc)

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	go func() {
		logger.Info("gRPC server starting", zap.String("addr", ":9090"))
		grpcServer.Serve(lis)
	}()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": ServiceName, "status": "healthy"})
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	sc.drmService.RegisterRoutes(router)

	srv := &http.Server{Addr: ":8082", Handler: router}
	go srv.ListenAndServe()
	logger.Info("Stream Controller HTTP started", zap.String("addr", ":8082"))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	logger.Info("server exited")
}
