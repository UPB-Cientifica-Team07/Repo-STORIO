package monitoring

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SystemMetrics struct {
	mu sync.Mutex

	lastProcessCPU float64
	lastTimestamp  time.Time

	clockTicks float64
}

func NewSystemMetrics() *SystemMetrics {

	return &SystemMetrics{
		lastTimestamp: time.Now(),
		clockTicks:    100,
	}
}

// =====================================
// MEMORIA REAL DEL PROCESO
// =====================================

func (m *SystemMetrics) MemoryUsageMB() float64 {

	var stats runtime.MemStats

	runtime.ReadMemStats(
		&stats,
	)

	return float64(
		stats.Alloc,
	) / 1024 / 1024
}

// =====================================
// CPU REAL DEL PROCESO
// Linux /proc/self/stat
// =====================================

func (m *SystemMetrics) CPUUsagePercent() (
	float64,
	error,
) {

	m.mu.Lock()
	defer m.mu.Unlock()

	file, err :=
		os.Open(
			"/proc/self/stat",
		)

	if err != nil {
		return 0, err
	}

	defer file.Close()

	scanner :=
		bufio.NewScanner(
			file,
		)

	if !scanner.Scan() {
		return 0, errors.New(
			"no se pudo leer /proc/self/stat",
		)
	}

	fields :=
		strings.Fields(
			scanner.Text(),
		)

	if len(fields) < 15 {
		return 0, errors.New(
			"formato inesperado en /proc/self/stat",
		)
	}

	/*
		Campos:
		14 = utime
		15 = stime

		En slice:
		13 y 14
	*/

	utime, err :=
		strconv.ParseFloat(
			fields[13],
			64,
		)

	if err != nil {
		return 0, err
	}

	stime, err :=
		strconv.ParseFloat(
			fields[14],
			64,
		)

	if err != nil {
		return 0, err
	}

	totalProcessCPU :=
		utime + stime

	now :=
		time.Now()

	elapsed :=
		now.Sub(
			m.lastTimestamp,
		).Seconds()

	if elapsed <= 0 {
		return 0, nil
	}

	processDelta :=
		totalProcessCPU -
			m.lastProcessCPU

	cpuSeconds :=
		processDelta /
			m.clockTicks

	cpuPercent :=
		(cpuSeconds / elapsed) *
			100

	m.lastProcessCPU =
		totalProcessCPU

	m.lastTimestamp =
		now

	if cpuPercent < 0 {
		cpuPercent = 0
	}

	return cpuPercent, nil
}

// =====================================
// STORAGE REAL
// =====================================

func StorageUsageBytes(
	path string,
) (int64, error) {

	var total int64

	err :=
		filepath.Walk(
			path,
			func(
				currentPath string,
				info os.FileInfo,
				err error,
			) error {

				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				total +=
					info.Size()

				return nil
			},
		)

	if err != nil {
		return 0, err
	}

	return total, nil
}
