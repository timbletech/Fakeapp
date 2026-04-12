package repository

import (
	"database/sql"
	"device_only/internal/models"
	"time"

	_ "github.com/lib/pq"
)

type Repository interface {
	CreateDevice(d *models.Device) error
	GetDevice(deviceBindingID string) (*models.Device, error)
	GetDevicesByUserRef(userRef string) ([]*models.Device, error)
	GetDevicesByClientAndUser(clientID string, userRef string) ([]*models.Device, error)
	ReplaceDeviceBinding(clientID string, userRef string, publicKey string, deviceID string, platform string, deviceModel string, osVersion string) (*models.Device, error)
	RevokeDevice(deviceBindingID string) error

	CreateAuthSession(s *models.AuthSession) error
	GetAuthSession(authSessionID string) (*models.AuthSession, error)
	CompleteAuthSession(authSessionID string, status string) error

	CreateAuthContextToken(t *models.AuthContextToken) error
	GetAuthContextToken(token string) (*models.AuthContextToken, error)

	LogAudit(l *models.AuditLog) error

	CreateDeviceApprovalRequest(r *models.DeviceApprovalRequest) error
	GetDeviceApprovalRequest(id string) (*models.DeviceApprovalRequest, error)
	UpdateDeviceApprovalStatus(id string, status string, resolvedBy string) error
	GetPendingApprovals(clientID string, userRef string) ([]*models.DeviceApprovalRequest, error)
}

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func fromNullString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func (r *PostgresRepo) CreateDevice(d *models.Device) error {
	query := `INSERT INTO devices (device_binding_id, client_id, user_ref, public_key, device_id, platform, device_model, os_version, created_at, status) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			  ON CONFLICT (client_id, user_ref) DO UPDATE SET
			    device_binding_id = EXCLUDED.device_binding_id,
			    public_key = EXCLUDED.public_key,
			    device_id = EXCLUDED.device_id,
			    platform = EXCLUDED.platform,
			    device_model = EXCLUDED.device_model,
			    os_version = EXCLUDED.os_version,
			    created_at = EXCLUDED.created_at,
			    status = EXCLUDED.status`
	_, err := r.db.Exec(query,
		d.DeviceBindingID,
		d.ClientID,
		d.UserRef,
		d.PublicKey,
		nullString(d.DeviceID),
		nullString(d.Platform),
		nullString(d.DeviceModel),
		nullString(d.OSVersion),
		d.CreatedAt,
		d.Status)
	return err
}

func (r *PostgresRepo) GetDevice(deviceBindingID string) (*models.Device, error) {
	d := &models.Device{}
	var devID, plat, model, os sql.NullString
	query := `SELECT device_binding_id, client_id, user_ref, public_key, device_id, platform, device_model, os_version, created_at, status 
			  FROM devices WHERE device_binding_id = $1`
	err := r.db.QueryRow(query, deviceBindingID).Scan(
		&d.DeviceBindingID, &d.ClientID, &d.UserRef, &d.PublicKey, &devID, &plat, &model, &os, &d.CreatedAt, &d.Status)
	if err != nil {
		return nil, err
	}
	d.DeviceID = fromNullString(devID)
	d.Platform = fromNullString(plat)
	d.DeviceModel = fromNullString(model)
	d.OSVersion = fromNullString(os)
	return d, nil
}

func (r *PostgresRepo) GetDevicesByUserRef(userRef string) ([]*models.Device, error) {
	query := `SELECT device_binding_id, client_id, user_ref, public_key, device_id, platform, device_model, os_version, created_at, status 
			  FROM devices WHERE user_ref = $1 AND status = 'ACTIVE'`
	rows, err := r.db.Query(query, userRef)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		d := &models.Device{}
		var devID, plat, model, os sql.NullString
		err := rows.Scan(&d.DeviceBindingID, &d.ClientID, &d.UserRef, &d.PublicKey, &devID, &plat, &model, &os, &d.CreatedAt, &d.Status)
		if err != nil {
			return nil, err
		}
		d.DeviceID = fromNullString(devID)
		d.Platform = fromNullString(plat)
		d.DeviceModel = fromNullString(model)
		d.OSVersion = fromNullString(os)
		devices = append(devices, d)
	}
	return devices, nil
}

