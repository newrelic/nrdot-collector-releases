package topnprocessfilter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

const (
	typeDouble = "Double"

	defaultCPUThresholdPercent    = 5.0
	defaultMemoryThresholdPercent = 10.0
	defaultRetentionMinutes       = 60
	defaultStoragePath            = "/var/lib/nrdot-collector/topnprocess.db"

	minCPUThresholdPercent    = 1.0
	minMemoryThresholdPercent = 2.0
	maxCPUThresholdPercent    = 10.0
	maxMemoryThresholdPercent = 20.0
	dynamicUpdateIntervalSecs = 60

	cpuScalingFactor    = 0.5
	memoryScalingFactor = 0.3
)

const DEBUG_FILTER = true

type TrackedProcess struct {
	PID                      int       `json:"pid"`
	Name                     string    `json:"name"`
	CommandLine              string    `json:"command_line"`
	Owner                    string    `json:"owner"`
	FirstSeen                time.Time `json:"first_seen"`
	LastExceeded             time.Time `json:"last_exceeded"`
	MaxCPUUtilization        float64   `json:"max_cpu_utilization"`
	MaxMemoryUtilization     float64   `json:"max_memory_utilization"`
	CurrentCPUUtilization    float64   `json:"current_cpu_utilization"`
	CurrentMemoryUtilization float64   `json:"current_memory_utilization"`
}

type processorImp struct {
	logger                   *zap.Logger
	config                   *Config
	nextConsumer             consumer.Metrics
	trackedProcesses         map[int]*TrackedProcess
	mu                       sync.RWMutex
	storage                  ProcessStateStorage
	lastPersistenceOp        time.Time
	persistenceEnabled       bool
	systemCPUUtilization     float64
	systemMemoryUtilization  float64
	lastThresholdUpdate      time.Time
	dynamicThresholdsEnabled bool
}

func newProcessor(logger *zap.Logger, config *Config, nextConsumer consumer.Metrics) (*processorImp, error) {
	p := &processorImp{
		logger:                   logger,
		config:                   config,
		nextConsumer:             nextConsumer,
		trackedProcesses:         make(map[int]*TrackedProcess),
		persistenceEnabled:       config.StoragePath != "",
		systemCPUUtilization:     0.0,
		systemMemoryUtilization:  0.0,
		dynamicThresholdsEnabled: config.EnableDynamicThresholds,
		lastThresholdUpdate:      time.Now(),
	}

	if p.persistenceEnabled {
		storageDir := filepath.Dir(config.StoragePath)
		if err := createDirectoryIfNotExists(storageDir); err != nil {
			logger.Warn("Failed to create storage directory", zap.String("path", storageDir), zap.Error(err))
			p.persistenceEnabled = false
		} else {
			storage, err := NewFileStorage(config.StoragePath)
			if err != nil {
				logger.Warn("Failed to initialize storage", zap.Error(err))
				p.persistenceEnabled = false
			} else {
				p.storage = storage
				if err := p.loadTrackedProcesses(); err != nil {
					logger.Warn("Failed to load tracked processes", zap.Error(err))
				}
			}
		}
	}

	return p, nil
}

func (p *processorImp) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	filteredMetrics, err := p.processMetrics(ctx, md)
	if err != nil {
		return err
	}

	if p.persistenceEnabled && time.Since(p.lastPersistenceOp) > time.Minute {
		if err := p.persistTrackedProcesses(); err != nil {
			p.logger.Warn("Failed to persist tracked processes", zap.Error(err))
		}
		p.lastPersistenceOp = time.Now()
	}

	return p.nextConsumer.ConsumeMetrics(ctx, filteredMetrics)
}

func (p *processorImp) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	filteredMetrics := pmetric.NewMetrics()

	resourceMetrics := md.ResourceMetrics()
	for i := 0; i < resourceMetrics.Len(); i++ {
		resourceMetric := resourceMetrics.At(i)

		resource := resourceMetric.Resource()
		pid, hasPID := getProcessPID(resource)

		if !hasPID {
			resourceMetric.CopyTo(filteredMetrics.ResourceMetrics().AppendEmpty())
			continue
		}

		if p.shouldIncludeProcess(resource, resourceMetric, pid) {
			resourceMetric.CopyTo(filteredMetrics.ResourceMetrics().AppendEmpty())
		}
	}

	p.cleanupExpiredProcesses()

	return filteredMetrics, nil
}

