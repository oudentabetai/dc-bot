package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Config struct {
	ServerRoles map[string][]string `json:"server_roles"`
}

type ConfigManager struct {
	FilePath string
	config   Config
	mu       sync.RWMutex
}

func NewConfigManager(path string) *ConfigManager {
	return &ConfigManager{
		FilePath: path,
		config:   Config{ServerRoles: make(map[string][]string)},
	}
}

func (m *ConfigManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	file, err := os.ReadFile(m.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // ファイルがない場合は空のまま進む
		}
		return err
	}

	var raw struct {
		ServerRoles map[string]json.RawMessage `json:"server_roles"`
	}
	if err := json.Unmarshal(file, &raw); err != nil {
		return err
	}

	m.config.ServerRoles = make(map[string][]string)
	for roleID, value := range raw.ServerRoles {
		var serverIDs []string
		if err := json.Unmarshal(value, &serverIDs); err == nil {
			m.config.ServerRoles[roleID] = uniqueStrings(serverIDs)
			continue
		}

		var serverID string
		if err := json.Unmarshal(value, &serverID); err == nil && serverID != "" {
			m.config.ServerRoles[roleID] = []string{serverID}
			continue
		}

		return fmt.Errorf("unsupported server_roles format for role %s", roleID)
	}

	return nil
}

func (m *ConfigManager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.FilePath, data, 0644)
}

func (m *ConfigManager) SetRole(roleID, serverID string) string {
	m.mu.Lock()
	if m.config.ServerRoles == nil {
		m.config.ServerRoles = make(map[string][]string)
	}
	serverIDs := m.config.ServerRoles[roleID]
	for _, existingID := range serverIDs {
		if existingID == serverID {
			m.mu.Unlock()
			return "ロールID " + roleID + " には既にサーバーID " + serverID + " が登録されています。"
		}
	}
	m.config.ServerRoles[roleID] = append(serverIDs, serverID)
	m.mu.Unlock()
	if err := m.Save(); err != nil {
		return "設定の保存に失敗しました: " + err.Error()
	}
	return "ロールID " + roleID + " にサーバーID " + serverID + " を追加しました。"
}

func (m *ConfigManager) RemoveRole(roleID, serverID string) string {
	m.mu.Lock()
	serverIDs, ok := m.config.ServerRoles[roleID]
	if !ok {
		m.mu.Unlock()
		return "ロールID " + roleID + " に対応するサーバーIDが見つかりませんでした。"
	}

	filtered := make([]string, 0, len(serverIDs))
	removed := false
	for _, existingID := range serverIDs {
		if existingID == serverID {
			removed = true
			continue
		}
		filtered = append(filtered, existingID)
	}

	if !removed {
		m.mu.Unlock()
		return "ロールID " + roleID + " にサーバーID " + serverID + " は登録されていません。"
	}

	if len(filtered) == 0 {
		delete(m.config.ServerRoles, roleID)
	} else {
		m.config.ServerRoles[roleID] = filtered
	}
	m.mu.Unlock()
	if err := m.Save(); err != nil {
		return "設定の保存に失敗しました: " + err.Error()
	}
	return "ロールID " + roleID + " からサーバーID " + serverID + " を削除しました。"
}

func (m *ConfigManager) GetRole(roleID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids, ok := m.config.ServerRoles[roleID]
	if !ok {
		return []string{}
	}
	return append([]string(nil), ids...)
}

func (m *ConfigManager) GetServerID(roleID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids, ok := m.config.ServerRoles[roleID]
	if !ok {
		return []string{}
	}
	return append([]string(nil), ids...)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
