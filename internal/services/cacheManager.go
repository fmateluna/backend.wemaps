package services

import (
	"fmt"
	"sync"
	"time"
)

// CacheEntry representa una entrada en el caché
type CacheEntry struct {
	Data       interface{}
	Expiration time.Time
}

// CacheManager maneja el caché en memoria
type CacheManager struct {
	cache map[string]CacheEntry
	mu    sync.RWMutex
}

var (
	// instance es la única instancia del CacheManager (singleton)
	instance *CacheManager
	// once asegura que la inicialización se realice solo una vez
	once sync.Once
)

// GetCacheManager retorna la instancia singleton del CacheManager
func GetCacheManager() *CacheManager {
	once.Do(func() {
		instance = &CacheManager{
			cache: make(map[string]CacheEntry),
		}
	})
	return instance
}

func (cm *CacheManager) cacheKey(nameReport, hash string) string {
	return fmt.Sprintf("%s:%s", nameReport, hash)
}

// addressCacheKey genera la clave para AddressInfo en caché
func (cm *CacheManager) addressCacheKey(userID int) string {
	return fmt.Sprintf("address:%d", userID)
}

// reportSummaryCacheKey genera la clave para ReportSummary en caché
func (cm *CacheManager) reportSummaryCacheKey(userID int) string {
	return fmt.Sprintf("report_summary:%d", userID)
}

// Set agrega o actualiza una entrada en el caché con TTL
func (cm *CacheManager) Set(key string, value interface{}, ttl time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.cache[key] = CacheEntry{
		Data:       value,
		Expiration: time.Now().Add(ttl),
	}
}

// Update actualiza una entrada existente en el caché con un nuevo valor y TTL
func (cm *CacheManager) Update(key string, value interface{}, ttl time.Duration) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.cache[key]; !exists {
		return false // La clave no existe, no se actualiza
	}

	cm.cache[key] = CacheEntry{
		Data:       value,
		Expiration: time.Now().Add(ttl),
	}
	return true
}

// Get obtiene una entrada del caché si no ha expirado
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	entry, exists := cm.cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.Expiration) {
		// Eliminar entrada expirada
		cm.mu.RUnlock()
		cm.Delete(key)
		cm.mu.RLock()
		return nil, false
	}

	return entry.Data, true
}

// Delete elimina una entrada específica del caché
func (cm *CacheManager) Delete(key string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.cache, key)
}
