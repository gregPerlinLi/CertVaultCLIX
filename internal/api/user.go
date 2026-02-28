package api

import (
	"context"
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
func (c *Client) GetUserCACert(ctx context.Context, uuid string, chain, needRoot bool) (string, error) {
	path := fmt.Sprintf("/api/v1/user/cert/ca/%s/cer?isChain=%v&needRootCa=%v", uuid, chain, needRoot)
	resp, err := c.get(ctx, path)
	if err != nil {
		return "", fmt.Errorf("get CA cert: %w", err)
	}
	result, err := decodeResponse[string](resp)
	if err != nil {
		return "", err
	}
	return result.Data, nil
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
// chain=true fetches the full certificate chain; needRoot=false excludes the root CA.
func (c *Client) GetUserSSLCert(ctx context.Context, uuid string, chain, needRoot bool) (string, error) {
	path := fmt.Sprintf("/api/v1/user/cert/ssl/%s/cer?isChain=%v&needRootCa=%v", uuid, chain, needRoot)
	resp, err := c.get(ctx, path)
	if err != nil {
		return "", fmt.Errorf("get SSL cert: %w", err)
	}
	result, err := decodeResponse[string](resp)
	if err != nil {
		return "", err
	}
	return result.Data, nil
}

// GetUserSSLPrivKey retrieves the encrypted private key.
// The API returns the private key as a base64-encoded PEM string in the data field.
func (c *Client) GetUserSSLPrivKey(ctx context.Context, uuid, password string) (string, error) {
	resp, err := c.post(ctx, fmt.Sprintf("/api/v1/user/cert/ssl/%s/privkey", uuid), GetPrivKeyRequest{Password: password})
	if err != nil {
		return "", fmt.Errorf("get SSL privkey: %w", err)
	}
	result, err := decodeResponse[string](resp)
	if err != nil {
		return "", err
	}
	return result.Data, nil
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
// The cert argument must already be base64-encoded (as returned by the cert fetch endpoints).
func (c *Client) AnalyzeCert(ctx context.Context, cert string) (*CertAnalysis, error) {
	resp, err := c.post(ctx, "/api/v1/user/cert/analyze", AnalyzeCertRequest{Cert: cert})
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
// The privKey argument must already be base64-encoded (as returned by the privkey fetch endpoints).
func (c *Client) AnalyzePrivKey(ctx context.Context, privKey, password string) (*PrivKeyAnalysis, error) {
	resp, err := c.post(ctx, "/api/v1/user/cert/privkey/analyze", AnalyzePrivKeyRequest{PrivKey: privKey, Password: password})
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

// CountUserSSLCerts returns the number of SSL certs owned by the current user.
func (c *Client) CountUserSSLCerts(ctx context.Context) (int64, error) {
resp, err := c.get(ctx, "/api/v1/user/cert/ssl/count")
if err != nil {
return 0, fmt.Errorf("count user SSL certs: %w", err)
}
result, err := decodeResponse[int64](resp)
if err != nil {
return 0, err
}
return result.Data, nil
}

// CountUserCAs returns the number of CAs bound to the current user.
func (c *Client) CountUserCAs(ctx context.Context) (int64, error) {
resp, err := c.get(ctx, "/api/v1/user/cert/ca/count")
if err != nil {
return 0, fmt.Errorf("count user CAs: %w", err)
}
result, err := decodeResponse[int64](resp)
if err != nil {
return 0, err
}
return result.Data, nil
}
