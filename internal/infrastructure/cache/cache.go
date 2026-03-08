package cache

import (
	"github.com/rickseven/logiq/internal/domain"
	"sync"
)

var (
	explainCache = make(map[string]domain.ExplainResult)
	doctorCache  *domain.DoctorResult
	mu           sync.RWMutex
)

// GetExplain retrieves cached explanations
func GetExplain(key string) (domain.ExplainResult, bool) {
	mu.RLock()
	defer mu.RUnlock()
	res, found := explainCache[key]
	return res, found
}

// SetExplain writes to cache
func SetExplain(key string, res domain.ExplainResult) {
	mu.Lock()
	defer mu.Unlock()
	explainCache[key] = res
}

// GetDoctor gets cached doctor diagnostic
func GetDoctor() (*domain.DoctorResult, bool) {
	mu.RLock()
	defer mu.RUnlock()
	if doctorCache != nil {
		return doctorCache, true
	}
	return nil, false
}

// SetDoctor caches doctor results
func SetDoctor(res domain.DoctorResult) {
	mu.Lock()
	defer mu.Unlock()
	doctorCache = &res
}
