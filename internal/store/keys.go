package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID         string     `json:"id"`
	Key        string     `json:"key"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type keyStore struct {
	Keys []APIKey `json:"keys"`
}

type KeyStore struct {
	mu       sync.RWMutex
	filePath string
	data     keyStore
}

func NewKeyStore(filePath string) (*KeyStore, error) {
	ks := &KeyStore{filePath: filePath}
	if err := ks.load(); err != nil {
		return nil, err
	}
	return ks, nil
}

func (ks *KeyStore) Validate(key string) bool {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	for _, k := range ks.data.Keys {
		if k.Key == key {
			return true
		}
	}
	return false
}

func (ks *KeyStore) Touch(key string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	now := time.Now()
	for i, k := range ks.data.Keys {
		if k.Key == key {
			ks.data.Keys[i].LastUsedAt = &now
			break
		}
	}
	ks.save()
}

func (ks *KeyStore) Create(name string) (*APIKey, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	k := APIKey{
		ID:        uuid.New().String(),
		Key:       "sk-sg-" + hex.EncodeToString(raw),
		Name:      name,
		CreatedAt: time.Now(),
	}

	ks.data.Keys = append(ks.data.Keys, k)
	if err := ks.save(); err != nil {
		return nil, err
	}
	return &k, nil
}

func (ks *KeyStore) List() []APIKey {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	out := make([]APIKey, len(ks.data.Keys))
	copy(out, ks.data.Keys)
	return out
}

func (ks *KeyStore) Revoke(id string) bool {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	for i, k := range ks.data.Keys {
		if k.ID == id {
			ks.data.Keys = append(ks.data.Keys[:i], ks.data.Keys[i+1:]...)
			ks.save()
			return true
		}
	}
	return false
}

func (ks *KeyStore) load() error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	data, err := os.ReadFile(ks.filePath)
	if os.IsNotExist(err) {
		ks.data = keyStore{}
		return ks.save()
	}
	if err != nil {
		return fmt.Errorf("read keys file: %w", err)
	}
	return json.Unmarshal(data, &ks.data)
}

func (ks *KeyStore) save() error {
	data, err := json.MarshalIndent(ks.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal keys: %w", err)
	}
	return os.WriteFile(ks.filePath, data, 0600)
}
