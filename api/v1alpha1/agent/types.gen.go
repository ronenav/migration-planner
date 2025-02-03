// Package v1alpha1 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.3.0 DO NOT EDIT.
package v1alpha1

import (
	externalRef0 "github.com/kubev2v/migration-planner/api/v1alpha1"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// AgentStatusUpdate defines model for AgentStatusUpdate.
type AgentStatusUpdate struct {
	CredentialUrl string             `json:"credentialUrl"`
	SourceId      openapi_types.UUID `json:"sourceId"`
	Status        string             `json:"status"`
	StatusInfo    string             `json:"statusInfo"`
	Version       string             `json:"version"`
}

// SourceStatusUpdate defines model for SourceStatusUpdate.
type SourceStatusUpdate struct {
	AgentId   openapi_types.UUID     `json:"agentId"`
	Inventory externalRef0.Inventory `json:"inventory"`
}

// UpdateAgentStatusJSONRequestBody defines body for UpdateAgentStatus for application/json ContentType.
type UpdateAgentStatusJSONRequestBody = AgentStatusUpdate

// UpdateSourceInventoryJSONRequestBody defines body for UpdateSourceInventory for application/json ContentType.
type UpdateSourceInventoryJSONRequestBody = SourceStatusUpdate
