package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/iptv-middleware/transcode-engine/internal/config"
	"github.com/nexus-platform/iptv-middleware/transcode-engine/internal/gpu"
	"github.com/nexus-platform/iptv-middleware/transcode-engine/internal/health"
	"github.com/nexus-platform/iptv-middleware/transcode-engine/internal/metrics"
	"github.com/nexus-platform/iptv-middleware/transcode-engine/internal/pipeline"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("transcode-engine")

type TranscodeJob struct {
	ID              string           `json:"id"`
	StreamID        string           `json:"stream_id"`
	InputURL        string           `json:"input_url"`
	OutputBasePath  string           `json:"output_base_path"`
	ProfileLadder   []BitrateProfile `json:"profile_ladder"`
	Codec           string           `json:"codec"`
	GPUDevice       int              `json:"gpu_device"`
	AudioTracks     []AudioConfig    `json:"audio_tracks"`
	SubtitleTracks  []SubtitleConfig `json:"subtitle_tracks"`
	DrmConfig       *DRMConfig        `json:"drm_config,omitempty"`
	SCTE35Config    *SCTE35Config     `json:"scte35_config,omitempty"`
	DVRWindowHours  int              `json:"dvr_window_hours"`
	CreatedAt       time.Time         `json:"created_at"`
	Priority        int              `json:"priority"`
	TenantID        string           `json:"tenant_id"`
	Region          string           `json:"region"`
}

