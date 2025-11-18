/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"encoding/json"
	"strings"
)

// MessageType represents different types of messages in the system
type MessageType int

const (
	MessageTypeUnknown MessageType = iota
	MessageTypeAPIResponse
	MessageTypeJobCallback
	MessageTypeBPMNError
)

// String returns string representation of MessageType
func (mt MessageType) String() string {
	switch mt {
	case MessageTypeAPIResponse:
		return "APIResponse"
	case MessageTypeJobCallback:
		return "JobCallback"
	case MessageTypeBPMNError:
		return "BPMNError"
	default:
		return "Unknown"
	}
}

// MessageClassifier classifies message types based on content
type MessageClassifier struct{}

// NewMessageClassifier creates a new message classifier
func NewMessageClassifier() *MessageClassifier {
	return &MessageClassifier{}
}

// ClassifyMessage determines the type of a message based on its JSON structure
func (mc *MessageClassifier) ClassifyMessage(messageJSON string) MessageType {
	if len(messageJSON) == 0 {
		return MessageTypeUnknown
	}

	// Parse as generic JSON to inspect structure
	var message map[string]interface{}
	if err := json.Unmarshal([]byte(messageJSON), &message); err != nil {
		return MessageTypeUnknown
	}

	// Check for API Response pattern
	if mc.isAPIResponse(message) {
		return MessageTypeAPIResponse
	}

	// Check for BPMN Error pattern
	if mc.isBPMNError(message) {
		return MessageTypeBPMNError
	}

	// Check for Job Callback pattern
	if mc.isJobCallback(message) {
		return MessageTypeJobCallback
	}

	return MessageTypeUnknown
}

// isAPIResponse checks if message is an API response
// API responses have: type field ending with "_response", request_id, success fields
func (mc *MessageClassifier) isAPIResponse(message map[string]interface{}) bool {
	// Check for required fields
	msgType, hasType := message["type"].(string)
	_, hasRequestID := message["request_id"]
	_, hasSuccess := message["success"]

	// API responses must have type ending with "_response"
	return hasType && hasRequestID && hasSuccess && strings.HasSuffix(msgType, "_response")
}

// isBPMNError checks if message is a BPMN error callback
// BPMN errors have: type="bpmn_error", job_id, element_id, error_code
func (mc *MessageClassifier) isBPMNError(message map[string]interface{}) bool {
	msgType, hasType := message["type"].(string)
	_, hasJobID := message["job_id"]
	_, hasElementID := message["element_id"]
	_, hasErrorCode := message["error_code"]

	return hasType && msgType == "bpmn_error" && hasJobID && hasElementID && hasErrorCode
}

// isJobCallback checks if message is a job completion callback
// Job callbacks have: job_id, element_id, token_id, process_instance_id, status
func (mc *MessageClassifier) isJobCallback(message map[string]interface{}) bool {
	_, hasJobID := message["job_id"]
	_, hasElementID := message["element_id"]
	_, hasTokenID := message["token_id"]
	_, hasProcessInstanceID := message["process_instance_id"]
	_, hasStatus := message["status"]

	// Must NOT have "type" field (to distinguish from BPMN errors)
	_, hasType := message["type"]

	return !hasType && hasJobID && hasElementID && hasTokenID && hasProcessInstanceID && hasStatus
}