func (p *processorImp) shouldIncludeProcess(resource pcommon.Resource, rm pmetric.ResourceMetrics, pid int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	processNameVal, hasName := resource.Attributes().Get("process.executable.name")
	processName := ""
	if hasName {
		processName = processNameVal.AsString()
	}

	commandLineVal, hasCmd := resource.Attributes().Get("process.command_line")
	commandLine := ""
	if hasCmd {
		commandLine = commandLineVal.AsString()
	}

	ownerVal, hasOwner := resource.Attributes().Get("process.owner")
	owner := ""
	if hasOwner {
		owner = ownerVal.AsString()
	}

	trackedProcess, exists := p.trackedProcesses[pid]

	cpuUtil, memUtil := p.extractUtilizationMetrics(rm)

	if exists {
		oldCurrentCPU := trackedProcess.CurrentCPUUtilization
		oldCurrentMem := trackedProcess.CurrentMemoryUtilization
		trackedProcess.CurrentCPUUtilization = cpuUtil
		trackedProcess.CurrentMemoryUtilization = memUtil

		if cpuUtil > trackedProcess.MaxCPUUtilization {
			trackedProcess.MaxCPUUtilization = cpuUtil
		}
		if memUtil > trackedProcess.MaxMemoryUtilization {
			trackedProcess.MaxMemoryUtilization = memUtil
		}

		oldLastExceeded := trackedProcess.LastExceeded
		exceeds := false
		var reasons []string

		if cpuUtil >= p.config.CPUThresholdPercent {
			reasons = append(reasons, fmt.Sprintf("CPU %.1f%% >= %.1f%%", cpuUtil, p.config.CPUThresholdPercent))
			exceeds = true
		}

		if memUtil >= p.config.MemoryThresholdPercent {
			reasons = append(reasons, fmt.Sprintf("Memory %.1f%% >= %.1f%%", memUtil, p.config.MemoryThresholdPercent))
			exceeds = true
		}

		if exceeds {
			trackedProcess.LastExceeded = time.Now()

			if DEBUG_FILTER {
				debugMsg := fmt.Sprintf("[DEBUG] PID %d (%s) metrics updated and threshold exceeded: CPU %.1f%% → %.1f%% (max: %.1f%%), Memory %.1f%% → %.1f%% (max: %.1f%%), Reasons: %v",
					pid, processName, oldCurrentCPU, cpuUtil, trackedProcess.MaxCPUUtilization, oldCurrentMem, memUtil, trackedProcess.MaxMemoryUtilization, reasons)
				p.logger.Info(debugMsg)
				fmt.Fprintln(os.Stderr, debugMsg)

				timeDiff := trackedProcess.LastExceeded.Sub(oldLastExceeded).Seconds()
				debugTimeMsg := fmt.Sprintf("[DEBUG] PID %d last_exceeded updated: was %.1fs ago", pid, timeDiff)
				p.logger.Info(debugTimeMsg)
				fmt.Fprintln(os.Stderr, debugTimeMsg)
			}
		} else if DEBUG_FILTER {
			debugMsg := fmt.Sprintf("[DEBUG] PID %d (%s) metrics updated but BELOW thresholds: CPU %.1f%% → %.1f%% (max: %.1f%%), Memory %.1f%% → %.1f%% (max: %.1f%%)",
				pid, processName, oldCurrentCPU, cpuUtil, trackedProcess.MaxCPUUtilization, oldCurrentMem, memUtil, trackedProcess.MaxMemoryUtilization)
			p.logger.Info(debugMsg)
			fmt.Fprintln(os.Stderr, debugMsg)
		}

		if DEBUG_FILTER {
			fmt.Fprintf(os.Stderr, "[DEBUG] Including already tracked process: PID %d (%s)\n", pid, processName)
		}
		return true
	}

	exceeds := false
	var reasons []string

	if cpuUtil >= p.config.CPUThresholdPercent {
		reasons = append(reasons, fmt.Sprintf("CPU %.1f%% >= %.1f%%", cpuUtil, p.config.CPUThresholdPercent))
		exceeds = true
	}

	if memUtil >= p.config.MemoryThresholdPercent {
		reasons = append(reasons, fmt.Sprintf("Memory %.1f%% >= %.1f%%", memUtil, p.config.MemoryThresholdPercent))
		exceeds = true
	}

	if exceeds {
		now := time.Now()
		p.trackedProcesses[pid] = &TrackedProcess{
			PID:                      pid,
			Name:                     processName,
			CommandLine:              commandLine,
			Owner:                    owner,
			FirstSeen:                now,
			LastExceeded:             now,
			MaxCPUUtilization:        cpuUtil,
			MaxMemoryUtilization:     memUtil,
			CurrentCPUUtilization:    cpuUtil,
			CurrentMemoryUtilization: memUtil,
		}

		debugMsg := fmt.Sprintf("Started tracking process: PID %d, Name: %s, CPU: %.1f%% (max: %.1f%%), Memory: %.1f%% (max: %.1f%%), Reasons: %v",
			pid, processName, cpuUtil, cpuUtil, memUtil, memUtil, reasons)
		p.logger.Info(debugMsg)

		if DEBUG_FILTER {
			fmt.Fprintln(os.Stderr, "[DEBUG] "+debugMsg)
		}

		return true
	}

	if DEBUG_FILTER {
		debugMsg := fmt.Sprintf("[DEBUG] PID %d (%s) EXCLUDED: below thresholds with CPU %.1f%% (< %.1f%%), Memory %.1f%% (< %.1f%%)",
			pid, processName, cpuUtil, p.config.CPUThresholdPercent, memUtil, p.config.MemoryThresholdPercent)
		p.logger.Info(debugMsg)
		fmt.Fprintln(os.Stderr, debugMsg)
	}

	return false
}