type BitrateProfile struct {
	Name         string `json:"name"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	VideoBitrate int    `json:"video_bitrate"`
	AudioBitrate int    `json:"audio_bitrate"`
	FrameRate    int    `json:"frame_rate"`
	GOPSize      int    `json:"gop_size"`
}

type AudioConfig struct {
	Index    int    `json:"index"`
	Language string `json:"language"`
	Codec    string `json:"codec"`
	Bitrate  int    `json:"bitrate"`
	Channels int    `json:"channels"`
}

type SubtitleConfig struct {
	Index    int    `json:"index"`
	Language string `json:"language"`
	Format   string `json:"format"`
}

type DRMConfig struct {
	Provider    string `json:"provider"`
	LicenseURL  string `json:"license_url"`
	KeyRotation int    `json:"key_rotation"`
	ContentID   string `json:"content_id"`
}

type SCTE35Config struct {
	Enabled        bool   `json:"enabled"`
	MarkerType     string `json:"marker_type"`
	AdServerURL    string `json:"ad_server_url"`
	PrerollSeconds int    `json:"preroll_seconds"`
}

type TranscodeEngine struct {
	config       *config.Config
	gpuManager   *gpu.Manager
	jobQueue     chan TranscodeJob
	activeJobs   map[string]*JobContext
	mu           sync.RWMutex
	metrics      *metrics.Collector
	healthServer *health.Server
	cancelFunc   context.CancelFunc
	wg           sync.WaitGroup
}

type JobContext struct {
	Job       TranscodeJob
	Cmd       *exec.Cmd
	Cancel    context.CancelFunc
	StartTime time.Time
	GPUDevice int
}

func NewTranscodeEngine(cfg *config.Config) (*TranscodeEngine, error) {
	gpuMgr, err := gpu.NewManager(cfg.GPU.Count, cfg.GPU.MemoryThreshold)
	if err != nil {
		return nil, fmt.Errorf("gpu manager init failed: %w", err)
	}

	return &TranscodeEngine{
		config:       cfg,
		gpuManager:   gpuMgr,
		jobQueue:     make(chan TranscodeJob, cfg.QueueSize),
		activeJobs:   make(map[string]*JobContext),
		metrics:      metrics.NewCollector(cfg.Metrics),
		healthServer: health.NewServer(cfg.HealthPort),
	}, nil
}

func (e *TranscodeEngine) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	e.cancelFunc = cancel

	if err := e.healthServer.Start(); err != nil {
		return fmt.Errorf("health server failed: %w", err)
	}

	e.gpuManager.StartMonitoring(ctx)

	for i := 0; i < e.config.Workers; i++ {
		e.wg.Add(1)
		go e.worker(ctx, i)
	}

	e.wg.Add(1)
	go e.metricsExporter(ctx)
	e.wg.Add(1)
	go e.jobRecovery(ctx)

	log.Printf("TranscodeEngine started with %d workers, %d GPUs", e.config.Workers, e.config.GPU.Count)
	return nil
}

func (e *TranscodeEngine) worker(ctx context.Context, id int) {
	defer e.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-e.jobQueue:
			e.processJob(ctx, job)
		}
	}
}

func (e *TranscodeEngine) processJob(ctx context.Context, job TranscodeJob) {
	ctx, span := tracer.Start(ctx, "process-job",
		trace.WithAttributes(
			attribute.String("job.id", job.ID),
			attribute.String("stream.id", job.StreamID),
			attribute.String("tenant.id", job.TenantID),
		))
	defer span.End()

	gpuDevice, err := e.gpuManager.Acquire(ctx, job.Priority)
	if err != nil {
		log.Printf("Failed to acquire GPU for job %s: %v", job.ID, err)
		e.metrics.IncrementCounter("transcode_gpu_wait_failed", job.TenantID)
		return
	}
	defer e.gpuManager.Release(gpuDevice)

	job.GPUDevice = gpuDevice

	cmd, cancel, err := e.buildFFmpegCommand(ctx, job)
	if err != nil {
		log.Printf("Failed to build FFmpeg command for job %s: %v", job.ID, err)
		e.metrics.IncrementCounter("transcode_build_failed", job.TenantID)
		return
	}

	jobCtx := &JobContext{
		Job:       job,
		Cmd:       cmd,
		Cancel:    cancel,
		StartTime: time.Now(),
		GPUDevice: gpuDevice,
	}

	e.mu.Lock()
	e.activeJobs[job.ID] = jobCtx
	e.mu.Unlock()

	e.metrics.IncrementCounter("transcode_started", job.TenantID)
	e.metrics.GaugeSet("transcode_active_jobs", float64(len(e.activeJobs)), job.TenantID)

	startTime := time.Now()
	err = cmd.Run()
	duration := time.Since(startTime)

	e.mu.Lock()
	delete(e.activeJobs, job.ID)
	e.mu.Unlock()

	if err != nil {
		if ctx.Err() == context.Canceled {
			log.Printf("Job %s cancelled", job.ID)
			e.metrics.IncrementCounter("transcode_cancelled", job.TenantID)
		} else {
			log.Printf("Job %s failed: %v", job.ID, err)
			e.metrics.IncrementCounter("transcode_failed", job.TenantID)
			e.metrics.HistogramObserve("transcode_duration_failed", duration.Seconds(), job.TenantID)
		}
		return
	}

	e.metrics.IncrementCounter("transcode_completed", job.TenantID)
	e.metrics.HistogramObserve("transcode_duration_success", duration.Seconds(), job.TenantID)
	e.metrics.GaugeSet("transcode_active_jobs", float64(len(e.activeJobs)), job.TenantID)

	log.Printf("Job %s completed in %v", job.ID, duration)
}

func (e *TranscodeEngine) buildFFmpegCommand(ctx context.Context, job TranscodeJob) (*exec.Cmd, context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(ctx)

	args := []string{
		"-hide_banner",
		"-y",
		"-hwaccel", "cuda",
		"-hwaccel_device", strconv.Itoa(job.GPUDevice),
		"-i", job.InputURL,
	}

	for i, profile := range job.ProfileLadder {
		args = append(args,
			"-filter:v:"+strconv.Itoa(i), fmt.Sprintf("scale_cuda=%d:%d", profile.Width, profile.Height),
			"-c:v:"+strconv.Itoa(i), job.Codec,
			"-b:v:"+strconv.Itoa(i), fmt.Sprintf("%dk", profile.VideoBitrate),
			"-maxrate:v:"+strconv.Itoa(i), fmt.Sprintf("%dk", int(float64(profile.VideoBitrate)*1.5)),
			"-bufsize:v:"+strconv.Itoa(i), fmt.Sprintf("%dk", profile.VideoBitrate*2),
			"-r:v:"+strconv.Itoa(i), strconv.Itoa(profile.FrameRate),
			"-g:v:"+strconv.Itoa(i), strconv.Itoa(profile.GOPSize),
			"-preset", "p4",
			"-rc", "vbr_hq",
			"-spatial_aq", "1",
			"-temporal_aq", "1",
		)

		profilePath := fmt.Sprintf("%s/%s", job.OutputBasePath, profile.Name)
		args = append(args, "-f", "hls",
			"-hls_time", "6",
			"-hls_playlist_type", "event",
			"-hls_flags", "independent_segments+iframes_only",
			"-hls_segment_type", "fmp4",
			"-master_pl_name", fmt.Sprintf("master_%s.m3u8", profile.Name),
			profilePath+"/stream.m3u8",
		)
	}

	for _, audio := range job.AudioTracks {
		args = append(args,
			"-map", fmt.Sprintf("0:a:%d", audio.Index),
			"-c:a", audio.Codec,
			"-b:a", fmt.Sprintf("%dk", audio.Bitrate),
			"-ac", strconv.Itoa(audio.Channels),
			"-metadata:s:a", fmt.Sprintf("language=%s", audio.Language),
		)
	}

	if job.DrmConfig != nil {
		args = append(args, e.buildDRMFlags(job.DrmConfig)...)
	}

	if job.SCTE35Config != nil && job.SCTE35Config.Enabled {
		args = append(args,
			"-scte35", "1",
			"-scte35_out", "1",
		)
	}

	if job.DVRWindowHours > 0 {
		args = append(args,
			"-hls_list_size", strconv.Itoa(job.DVRWindowHours*600),
			"-hls_delete_threshold", "1",
		)
	}

	args = append(args,
		"-f", "hls",
		"-master_pl_name", "master.m3u8",
		job.OutputBasePath+"/master.m3u8",
	)

	args = append(args,
		"-progress", "unix:///tmp/ffmpeg-progress-"+job.ID,
		"-stats_period", "5",
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Env = append(os.Environ(),
		"CUDA_VISIBLE_DEVICES="+strconv.Itoa(job.GPUDevice),
		"NVENC_PRESET="+e.config.NVEncPreset,
	)

	cmd.Stdout = &ffmpegLogWriter{jobID: job.ID, level: "info"}
	cmd.Stderr = &ffmpegLogWriter{jobID: job.ID, level: "error"}

	return cmd, cancel, nil
}

func (e *TranscodeEngine) buildDRMFlags(cfg *DRMConfig) []string {
	flags := []string{}
	switch cfg.Provider {
	case "widevine":
		flags = append(flags,
			"-encryption_scheme", "cenc",
			"-key_info_file", fmt.Sprintf("/tmp/keys/%s.widevine", cfg.ContentID),
		)
	case "playready":
		flags = append(flags,
			"-encryption_scheme", "cbc1",
			"-key_info_file", fmt.Sprintf("/tmp/keys/%s.playready", cfg.ContentID),
		)
	case "fairplay":
		flags = append(flags,
			"-encryption_scheme", "cbcs",
			"-hls_key_info_file", fmt.Sprintf("/tmp/keys/%s.fairplay", cfg.ContentID),
		)
	}
	return flags
}

func (e *TranscodeEngine) SubmitJob(job TranscodeJob) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	job.CreatedAt = time.Now()

	select {
	case e.jobQueue <- job:
		e.metrics.IncrementCounter("transcode_queued", job.TenantID)
		return nil
	default:
		return fmt.Errorf("job queue full (size: %d)", len(e.jobQueue))
	}
}

func (e *TranscodeEngine) CancelJob(jobID string) error {
	e.mu.Lock()
	jobCtx, exists := e.activeJobs[jobID]
	e.mu.Unlock()

	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}
	jobCtx.Cancel()
	return nil
}

func (e *TranscodeEngine) GetJobStatus(jobID string) (*JobStatus, error) {
	e.mu.RLock()
	jobCtx, exists := e.activeJobs[jobID]
	e.mu.RUnlock()

	if !exists {
		return &JobStatus{ID: jobID, State: "completed_or_unknown"}, nil
	}

	return &JobStatus{
		ID:        jobID,
		State:     "running",
		Progress:  e.getProgress(jobID),
		GPUDevice: jobCtx.GPUDevice,
		Duration:  time.Since(jobCtx.StartTime),
	}, nil
}

func (e *TranscodeEngine) getProgress(jobID string) float64 {
	progressFile := "/tmp/ffmpeg-progress-" + jobID
	data, err := os.ReadFile(progressFile)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "progress=") {
			val := strings.TrimPrefix(line, "progress=")
			if val == "end" {
				return 100
			}
		}
	}
	return 0
}

func (e *TranscodeEngine) metricsExporter(ctx context.Context) {
	defer e.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := e.gpuManager.GetStats()
			e.metrics.GaugeSet("gpu_utilization", stats.AvgUtilization, "")
			e.metrics.GaugeSet("gpu_memory_used", stats.AvgMemoryUsed, "")
			e.metrics.GaugeSet("gpu_temperature", stats.AvgTemperature, "")
		}
	}
}

func (e *TranscodeEngine) jobRecovery(ctx context.Context) {
	defer e.wg.Done()
}

func (e *TranscodeEngine) Shutdown(timeout time.Duration) error {
	log.Println("Shutting down TranscodeEngine...")
	e.cancelFunc()

	e.mu.Lock()
	for _, jobCtx := range e.activeJobs {
		jobCtx.Cancel()
	}
	e.mu.Unlock()

	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Graceful shutdown complete")
	case <-time.After(timeout):
		log.Println("Force shutdown after timeout")
	}

	e.healthServer.Stop()
	return nil
}

type JobStatus struct {
	ID        string        `json:"id"`
	State     string        `json:"state"`
	Progress  float64       `json:"progress"`
	GPUDevice int           `json:"gpu_device"`
	Duration  time.Duration `json:"duration"`
}

type ffmpegLogWriter struct {
	jobID string
	level string
}

func (w *ffmpegLogWriter) Write(p []byte) (n int, err error) {
	log.Printf("[ffmpeg][%s][%s] %s", w.jobID, w.level, string(p))
	return len(p), nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	engine, err := NewTranscodeEngine(cfg)
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := engine.Start(ctx); err != nil {
		log.Fatalf("Failed to start engine: %v", err)
	}

	go func() {
		time.Sleep(5 * time.Second)
		job := TranscodeJob{
			StreamID:       "test-channel-1",
			InputURL:       "udp://239.1.1.1:1234",
			OutputBasePath: "/var/streams/test-channel-1",
			ProfileLadder: []BitrateProfile{
				{Name: "1080p", Width: 1920, Height: 1080, VideoBitrate: 5000, AudioBitrate: 192, FrameRate: 30, GOPSize: 60},
				{Name: "720p", Width: 1280, Height: 720, VideoBitrate: 2500, AudioBitrate: 128, FrameRate: 30, GOPSize: 60},
				{Name: "480p", Width: 854, Height: 480, VideoBitrate: 1000, AudioBitrate: 96, FrameRate: 30, GOPSize: 60},
				{Name: "360p", Width: 640, Height: 360, VideoBitrate: 500, AudioBitrate: 64, FrameRate: 30, GOPSize: 60},
			},
			Codec:          "h264_nvenc",
			AudioTracks:    []AudioConfig{{Index: 0, Language: "eng", Codec: "aac", Bitrate: 192, Channels: 2}},
			DVRWindowHours: 72,
			TenantID:       "demo-tenant",
			Region:         "us-east-1",
		}
		if err := engine.SubmitJob(job); err != nil {
			log.Printf("Failed to submit job: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	if err := engine.Shutdown(30 * time.Second); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}
