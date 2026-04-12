package service

import (
	"device_only/internal/config"
	"device_only/internal/models"
	"device_only/internal/repository"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type DeviceVerifyService struct {
	repo repository.Repository
	cfg  *config.Config
}

func NewDeviceVerifyService(repo repository.Repository, cfg *config.Config) *DeviceVerifyService {
	return &DeviceVerifyService{repo: repo, cfg: cfg}
}

// InitiateVerification checks if the requesting device is already known.
// If known → returns KNOWN_DEVICE immediately.
// If unknown → creates a pending approval request for the main device to act on.
func (s *DeviceVerifyService) InitiateVerification(req *models.DeviceVerifyRequest) (*models.DeviceVerifyResponse, error) {
	if req.ClientID == "" || req.UserRef == "" || req.DeviceInfo.DeviceID == "" {
		return nil, errors.New("client_id, user_ref, and device_info.device_id are required")
	}

	// Check if this device is already registered and active
	devices, err := s.repo.GetDevicesByClientAndUser(req.ClientID, req.UserRef)
	if err != nil {
		return nil, err
	}

	for _, d := range devices {
		if d.DeviceID == req.DeviceInfo.DeviceID && d.Status == "ACTIVE" {
			return &models.DeviceVerifyResponse{
				RequestID: "req_" + uuid.New().String(),
				Timestamp: time.Now().Format(time.RFC3339),
				Status:    "KNOWN_DEVICE",
				Message:   "Device is already registered and trusted",
			}, nil
		}
	}

	// Device is unknown — find the main (first active) device to send the alert to
	if len(devices) == 0 {
		return nil, errors.New("no registered device found for this user; register a device first")
	}
	mainDevice := devices[0]

	// Serialize requesting device info as JSON
	deviceInfoJSON, _ := json.Marshal(req.DeviceInfo)

	approvalID := "appr_" + uuid.New().String()
	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(s.cfg.DeviceApprovalExpirySeconds) * time.Second)

	approval := &models.DeviceApprovalRequest{
		ID:                   approvalID,
		ClientID:             req.ClientID,
		UserRef:              req.UserRef,
		RequestingDeviceID:   req.DeviceInfo.DeviceID,
		RequestingDeviceInfo: string(deviceInfoJSON),
		RequestingPublicKey:  req.PublicKey,
		MainDeviceBindingID:  mainDevice.DeviceBindingID,
		Status:               "PENDING",
		CreatedAt:            now,
		ExpiresAt:            expiresAt,
	}

	if err := s.repo.CreateDeviceApprovalRequest(approval); err != nil {
		return nil, err
	}

	_ = s.repo.LogAudit(&models.AuditLog{
		UserRef:  req.UserRef,
		Action:   "DEVICE_VERIFY_INITIATED",
		Decision: "PENDING",
		DeviceID: req.DeviceInfo.DeviceID,
	})

	return &models.DeviceVerifyResponse{
		RequestID:  "req_" + uuid.New().String(),
		Timestamp:  time.Now().Format(time.RFC3339),
		ApprovalID: approvalID,
		Status:     "PENDING",
		Message:    "Approval request sent to your main device",
	}, nil
}

