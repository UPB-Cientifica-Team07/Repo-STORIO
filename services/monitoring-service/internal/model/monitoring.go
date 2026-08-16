package model

import "time"

type Metric struct {
	ComponentID       string
	ComponentName     string
	CPUUsage          float64
	MemoryUsage       float64
	ActiveConnections int64
	TotalRequests     int64
	Timestamp         time.Time
}

type ServiceStatus struct {
	ComponentID   string
	ComponentName string
	Status        string
	Message       string
	LastUpdated   time.Time
}