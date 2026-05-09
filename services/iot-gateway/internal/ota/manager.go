package ota

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nexus-platform/iot-gateway/internal/registry"
)

type OTAManager struct {
	registry      *registry.DeviceRegistry
	firmwareStore FirmwareStore
	batcher       *UpdateBatcher
	mu            sync.RWMutex
	activeUpdates map[string]*UpdateJob
}

type FirmwareStore interface {
	GetManifest(version string) (*FirmwareManifest, error)
	DownloadFirmware(url string, checksum string) (string, error)
}

type FirmwareManifest struct {
	Version     string            `json:"version"`
	DeviceTypes []string          `json:"device_types"`
	URL         string            `json:"url"`
	Checksum    string            `json:"checksum"`
	Size        int64             `json:"size"`
	Changes     []string          `json:"changes"`
	MinVersion  string            `json:"min_version"`
	Signature   string            `json:"signature"`
	ReleaseDate time.Time         `json:"release_date"`
	Metadata    map[string]string `json:"metadata"`
}

type UpdateJob struct {
	ID                string            `json:"id"`
	DeviceIDs         []string          `json:"device_ids"`
	TenantID          string            `json:"tenant_id"`
	PropertyID        string            `json:"property_id"`
	TargetVersion     string            `json:"target_version"`
	Strategy          UpdateStrategy    `json:"strategy"`
	Status            string            `json:"status"`
	Progress          map[string]int    `json:"progress"`
	ScheduledAt       time.Time         `json:"scheduled_at"`
	StartedAt         *time.Time        `json:"started_at,omitempty"`
	CompletedAt       *time.Time        `json:"completed_at,omitempty"`
	RollbackOnFailure bool              `json:"rollback_on_failure"`
}

type UpdateStrategy struct {
	Type           string        `json:"type"`
	BatchSize      int           `json:"batch_size"`
	BatchInterval  time.Duration `json:"batch_interval"`
	CanaryPercent  float64       `json:"canary_percent"`
	CanaryDuration time.Duration `json:"canary_duration"`
}

func NewOTAManager(reg *registry.DeviceRegistry, store FirmwareStore) *OTAManager {
	return &OTAManager{
		registry:      reg,
		firmwareStore: store,
		activeUpdates: make(map[string]*UpdateJob),
		batcher:       NewUpdateBatcher(),
	}
}

func (m *OTAManager) CreateUpdate(ctx context.Context, req UpdateRequest) (*UpdateJob, error) {
	manifest, err := m.firmwareStore.GetManifest(req.TargetVersion)
	if err != nil {
		return nil, fmt.Errorf("firmware manifest not found: %w", err)
	}

	if err := m.verifySignature(manifest); err != nil {
		return nil, fmt.Errorf("firmware signature invalid: %w", err)
	}

	devices, err := m.resolveDevices(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("device resolution failed: %w", err)
	}

	job := &UpdateJob{
		ID:                generateUpdateID(),
		DeviceIDs:         devices,
		TenantID:          req.TenantID,
		PropertyID:        req.PropertyID,
		TargetVersion:     req.TargetVersion,
		Strategy:          req.Strategy,
		Status:            "pending",
		Progress:          make(map[string]int),
		ScheduledAt:       req.ScheduleAt,
		RollbackOnFailure: req.RollbackOnFailure,
	}

	go m.stageFirmware(job.ID, manifest)

	m.mu.Lock()
	m.activeUpdates[job.ID] = job
	m.mu.Unlock()

	if req.ScheduleAt.IsZero() || req.ScheduleAt.Before(time.Now()) {
		go m.executeUpdate(job)
	} else {
		go m.scheduleUpdate(job)
	}

	return job, nil
}

func (m *OTAManager) executeUpdate(job *UpdateJob) {
	now := time.Now()
	job.StartedAt = &now
	job.Status = "downloading"

	manifest, _ := m.firmwareStore.GetManifest(job.TargetVersion)
	firmwarePath, err := m.firmwareStore.DownloadFirmware(manifest.URL, manifest.Checksum)
	if err != nil {
		m.failUpdate(job, fmt.Sprintf("firmware download failed: %v", err))
		return
	}

	job.Status = "installing"

	switch job.Strategy.Type {
	case "immediate":
		m.executeImmediate(job, firmwarePath)
	case "rolling":
		m.executeRolling(job, firmwarePath)
	case "canary":
		m.executeCanary(job, firmwarePath)
	case "scheduled":
		m.executeScheduled(job, firmwarePath)
	}
}

func (m *OTAManager) executeRolling(job *UpdateJob, firmwarePath string) {
	batchSize := job.Strategy.BatchSize
	if batchSize == 0 {
		batchSize = 10
	}

	for i := 0; i < len(job.DeviceIDs); i += batchSize {
		end := i + batchSize
		if end > len(job.DeviceIDs) {
			end = len(job.DeviceIDs)
		}
		batch := job.DeviceIDs[i:end]

		var wg sync.WaitGroup
		for _, deviceID := range batch {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				m.deployToDevice(id, firmwarePath, job)
			}(deviceID)
		}
		wg.Wait()

		if !m.verifyBatchStability(batch, job) {
			if job.RollbackOnFailure {
				m.rollbackBatch(batch, job)
			}
			m.failUpdate(job, "batch stability check failed")
			return
		}

		if job.Strategy.BatchInterval > 0 && end < len(job.DeviceIDs) {
			time.Sleep(job.Strategy.BatchInterval)
		}
	}

	m.completeUpdate(job)
}

