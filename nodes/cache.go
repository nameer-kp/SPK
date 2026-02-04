package nodes

import (
	"fmt"
	"sync"
	"time"

	"github.com/nameer-kp/atem/pkg/node"
)

// Cache entry with TTL
type cacheEntry struct {
	value     any
	expiresAt time.Time
	createdAt time.Time
}

// Global cache store
var (
	cacheStore   = make(map[string]cacheEntry)
	cacheStoreMu sync.RWMutex
)

// CacheNode provides key-value caching with TTL
type CacheNode struct{}

func NewCacheNode() *CacheNode {
	return &CacheNode{}
}

func (n *CacheNode) Name() string {
	return "cache"
}

func (n *CacheNode) Description() string {
	return "Key-value cache with TTL support"
}

func (n *CacheNode) ConfigSchema() map[string]any {
	return map[string]any{
		"operation": "string", // get, set, delete, exists, clear, list, stats
		"key":       "string", // cache key
		"value":     "any",    // value to cache
		"ttl":       "string", // time-to-live (e.g., "5m", "1h")
		"default":   "any",    // default value for get
		"prefix":    "string", // prefix for list/clear operations
	}
}

func (n *CacheNode) Execute(ctx node.Context, config map[string]any) (node.Result, error) {
	operation, _ := config["operation"].(string)
	if operation == "" {
		operation = "get"
	}

	switch operation {
	case "get":
		key, _ := config["key"].(string)
		if key == "" {
			return node.Result{}, fmt.Errorf("key is required for get")
		}

		cacheStoreMu.RLock()
		entry, exists := cacheStore[key]
		cacheStoreMu.RUnlock()

		if !exists || (!entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt)) {
			// Expired or not found
			if exists {
				// Clean up expired
				cacheStoreMu.Lock()
				delete(cacheStore, key)
				cacheStoreMu.Unlock()
			}

			defaultVal := config["default"]
			return node.Result{
				Success: true,
				Data: map[string]any{
					"key":    key,
					"value":  defaultVal,
					"hit":    false,
					"exists": false,
				},
				Status: "miss",
			}, nil
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"key":       key,
				"value":     entry.value,
				"hit":       true,
				"exists":    true,
				"createdAt": entry.createdAt.Format(time.RFC3339),
				"expiresAt": formatExpiry(entry.expiresAt),
			},
			Status: "hit",
		}, nil

	case "set":
		key, _ := config["key"].(string)
		if key == "" {
			return node.Result{}, fmt.Errorf("key is required for set")
		}

		value := config["value"]
		ttlStr, _ := config["ttl"].(string)

		entry := cacheEntry{
			value:     value,
			createdAt: time.Now(),
		}

		if ttlStr != "" {
			ttl, err := time.ParseDuration(ttlStr)
			if err != nil {
				return node.Result{}, fmt.Errorf("invalid TTL: %w", err)
			}
			entry.expiresAt = time.Now().Add(ttl)
		}

		cacheStoreMu.Lock()
		cacheStore[key] = entry
		cacheStoreMu.Unlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"key":       key,
				"ttl":       ttlStr,
				"expiresAt": formatExpiry(entry.expiresAt),
			},
			Status: fmt.Sprintf("cached %s", key),
		}, nil

	case "delete":
		key, _ := config["key"].(string)
		if key == "" {
			return node.Result{}, fmt.Errorf("key is required for delete")
		}

		cacheStoreMu.Lock()
		_, existed := cacheStore[key]
		delete(cacheStore, key)
		cacheStoreMu.Unlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"key":     key,
				"deleted": existed,
			},
			Status: fmt.Sprintf("deleted %s", key),
		}, nil

	case "exists":
		key, _ := config["key"].(string)
		if key == "" {
			return node.Result{}, fmt.Errorf("key is required for exists")
		}

		cacheStoreMu.RLock()
		entry, exists := cacheStore[key]
		cacheStoreMu.RUnlock()

		// Check expiry
		if exists && !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
			exists = false
		}

		return node.Result{
			Success: true,
			Data: map[string]any{
				"key":    key,
				"exists": exists,
			},
			Status: fmt.Sprintf("exists: %v", exists),
		}, nil

	case "clear":
		prefix, _ := config["prefix"].(string)

		cacheStoreMu.Lock()
		count := 0
		if prefix == "" {
			count = len(cacheStore)
			cacheStore = make(map[string]cacheEntry)
		} else {
			for key := range cacheStore {
				if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
					delete(cacheStore, key)
					count++
				}
			}
		}
		cacheStoreMu.Unlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"cleared": count,
				"prefix":  prefix,
			},
			Status: fmt.Sprintf("cleared %d entries", count),
		}, nil

	case "list":
		prefix, _ := config["prefix"].(string)

		cacheStoreMu.RLock()
		keys := make([]string, 0)
		now := time.Now()
		for key, entry := range cacheStore {
			// Skip expired
			if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
				continue
			}
			if prefix == "" || (len(key) >= len(prefix) && key[:len(prefix)] == prefix) {
				keys = append(keys, key)
			}
		}
		cacheStoreMu.RUnlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"keys":   keys,
				"count":  len(keys),
				"prefix": prefix,
			},
			Status: fmt.Sprintf("%d keys", len(keys)),
		}, nil

	case "stats":
		cacheStoreMu.RLock()
		total := len(cacheStore)
		expired := 0
		now := time.Now()
		for _, entry := range cacheStore {
			if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
				expired++
			}
		}
		cacheStoreMu.RUnlock()

		return node.Result{
			Success: true,
			Data: map[string]any{
				"total":   total,
				"active":  total - expired,
				"expired": expired,
			},
			Status: fmt.Sprintf("%d entries", total),
		}, nil

	default:
		return node.Result{}, fmt.Errorf("unknown operation: %s", operation)
	}
}

func formatExpiry(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format(time.RFC3339)
}

// ResetCache clears the cache store (for testing)
func ResetCache() {
	cacheStoreMu.Lock()
	cacheStore = make(map[string]cacheEntry)
	cacheStoreMu.Unlock()
}
