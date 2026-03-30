package service

import (
	"device_only/internal/models"
	"device_only/internal/repository"
	"errors"
	"time"

	"github.com/google/uuid"
)

type DeviceService struct {
	repo repository.Repository
}

func NewDeviceService(repo repository.Repository) *DeviceService {
	return &DeviceService{repo: repo}
}

func (s *DeviceService) RegisterDevice(req *models.RegisterDeviceRequest) (*models.RegisterDeviceResponse, error) {
	if req.UserRef == "" || req.PublicKey == "" {
		return nil, errors.New("missing user_ref or public_key")
	}

	bindingID := "bind_" + uuid.New().String()

	device := &models.Device{
		DeviceBindingID: bindingID,
		ClientID:        req.ClientID,
		UserRef:         req.UserRef,
		PublicKey:       req.PublicKey,
		DeviceID:        req.DeviceInfo.DeviceID,
		Platform:        req.DeviceInfo.Platform,
		DeviceModel:     req.DeviceInfo.DeviceModel,
		OSVersion:       req.DeviceInfo.OSVersion,
		CreatedAt:       time.Now(),
		Status:          "ACTIVE",
	}

	if err := s.repo.CreateDevice(device); err != nil {
		return nil, err
	}

	return &models.RegisterDeviceResponse{
		RequestID:       "req_" + uuid.New().String(),
		Timestamp:       time.Now().Format(time.RFC3339),
		ClientID:        req.ClientID,
		DeviceBindingID: bindingID,
		Status:          "REGISTERED",
	}, nil
}

func (s *DeviceService) RevokeDevice(req *models.RevokeDeviceRequest) (*models.RevokeDeviceResponse, error) {
	if req.DeviceBindingID == "" {
		return nil, errors.New("missing device_binding_id")
	}

	device, err := s.repo.GetDevice(req.DeviceBindingID)
	if err != nil {
		return nil, errors.New("device binding not found")
	}

	if device.UserRef != req.UserRef {
		return nil, errors.New("unauthorized operation")
	}

	if err := s.repo.RevokeDevice(req.DeviceBindingID); err != nil {
		return nil, err
	}

	return &models.RevokeDeviceResponse{
		RequestID: "req_" + uuid.New().String(),
		Timestamp: time.Now().Format(time.RFC3339),
		Status:    "REVOKED",
	}, nil
}

func (s *DeviceService) CheckDeviceBinding(clientID string, userRef string, deviceID string) (*models.CheckDeviceResponse, error) {
	if clientID == "" || userRef == "" {
		return nil, errors.New("missing client_id or user_ref")
	}

	devices, err := s.repo.GetDevicesByClientAndUser(clientID, userRef)
	if err != nil {
		return nil, err
	}

	resp := &models.CheckDeviceResponse{
		RequestID: "req_" + uuid.New().String(),
		Timestamp: time.Now().Format(time.RFC3339),
		Status:    "SUCCESS",
	}

	if len(devices) == 0 {
		resp.HasActiveDevice = false
		resp.Message = "User has no active devices registered for this client"
	} else {
		if deviceID != "" {
			found := false
			for _, d := range devices {
				if d.DeviceID == deviceID {
					resp.DeviceBindingID = d.DeviceBindingID
					found = true
					break
				}
			}
			if found {
				resp.HasActiveDevice = true
				resp.Message = "Device is already registered"
			} else {
				resp.HasActiveDevice = false
				resp.Message = "User exists but this device is not registered"
			}
		} else {
			resp.HasActiveDevice = true
			resp.DeviceBindingID = devices[0].DeviceBindingID
			resp.Message = "User has active devices"
		}
	}

	return resp, nil
}

func (s *DeviceService) UpdateDevice(req *models.UpdateDeviceRequest) (*models.UpdateDeviceResponse, error) {
	if req.ClientID == "" || req.UserRef == "" || req.PublicKey == "" {
		return nil, errors.New("missing client_id, user_ref, or public_key")
	}

	d, err := s.repo.ReplaceDeviceBinding(req.ClientID, req.UserRef, req.PublicKey, req.DeviceInfo.DeviceID, req.DeviceInfo.Platform, req.DeviceInfo.DeviceModel, req.DeviceInfo.OSVersion)
	if err != nil {
		return nil, errors.New("device not found or update failed")
	}

	return &models.UpdateDeviceResponse{
		RequestID:       "req_" + uuid.New().String(),
		Timestamp:       time.Now().Format(time.RFC3339),
		ClientID:        req.ClientID,
		DeviceBindingID: d.DeviceBindingID,
		Status:          "REGISTERED",
	}, nil
}
