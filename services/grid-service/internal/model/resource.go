package model

import "time"

const (
	DeviceTypeCompute = "COMPUTE"
	DeviceTypeMobile  = "MOBILE"
	DeviceTypeIoT     = "IOT"
)

const (
	StatusAvailable = "AVAILABLE"
	StatusBusy      = "BUSY"
	StatusInactive  = "INACTIVE"
)

type Resource struct {
	ID            string
	ResourceID    string
	Name          string
	DeviceType    string
	IP            string
	CPUCores      int32
	MemoryMB      int64
	MPICapable    bool
	Capabilities  []string
	Status        string
	LastHeartbeat time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
