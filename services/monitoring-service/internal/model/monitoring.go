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

// =====================================
// NODO HPC
// =====================================

type HpcNode struct {
	NodeID      string
	Hostname    string
	Status      string
	CPUCores    int32
	MemoryMB    int64
	IP          string
	Location    string
	CPUUsage    float64
	MemoryUsage float64
	LastUpdated time.Time
}

// =====================================
// RESUMEN HPC
// =====================================

type HpcSummary struct {
	TotalJobs          int64
	PendingJobs        int64
	RunningJobs        int64
	CompletedJobs      int64
	FailedJobs         int64
	CancelledJobs      int64
	AverageExecutionMs float64

	TotalNodes     int64
	AvailableNodes int64
	BusyNodes      int64
	InactiveNodes  int64

	Timestamp time.Time
}