func (p *processorImp) extractUtilizationMetrics(rm pmetric.ResourceMetrics) (float64, float64) {
	var cpuUtilization, memoryUtilization float64
	cpuFound, memFound := false, false

	for i := 0; i < rm.ScopeMetrics().Len(); i++ {
		scopeMetrics := rm.ScopeMetrics().At(i)

		for j := 0; j < scopeMetrics.Metrics().Len(); j++ {
			metric := scopeMetrics.Metrics().At(j)

			switch metric.Name() {
			case "process.cpu.utilization":
				if !cpuFound && metric.Gauge().DataPoints().Len() > 0 {
					cpuFound = true
					cpuUtilization = 0

					for k := 0; k < metric.Gauge().DataPoints().Len(); k++ {
						dataPoint := metric.Gauge().DataPoints().At(k)
						cpuUtilization += dataPoint.DoubleValue() * 100
					}
				}
			case "process.memory.utilization":
				if !memFound && metric.Gauge().DataPoints().Len() > 0 {
					memFound = true
					dataPoint := metric.Gauge().DataPoints().At(0)
					memoryUtilization = dataPoint.DoubleValue() * 100
				}
			}

			if cpuFound && memFound {
				break
			}
		}

		if cpuFound && memFound {
			break
		}
	}

	return cpuUtilization, memoryUtilization
}

func (p *processorImp) cleanupExpiredProcesses() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if DEBUG_FILTER {
		fmt.Fprintf(os.Stderr, "\n[DEBUG] Checking for expired processes among %d tracked processes\n", len(p.trackedProcesses))
	}

	expirationTime := time.Now().Add(-time.Duration(p.config.RetentionMinutes) * time.Minute)

	for pid, process := range p.trackedProcesses {
		if process.LastExceeded.Before(expirationTime) {
			timeSinceExceeded := time.Since(process.LastExceeded).Seconds()

			debugMsg := fmt.Sprintf("Removing expired process from tracking: PID %d, Name: %s, Last exceeded %.1f seconds ago (retention period: %d minutes)",
				pid, process.Name, timeSinceExceeded, p.config.RetentionMinutes)
			p.logger.Info(debugMsg)

			if DEBUG_FILTER {
				fmt.Fprintln(os.Stderr, "[DEBUG] "+debugMsg)
			}

			delete(p.trackedProcesses, pid)
		} else if DEBUG_FILTER {
			timeSinceExceeded := time.Since(process.LastExceeded).Seconds()
			fmt.Fprintf(os.Stderr, "[DEBUG] PID %d (%s) still within retention period: last exceeded %.1f seconds ago (retention: %d minutes)\n",
				pid, process.Name, timeSinceExceeded, p.config.RetentionMinutes)
		}
	}
}

func getProcessPID(resource pcommon.Resource) (int, bool) {
	pidAttr, exists := resource.Attributes().Get("process.pid")
	if !exists {
		return 0, false
	}

	// Check that the attribute is of the integer type
	if pidAttr.Type() != pcommon.ValueTypeInt {
		return 0, false
	}

	return int(pidAttr.Int()), true
}

func (p *processorImp) loadTrackedProcesses() error {
	if p.storage == nil {
		return nil
	}

	processes, err := p.storage.Load()
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.trackedProcesses = processes
	p.logger.Info("Loaded tracked processes from storage", zap.Int("count", len(processes)))

	return nil
}

func (p *processorImp) persistTrackedProcesses() error {
	if p.storage == nil {
		return nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	err := p.storage.Save(p.trackedProcesses)
	if err != nil {
		return err
	}

	p.logger.Debug("Persisted tracked processes to storage", zap.Int("count", len(p.trackedProcesses)))
	return nil
}

func (p *processorImp) Shutdown(ctx context.Context) error {
	if p.persistenceEnabled && p.storage != nil {
		if err := p.persistTrackedProcesses(); err != nil {
			p.logger.Warn("Failed to persist tracked processes during shutdown", zap.Error(err))
		}

		if err := p.storage.Close(); err != nil {
			p.logger.Warn("Failed to close storage during shutdown", zap.Error(err))
		}
	}
	return nil
}

func (p *processorImp) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (p *processorImp) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}
