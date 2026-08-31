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

// =====================================
// REGLA DE ALERTA
// =====================================

type AlertRule struct {
	ID            string
	Name          string
	Metric        string
	Operator      string
	Threshold     float64
	ComponentName string
	Enabled       bool
}

// =====================================
// ALERTA
// =====================================

type Alert struct {
	ID            string
	RuleID        string
	RuleName      string
	ComponentID   string
	ComponentName string
	Metric        string
	CurrentValue  float64
	Threshold     float64
	Message       string

	CreatedAt time.Time

	Active bool

	ResolvedAt *time.Time
}