func (r *PostgresRepo) GetDevicesByClientAndUser(clientID string, userRef string) ([]*models.Device, error) {
	query := `SELECT device_binding_id, client_id, user_ref, public_key, device_id, platform, device_model, os_version, created_at, status 
			  FROM devices WHERE client_id = $1 AND user_ref = $2 AND status = 'ACTIVE'`
	rows, err := r.db.Query(query, clientID, userRef)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		d := &models.Device{}
		var devID, plat, model, os sql.NullString
		err := rows.Scan(&d.DeviceBindingID, &d.ClientID, &d.UserRef, &d.PublicKey, &devID, &plat, &model, &os, &d.CreatedAt, &d.Status)
		if err != nil {
			return nil, err
		}
		d.DeviceID = fromNullString(devID)
		d.Platform = fromNullString(plat)
		d.DeviceModel = fromNullString(model)
		d.OSVersion = fromNullString(os)
		devices = append(devices, d)
	}
	return devices, nil
}

func (r *PostgresRepo) ReplaceDeviceBinding(clientID string, userRef string, publicKey string, deviceID string, platform string, deviceModel string, osVersion string) (*models.Device, error) {
	query := `UPDATE devices SET public_key = $1, device_id = $2, platform = $3, device_model = $4, os_version = $5 
	          WHERE client_id = $6 AND user_ref = $7 AND status = 'ACTIVE' RETURNING device_binding_id, created_at, status`

	d := &models.Device{
		ClientID:    clientID,
		UserRef:     userRef,
		PublicKey:   publicKey,
		DeviceID:    deviceID,
		Platform:    platform,
		DeviceModel: deviceModel,
		OSVersion:   osVersion,
	}

	err := r.db.QueryRow(query, publicKey, nullString(deviceID), nullString(platform), nullString(deviceModel), nullString(osVersion), clientID, userRef).Scan(&d.DeviceBindingID, &d.CreatedAt, &d.Status)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (r *PostgresRepo) RevokeDevice(deviceBindingID string) error {
	query := `UPDATE devices SET status = 'REVOKED' WHERE device_binding_id = $1`
	_, err := r.db.Exec(query, deviceBindingID)
	return err
}

func (r *PostgresRepo) CreateAuthSession(s *models.AuthSession) error {
	query := `INSERT INTO auth_sessions (auth_session_id, user_ref, challenge, challenge_id, device_binding_id, expires_at, status, created_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(query, s.AuthSessionID, s.UserRef, s.Challenge, nullString(s.ChallengeID), nullString(s.DeviceBindingID), s.ExpiresAt, s.Status, s.CreatedAt)
	return err
}

func (r *PostgresRepo) GetAuthSession(authSessionID string) (*models.AuthSession, error) {
	s := &models.AuthSession{}
	var chalID, devBindID sql.NullString
	query := `SELECT auth_session_id, user_ref, challenge, challenge_id, device_binding_id, expires_at, status, created_at 
			  FROM auth_sessions WHERE auth_session_id = $1`
	err := r.db.QueryRow(query, authSessionID).Scan(
		&s.AuthSessionID, &s.UserRef, &s.Challenge, &chalID, &devBindID, &s.ExpiresAt, &s.Status, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	s.ChallengeID = fromNullString(chalID)
	s.DeviceBindingID = fromNullString(devBindID)
	return s, nil
}

func (r *PostgresRepo) CompleteAuthSession(authSessionID string, status string) error {
	query := `UPDATE auth_sessions SET status = $1 WHERE auth_session_id = $2`
	_, err := r.db.Exec(query, status, authSessionID)
	return err
}

func (r *PostgresRepo) CreateAuthContextToken(t *models.AuthContextToken) error {
	query := `INSERT INTO auth_context_tokens (token, user_ref, expires_at, status, created_at) 
			  VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, t.Token, t.UserRef, t.ExpiresAt, t.Status, t.CreatedAt)
	return err
}

