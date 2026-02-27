package api

import (
	"context"
	"encoding/base64"
	"fmt"
)

// GetProfile returns the current user's profile.
func (c *Client) GetProfile(ctx context.Context) (*UserProfile, error) {
	resp, err := c.get(ctx, "/api/v1/user/profile")
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	result, err := decodeResponse[UserProfile](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// UpdateProfile updates the current user's profile.
func (c *Client) UpdateProfile(ctx context.Context, req UpdateProfileRequest) error {
	resp, err := c.patch(ctx, "/api/v1/user/profile", req)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	_, err = decodeResponse[any](resp)
	return err
}

// ListUserSessions lists the current user's sessions.
func (c *Client) ListUserSessions(ctx context.Context, page, size int) (*PageDTO[LoginRecord], error) {
	path := fmt.Sprintf("/api/v1/user/session?page=%d&limit=%d", page, size)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	result, err := decodeResponse[PageDTO[LoginRecord]](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// LogoutSession logs out a specific session by UUID.
func (c *Client) LogoutSession(ctx context.Context, uuid string) error {
	resp, err := c.delete(ctx, fmt.Sprintf("/api/v1/user/session/%s/logout", uuid))
	if err != nil {
		return fmt.Errorf("logout session: %w", err)
	}
	_, err = decodeResponse[any](resp)
	return err
}

// LogoutAllSessions logs out all sessions for the current user.
func (c *Client) LogoutAllSessions(ctx context.Context) error {
	resp, err := c.delete(ctx, "/api/v1/user/logout")
	if err != nil {
		return fmt.Errorf("logout all sessions: %w", err)
	}
	_, err = decodeResponse[any](resp)
	if err != nil {
		return err
	}
	c.session = ""
	return nil
}

// ListUserCAs lists CAs bound to the current user.
func (c *Client) ListUserCAs(ctx context.Context, page, size int) (*PageDTO[CACert], error) {
	path := fmt.Sprintf("/api/v1/user/cert/ca?page=%d&limit=%d", page, size)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list user CAs: %w", err)
	}
	result, err := decodeResponse[PageDTO[CACert]](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetUserCACert gets the PEM certificate for a CA.
func (c *Client) GetUserCACert(ctx context.Context, uuid string, chain, needRoot bool) (*CertContent, error) {
	path := fmt.Sprintf("/api/v1/user/cert/ca/%s/cer?chain=%v&needRootCa=%v", uuid, chain, needRoot)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("get CA cert: %w", err)
	}
	result, err := decodeResponse[CertContent](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// ListUserSSLCerts lists SSL certs belonging to the current user.
func (c *Client) ListUserSSLCerts(ctx context.Context, page, size int) (*PageDTO[SSLCert], error) {
	path := fmt.Sprintf("/api/v1/user/cert/ssl?page=%d&limit=%d", page, size)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list SSL certs: %w", err)
	}
	result, err := decodeResponse[PageDTO[SSLCert]](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetUserSSLCert gets the PEM certificate content.
func (c *Client) GetUserSSLCert(ctx context.Context, uuid string) (*CertContent, error) {
	resp, err := c.get(ctx, fmt.Sprintf("/api/v1/user/cert/ssl/%s/cer", uuid))
	if err != nil {
		return nil, fmt.Errorf("get SSL cert: %w", err)
	}
	result, err := decodeResponse[CertContent](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetUserSSLPrivKey retrieves the encrypted private key.
func (c *Client) GetUserSSLPrivKey(ctx context.Context, uuid, password string) (*PrivKeyResponse, error) {
	resp, err := c.post(ctx, fmt.Sprintf("/api/v1/user/cert/ssl/%s/privkey", uuid), GetPrivKeyRequest{Password: password})
	if err != nil {
		return nil, fmt.Errorf("get SSL privkey: %w", err)
	}
	result, err := decodeResponse[PrivKeyResponse](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// RequestSSLCert requests a new SSL certificate.
func (c *Client) RequestSSLCert(ctx context.Context, req RequestSSLCertRequest) (*SSLCert, error) {
	resp, err := c.post(ctx, "/api/v1/user/cert/ssl", req)
	if err != nil {
		return nil, fmt.Errorf("request SSL cert: %w", err)
	}
	result, err := decodeResponse[SSLCert](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// RenewSSLCert renews an SSL certificate.
func (c *Client) RenewSSLCert(ctx context.Context, uuid string, req RenewSSLCertRequest) (*SSLCert, error) {
	resp, err := c.put(ctx, fmt.Sprintf("/api/v1/user/cert/ssl/%s", uuid), req)
	if err != nil {
		return nil, fmt.Errorf("renew SSL cert: %w", err)
	}
	result, err := decodeResponse[SSLCert](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// DeleteSSLCert deletes an SSL certificate.
func (c *Client) DeleteSSLCert(ctx context.Context, uuid string) error {
	resp, err := c.delete(ctx, fmt.Sprintf("/api/v1/user/cert/ssl/%s", uuid))
	if err != nil {
		return fmt.Errorf("delete SSL cert: %w", err)
	}
	_, err = decodeResponse[any](resp)
	return err
}

// UpdateSSLCertComment updates the comment on an SSL cert.
func (c *Client) UpdateSSLCertComment(ctx context.Context, uuid, comment string) error {
	resp, err := c.patch(ctx, fmt.Sprintf("/api/v1/user/cert/ssl/%s/comment", uuid), UpdateCommentRequest{Comment: comment})
	if err != nil {
		return fmt.Errorf("update SSL cert comment: %w", err)
	}
	_, err = decodeResponse[any](resp)
	return err
}

// AnalyzeCert analyzes a PEM certificate.
// The API expects the PEM content to be base64-encoded before sending.
func (c *Client) AnalyzeCert(ctx context.Context, pem string) (*CertAnalysis, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(pem))
	resp, err := c.post(ctx, "/api/v1/user/cert/analyze", AnalyzeCertRequest{Cert: encoded})
	if err != nil {
		return nil, fmt.Errorf("analyze cert: %w", err)
	}
	result, err := decodeResponse[CertAnalysis](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// AnalyzePrivKey analyzes a private key.
// The API expects the PEM content to be base64-encoded before sending.
func (c *Client) AnalyzePrivKey(ctx context.Context, privKey, password string) (*PrivKeyAnalysis, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(privKey))
	resp, err := c.post(ctx, "/api/v1/user/cert/privkey/analyze", AnalyzePrivKeyRequest{PrivKey: encoded, Password: password})
	if err != nil {
		return nil, fmt.Errorf("analyze privkey: %w", err)
	}
	result, err := decodeResponse[PrivKeyAnalysis](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// ConvertPEMtoPFX converts a PEM cert+key to PFX format.
func (c *Client) ConvertPEMtoPFX(ctx context.Context, req ConvertPEMtoPFXRequest) (*ConvertResult, error) {
	resp, err := c.post(ctx, "/api/v1/user/cert/convert/pem/to/pfx", req)
	if err != nil {
		return nil, fmt.Errorf("convert PEM to PFX: %w", err)
	}
	result, err := decodeResponse[ConvertResult](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// ConvertPEMtoDER converts a PEM certificate to DER format.
func (c *Client) ConvertPEMtoDER(ctx context.Context, pem string) (*ConvertResult, error) {
	resp, err := c.post(ctx, "/api/v1/user/cert/convert/pem/to/der", ConvertRequest{Cert: pem})
	if err != nil {
		return nil, fmt.Errorf("convert PEM to DER: %w", err)
	}
	result, err := decodeResponse[ConvertResult](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// ConvertDERtoPEM converts a base64-encoded DER certificate to PEM format.
func (c *Client) ConvertDERtoPEM(ctx context.Context, der string) (*ConvertResult, error) {
	resp, err := c.post(ctx, "/api/v1/user/cert/convert/der/to/pem", ConvertRequest{Cert: der})
	if err != nil {
		return nil, fmt.Errorf("convert DER to PEM: %w", err)
	}
	result, err := decodeResponse[ConvertResult](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}
