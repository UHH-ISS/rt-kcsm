package structure

import "sync"

type RiskLevel float32

const (
	HighRisk   RiskLevel = 1.5
	MediumRisk RiskLevel = 1
	LowRisk    RiskLevel = 0.5
)

type HostRisk struct {
	IpAddress string  `json:"ip_address"`
	RiskLevel float32 `json:"risk_level"`
}

// Risk score of host
type HostRiskManager struct {
	defaultRiskLevel RiskLevel
	mutex            sync.RWMutex
	riskLevels       map[IPAddress]RiskLevel
}

func NewHostRiskManager(defaultRiskLevel RiskLevel) HostRiskManager {
	return HostRiskManager{
		mutex:            sync.RWMutex{},
		riskLevels:       map[IPAddress]RiskLevel{},
		defaultRiskLevel: defaultRiskLevel,
	}
}

func (h *HostRiskManager) AddHostRiskLevel(address IPAddress, riskLevel RiskLevel) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.riskLevels[address] = riskLevel
}

func (h *HostRiskManager) DeleteHostRiskLevel(address IPAddress) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.riskLevels, address)
}

func (h *HostRiskManager) GetHostRiskLevel(address IPAddress) RiskLevel {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	if riskLevel, ok := h.riskLevels[address]; ok {
		return riskLevel
	} else {
		return h.defaultRiskLevel
	}
}

func (h *HostRiskManager) GetHosts() map[IPAddress]RiskLevel {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.riskLevels
}

var HostManager = NewHostRiskManager(MediumRisk)
