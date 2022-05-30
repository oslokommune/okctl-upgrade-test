// Package loki knows how to add persistence to a Loki deployment
package loki

import "time"

const oneDay = 24 * time.Hour

type schemaConfig struct {
	From        string            `json:"from"`
	Store       string            `json:"store"`
	ObjectStore string            `json:"object_store"`
	Schema      string            `json:"schema"`
	Index       schemaConfigIndex `json:"index"`
}

type schemaConfigIndex struct {
	Prefix string `json:"prefix"`
	Period string `json:"period"`
}

type storageConfig struct {
	S3                   string            `json:"s3"`
	BucketNames          string            `json:"bucketnames"`
	DynamoDB             map[string]string `json:"dynamodb"`
	ServerSideEncryption bool              `json:"sse_encryption"`
}

type tableManagerIndexTablesProvisioning struct {
	EnableOnDemandThroughputMode         bool `json:"enable_ondemand_throughput_mode"`
	EnableInactiveThroughputOnDemandMode bool `json:"enable_inactive_throughput_on_demand_mode"`
}

type tableManager struct {
	RetentionDeletesEnabled bool                                `json:"retention_deletes_enabled"`
	RetentionPeriod         string                              `json:"retention_period"`
	IndexTablesProvisioning tableManagerIndexTablesProvisioning `json:"index_tables_provisioning"`
}
