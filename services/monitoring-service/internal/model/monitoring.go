package model

import "time"

// =====================================
// MÉTRICA
// =====================================

type Metric struct {
	ComponentID       string
	ComponentName     string
	CPUUsage          float64
	MemoryUsage       float64
	StorageUsage      int64
	ActiveConnections int64
	TotalRequests     int64
	Timestamp         time.Time
}

// =====================================
// ESTADO DE SERVICIO
// =====================================

type ServiceStatus struct {
	ComponentID   string
	ComponentName string
	Status        string
	Message       string
	LastUpdated   time.Time
}
