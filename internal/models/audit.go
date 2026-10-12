package models

import "time"

type AuditEntry struct {
	ID           string    `json:"id"`
	Actor        string    `json:"actor"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resourceType"`
	ResourceID   string    `json:"resourceId"`
	Detail       string    `json:"detail,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type BackupManifest struct {
	ID            string    `json:"id"`
	CreatedAt     time.Time `json:"createdAt"`
	Path          string    `json:"path"`
	SizeBytes     int64     `json:"sizeBytes"`
	DataPath      string    `json:"dataPath,omitempty"`
	DataSizeBytes int64     `json:"dataSizeBytes,omitempty"`
	Tables        []string  `json:"tables"`
}