func (r *PostgresRepo) GetAuthContextToken(token string) (*models.AuthContextToken, error) {
	t := &models.AuthContextToken{}
	query := `SELECT token, user_ref, expires_at, status, created_at 
			  FROM auth_context_tokens WHERE token = $1`
	err := r.db.QueryRow(query, token).Scan(&t.Token, &t.UserRef, &t.ExpiresAt, &t.Status, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *PostgresRepo) LogAudit(l *models.AuditLog) error {
	query := `INSERT INTO audit_logs (user_ref, action, decision, ip_address, device_id, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, nullString(l.UserRef), l.Action, l.Decision, nullString(l.IPAddress), nullString(l.DeviceID), time.Now())
	return err
}

func (r *PostgresRepo) CreateDeviceApprovalRequest(req *models.DeviceApprovalRequest) error {
	query := `INSERT INTO device_approval_requests
		(id, client_id, user_ref, requesting_device_id, requesting_device_info, requesting_public_key, main_device_binding_id, status, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(query,
		req.ID, req.ClientID, req.UserRef, req.RequestingDeviceID,
		nullString(req.RequestingDeviceInfo), nullString(req.RequestingPublicKey),
		req.MainDeviceBindingID, req.Status, req.CreatedAt, req.ExpiresAt)
	return err
}

func (r *PostgresRepo) GetDeviceApprovalRequest(id string) (*models.DeviceApprovalRequest, error) {
	a := &models.DeviceApprovalRequest{}
	var deviceInfo, publicKey, resolvedBy sql.NullString
	var resolvedAt sql.NullTime
	query := `SELECT id, client_id, user_ref, requesting_device_id, requesting_device_info,
		requesting_public_key, main_device_binding_id, status, created_at, expires_at, resolved_at, resolved_by
		FROM device_approval_requests WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&a.ID, &a.ClientID, &a.UserRef, &a.RequestingDeviceID, &deviceInfo,
		&publicKey, &a.MainDeviceBindingID, &a.Status, &a.CreatedAt, &a.ExpiresAt,
		&resolvedAt, &resolvedBy)
	if err != nil {
		return nil, err
	}
	a.RequestingDeviceInfo = fromNullString(deviceInfo)
	a.RequestingPublicKey = fromNullString(publicKey)
	a.ResolvedBy = fromNullString(resolvedBy)
	if resolvedAt.Valid {
		a.ResolvedAt = &resolvedAt.Time
	}
	return a, nil
}

func (r *PostgresRepo) UpdateDeviceApprovalStatus(id string, status string, resolvedBy string) error {
	query := `UPDATE device_approval_requests SET status = $1, resolved_at = NOW(), resolved_by = $2 WHERE id = $3`
	_, err := r.db.Exec(query, status, nullString(resolvedBy), id)
	return err
}

func (r *PostgresRepo) GetPendingApprovals(clientID string, userRef string) ([]*models.DeviceApprovalRequest, error) {
	query := `SELECT id, client_id, user_ref, requesting_device_id, requesting_device_info,
		requesting_public_key, main_device_binding_id, status, created_at, expires_at, resolved_at, resolved_by
		FROM device_approval_requests
		WHERE client_id = $1 AND user_ref = $2 AND status = 'PENDING' AND expires_at > NOW()
		ORDER BY created_at DESC`
	rows, err := r.db.Query(query, clientID, userRef)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.DeviceApprovalRequest
	for rows.Next() {
		a := &models.DeviceApprovalRequest{}
		var deviceInfo, publicKey, resolvedBy sql.NullString
		var resolvedAt sql.NullTime
		err := rows.Scan(
			&a.ID, &a.ClientID, &a.UserRef, &a.RequestingDeviceID, &deviceInfo,
			&publicKey, &a.MainDeviceBindingID, &a.Status, &a.CreatedAt, &a.ExpiresAt,
			&resolvedAt, &resolvedBy)
		if err != nil {
			return nil, err
		}
		a.RequestingDeviceInfo = fromNullString(deviceInfo)
		a.RequestingPublicKey = fromNullString(publicKey)
		a.ResolvedBy = fromNullString(resolvedBy)
		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}
		results = append(results, a)
	}
	return results, nil
}
