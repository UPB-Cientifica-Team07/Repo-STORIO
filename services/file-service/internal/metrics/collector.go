package metrics

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Collector struct {
	mu sync.Mutex

	lastCPUTime uint64
	lastTime    time.Time

	storageDir string
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewCollector(storageDir string) *Collector {
	c := &Collector{
		lastTime:   time.Now(),
		storageDir: storageDir,
	}

	c.lastCPUTime = readProcessCPUTime()

	return c
}

// =====================================
// CPU REAL DEL PROCESO
// =====================================

func (c *Collector) CPUUsage() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	currentCPUTime := readProcessCPUTime()

	elapsed := now.Sub(
		c.lastTime,
	).Seconds()

	if elapsed <= 0 {
		return 0
	}

	if currentCPUTime < c.lastCPUTime {
		c.lastCPUTime = currentCPUTime
		c.lastTime = now

		return 0
	}

	cpuDelta := currentCPUTime - c.lastCPUTime

	// En Linux x86/x86_64 normalmente
	// CLK_TCK es 100.
	const clockTicksPerSecond = 100.0

	cpuSeconds :=
		float64(cpuDelta) /
			clockTicksPerSecond

	cpuUsage :=
		(cpuSeconds / elapsed) *
			100.0

	c.lastCPUTime = currentCPUTime
	c.lastTime = now

	if cpuUsage < 0 {
		return 0
	}

	return cpuUsage
}

// =====================================
// MEMORIA RSS REAL DEL PROCESO
// =====================================

func (c *Collector) MemoryUsageMB() float64 {
	file, err := os.Open(
		"/proc/self/status",
	)

	if err != nil {
		return 0
	}

	defer file.Close()

	scanner := bufio.NewScanner(
		file,
	)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(
			line,
			"VmRSS:",
		) {
			continue
		}

		fields := strings.Fields(
			line,
		)

		if len(fields) < 2 {
			return 0
		}

		valueKB, err := strconv.ParseInt(
			fields[1],
			10,
			64,
		)

		if err != nil {
			return 0
		}

		return float64(valueKB) / 1024.0
	}

	return 0
}

// =====================================
// STORAGE REAL DEL FILE SERVICE
// =====================================
//
// Se contabilizan dos zonas:
//
// 1. <shared-storage>/files
//    Compatibilidad con archivos legacy.
//
// 2. <shared-storage>/homes
//    Nuevo modelo:
//    homes/<usuario>/documentos
//    homes/<usuario>/imagenes
//    homes/<usuario>/videos
//    homes/<usuario>/sync
//
// metadata.json NO se contabiliza porque
// solamente representa metadata interna.
//
// La métrica representa contenido físico
// de usuario almacenado por File Service.
// =====================================

func (c *Collector) StorageUsageBytes() (
	int64,
	error,
) {
	storageAreas := []string{
		filepath.Join(
			c.storageDir,
			"files",
		),
		filepath.Join(
			c.storageDir,
			"homes",
		),
	}

	var totalBytes int64

	for _, storageArea := range storageAreas {
		bytes, err := directorySize(
			storageArea,
		)

		if err != nil {
			return 0, err
		}

		totalBytes += bytes
	}

	return totalBytes, nil
}

// =====================================
// CALCULAR TAMAÑO DE DIRECTORIO
// =====================================

func directorySize(
	directory string,
) (
	int64,
	error,
) {
	info, err := os.Stat(
		directory,
	)

	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}

		return 0,
			fmt.Errorf(
				"no se pudo acceder al almacenamiento %s: %w",
				directory,
				err,
			)
	}

	if !info.IsDir() {
		return 0,
			fmt.Errorf(
				"la ruta de almacenamiento no es un directorio: %s",
				directory,
			)
	}

	var totalBytes int64

	err = filepath.WalkDir(
		directory,
		func(
			path string,
			entry fs.DirEntry,
			walkErr error,
		) error {
			if walkErr != nil {
				return walkErr
			}

			if entry.IsDir() {
				return nil
			}

			info, err := entry.Info()

			if err != nil {
				return err
			}

			// Solo contenido físico regular.
			// Symlinks, sockets, dispositivos,
			// pipes, etc. no se contabilizan.
			if !info.Mode().IsRegular() {
				return nil
			}

			totalBytes += info.Size()

			return nil
		},
	)

	if err != nil {
		return 0,
			fmt.Errorf(
				"error calculando uso de almacenamiento en %s: %w",
				directory,
				err,
			)
	}

	return totalBytes, nil
}

// =====================================
// OBTENER TODAS LAS MÉTRICAS
// =====================================

func (c *Collector) Collect() (
	float64,
	float64,
	int64,
	error,
) {
	cpu := c.CPUUsage()

	memory := c.MemoryUsageMB()

	storage, err :=
		c.StorageUsageBytes()

	if err != nil {
		return 0,
			0,
			0,
			err
	}

	if cpu < 0 {
		return 0,
			0,
			0,
			fmt.Errorf(
				"valor de CPU inválido",
			)
	}

	if memory < 0 {
		return 0,
			0,
			0,
			fmt.Errorf(
				"valor de memoria inválido",
			)
	}

	if storage < 0 {
		return 0,
			0,
			0,
			fmt.Errorf(
				"valor de almacenamiento inválido",
			)
	}

	return cpu,
		memory,
		storage,
		nil
}

// =====================================
// LEER TIEMPO CPU DEL PROCESO
// =====================================

func readProcessCPUTime() uint64 {
	data, err := os.ReadFile(
		"/proc/self/stat",
	)

	if err != nil {
		return 0
	}

	fields := strings.Fields(
		string(data),
	)

	// /proc/[pid]/stat:
	//
	// campo 14 = utime
	// campo 15 = stime
	//
	// índices Go:
	// 13 = utime
	// 14 = stime

	if len(fields) <= 14 {
		return 0
	}

	userTime, err := strconv.ParseUint(
		fields[13],
		10,
		64,
	)

	if err != nil {
		return 0
	}

	systemTime, err := strconv.ParseUint(
		fields[14],
		10,
		64,
	)

	if err != nil {
		return 0
	}

	return userTime + systemTime
}