// RespondToApproval lets the main device approve or deny the request.
func (s *DeviceVerifyService) RespondToApproval(req *models.DeviceApprovalActionRequest) (*models.DeviceApprovalActionResponse, error) {
	if req.ApprovalID == "" || req.Action == "" {
		return nil, errors.New("approval_id and action are required")
	}

	action := req.Action
	if action != "approve" && action != "deny" {
		return nil, errors.New("action must be 'approve' or 'deny'")
	}

	approval, err := s.repo.GetDeviceApprovalRequest(req.ApprovalID)
	if err != nil {
		return nil, errors.New("approval request not found")
	}

	if approval.ClientID != req.ClientID || approval.UserRef != req.UserRef {
		return nil, errors.New("unauthorized: client_id or user_ref mismatch")
	}

	if approval.Status != "PENDING" {
		return nil, errors.New("approval request is no longer pending (status: " + approval.Status + ")")
	}

	if time.Now().After(approval.ExpiresAt) {
		_ = s.repo.UpdateDeviceApprovalStatus(approval.ID, "EXPIRED", "")
		return nil, errors.New("approval request has expired")
	}

	var newStatus string
	if action == "approve" {
		newStatus = "APPROVED"
	} else {
		newStatus = "DENIED"
	}

	if err := s.repo.UpdateDeviceApprovalStatus(approval.ID, newStatus, approval.MainDeviceBindingID); err != nil {
		return nil, err
	}

	// If approved, auto-register the new device as trusted using the public key
	// that was submitted with the original verification request
	if newStatus == "APPROVED" && approval.RequestingPublicKey != "" {
		var devInfo models.DeviceInfo
		_ = json.Unmarshal([]byte(approval.RequestingDeviceInfo), &devInfo)

		newBinding := &models.Device{
			DeviceBindingID: "bind_" + uuid.New().String(),
			ClientID:        approval.ClientID,
			UserRef:         approval.UserRef,
			PublicKey:       approval.RequestingPublicKey,
			DeviceID:        approval.RequestingDeviceID,
			Platform:        devInfo.Platform,
			DeviceModel:     devInfo.DeviceModel,
			OSVersion:       devInfo.OSVersion,
			CreatedAt:       time.Now(),
			Status:          "ACTIVE",
		}
		_ = s.repo.CreateDevice(newBinding)
	}

	_ = s.repo.LogAudit(&models.AuditLog{
		UserRef:  req.UserRef,
		Action:   "DEVICE_VERIFY_" + newStatus,
		Decision: newStatus,
		DeviceID: approval.RequestingDeviceID,
	})

	return &models.DeviceApprovalActionResponse{
		RequestID: "req_" + uuid.New().String(),
		Timestamp: time.Now().Format(time.RFC3339),
		Status:    newStatus,
		Message:   "Device " + action + "d successfully",
	}, nil
}

// CheckApprovalStatus is polled by the new device to see if it was approved.
func (s *DeviceVerifyService) CheckApprovalStatus(req *models.DeviceApprovalStatusRequest) (*models.DeviceApprovalStatusResponse, error) {
	if req.ApprovalID == "" {
		return nil, errors.New("approval_id is required")
	}

	approval, err := s.repo.GetDeviceApprovalRequest(req.ApprovalID)
	if err != nil {
		return nil, errors.New("approval request not found")
	}

	if approval.ClientID != req.ClientID {
		return nil, errors.New("unauthorized: client_id mismatch")
	}

	resp := &models.DeviceApprovalStatusResponse{
		RequestID:  "req_" + uuid.New().String(),
		Timestamp:  time.Now().Format(time.RFC3339),
		ApprovalID: approval.ID,
		Status:     approval.Status,
	}

	// Check expiry for still-pending requests
	if approval.Status == "PENDING" && time.Now().After(approval.ExpiresAt) {
		_ = s.repo.UpdateDeviceApprovalStatus(approval.ID, "EXPIRED", "")
		resp.Status = "EXPIRED"
		resp.Message = "Approval request has expired"
		return resp, nil
	}

	switch approval.Status {
	case "PENDING":
		remaining := int(approval.ExpiresAt.Sub(time.Now()).Seconds())
		resp.ExpiresInSeconds = remaining
		resp.Message = "Waiting for approval from main device"
	case "APPROVED":
		resp.Message = "Device approved — proceed with login"
	case "DENIED":
		resp.Message = "Device denied by user"
	case "EXPIRED":
		resp.Message = "Approval request has expired"
	}

	return resp, nil
}

// GetPendingApprovals returns all pending approval requests for a user (shown on the main device).
func (s *DeviceVerifyService) GetPendingApprovals(clientID, userRef string) (*models.PendingApprovalsResponse, error) {
	if clientID == "" || userRef == "" {
		return nil, errors.New("client_id and user_ref are required")
	}

	approvals, err := s.repo.GetPendingApprovals(clientID, userRef)
	if err != nil {
		return nil, err
	}

	var items []models.PendingApprovalItem
	for _, a := range approvals {
		var devInfo models.DeviceInfo
		_ = json.Unmarshal([]byte(a.RequestingDeviceInfo), &devInfo)

		items = append(items, models.PendingApprovalItem{
			ApprovalID:       a.ID,
			RequestingDevice: devInfo,
			CreatedAt:        a.CreatedAt.Format(time.RFC3339),
			ExpiresAt:        a.ExpiresAt.Format(time.RFC3339),
		})
	}

	if items == nil {
		items = []models.PendingApprovalItem{}
	}

	return &models.PendingApprovalsResponse{
		RequestID: "req_" + uuid.New().String(),
		Timestamp: time.Now().Format(time.RFC3339),
		Pending:   items,
	}, nil
}