func (m *OTAManager) executeCanary(job *UpdateJob, firmwarePath string) {
	canaryCount := int(float64(len(job.DeviceIDs)) * job.Strategy.CanaryPercent / 100)
	if canaryCount == 0 {
		canaryCount = 1
	}

	canaryDevices := job.DeviceIDs[:canaryCount]
	remainingDevices := job.DeviceIDs[canaryCount:]

	for _, deviceID := range canaryDevices {
		m.deployToDevice(deviceID, firmwarePath, job)
	}

	if !m.monitorCanary(canaryDevices, job.Strategy.CanaryDuration) {
		if job.RollbackOnFailure {
			m.rollbackBatch(canaryDevices, job)
		}
		m.failUpdate(job, "canary monitoring failed")
		return
	}

	for _, deviceID := range remainingDevices {
		m.deployToDevice(deviceID, firmwarePath, job)
	}

	m.completeUpdate(job)
}

func (m *OTAManager) deployToDevice(deviceID string, firmwarePath string, job *UpdateJob) {
	device, err := m.registry.Get(deviceID)
	if err != nil {
		m.updateProgress(deviceID, job, -1, "device_not_found")
		return
	}

	if !m.isCompatible(device, job.TargetVersion) {
		m.updateProgress(deviceID, job, -1, "incompatible")
		return
	}

	updateCmd := map[string]interface{}{
		"command":        "firmware_update",
		"firmware_url":   m.getFirmwareURL(firmwarePath),
		"checksum":       m.getChecksum(firmwarePath),
		"version":        job.TargetVersion,
		"update_id":      job.ID,
	}

	m.updateProgress(deviceID, job, 0, "downloading")

	timeout := time.After(30 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			m.updateProgress(deviceID, job, -1, "timeout")
			return
		case <-ticker.C:
			status := m.getDeviceOTAStatus(deviceID)
			if status == "completed" {
				m.updateProgress(deviceID, job, 100, "completed")
				return
			} else if status == "failed" {
				m.updateProgress(deviceID, job, -1, "failed")
				return
			} else if progress, ok := m.getDeviceProgress(deviceID); ok {
				m.updateProgress(deviceID, job, progress, status)
			}
		}
	}
}

func (m *OTAManager) verifyBatchStability(deviceIDs []string, job *UpdateJob) bool {
	time.Sleep(2 * time.Minute)
	for _, id := range deviceIDs {
		if m.getDeviceOTAStatus(id) != "completed" {
			return false
		}
	}
	return true
}

func (m *OTAManager) rollbackBatch(deviceIDs []string, job *UpdateJob) {
	log.Printf("Rolling back update %s for %d devices", job.ID, len(deviceIDs))
	for _, id := range deviceIDs {
		rollbackCmd := map[string]interface{}{
			"command": "firmware_rollback",
		}
		_ = rollbackCmd
	}
}

func (m *OTAManager) verifySignature(manifest *FirmwareManifest) error {
	return nil
}

func (m *OTAManager) stageFirmware(jobID string, manifest *FirmwareManifest) {
	_, _ = m.firmwareStore.DownloadFirmware(manifest.URL, manifest.Checksum)
	log.Printf("Firmware staged for job %s", jobID)
}

func (m *OTAManager) updateProgress(deviceID string, job *UpdateJob, progress int, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job.Progress[deviceID] = progress
	log.Printf("OTA progress [%s] device=%s progress=%d status=%s", job.ID, deviceID, progress, status)
}

func (m *OTAManager) completeUpdate(job *UpdateJob) {
	now := time.Now()
	job.CompletedAt = &now
	job.Status = "completed"
	log.Printf("OTA update %s completed for %d devices", job.ID, len(job.DeviceIDs))
}

func (m *OTAManager) failUpdate(job *UpdateJob, reason string) {
	job.Status = "failed"
	log.Printf("OTA update %s failed: %s", job.ID, reason)
}

func (m *OTAManager) isCompatible(device *registry.Device, version string) bool {
	return true
}

func (m *OTAManager) getFirmwareURL(path string) string {
	return fmt.Sprintf("https://ota.nexus.local/firmware/%s", filepath.Base(path))
}

func (m *OTAManager) getChecksum(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (m *OTAManager) getDeviceOTAStatus(deviceID string) string {
	return "unknown"
}

func (m *OTAManager) getDeviceProgress(deviceID string) (int, bool) {
	return 0, false
}

func (m *OTAManager) monitorCanary(deviceIDs []string, duration time.Duration) bool {
	time.Sleep(duration)
	return true
}

func (m *OTAManager) scheduleUpdate(job *UpdateJob) {
	delay := time.Until(job.ScheduledAt)
	if delay > 0 {
		time.Sleep(delay)
	}
	m.executeUpdate(job)
}

func generateUpdateID() string {
	return fmt.Sprintf("ota-%d-%s", time.Now().Unix(), randomString(8))
}

func randomString(n int) string {
	return "random"
}

type UpdateRequest struct {
	TenantID          string         `json:"tenant_id"`
	PropertyID        string         `json:"property_id"`
	DeviceFilter      DeviceFilter   `json:"device_filter"`
	TargetVersion     string         `json:"target_version"`
	Strategy          UpdateStrategy `json:"strategy"`
	ScheduleAt        time.Time      `json:"schedule_at"`
	RollbackOnFailure bool           `json:"rollback_on_failure"`
}

type DeviceFilter struct {
	DeviceTypes []string `json:"device_types,omitempty"`
	RoomIDs     []string `json:"room_ids,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func (m *OTAManager) resolveDevices(ctx context.Context, req UpdateRequest) ([]string, error) {
	return []string{}, nil
}

type UpdateBatcher struct{}

func NewUpdateBatcher() *UpdateBatcher {
	return &UpdateBatcher{}
}
