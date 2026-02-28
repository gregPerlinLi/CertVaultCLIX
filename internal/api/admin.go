package api

import (
	"context"
	"fmt"
)

// ListAdminUsers lists all users (admin only).
func (c *Client) ListAdminUsers(ctx context.Context, page, size int) (*PageDTO[AdminUser], error) {
	path := fmt.Sprintf("/api/v1/admin/users?page=%d&limit=%d", page, size)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list admin users: %w", err)
	}
	result, err := decodeResponse[PageDTO[AdminUser]](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// ListAdminCAs lists all CA certificates (admin only).
func (c *Client) ListAdminCAs(ctx context.Context, page, size int) (*PageDTO[CACert], error) {
	path := fmt.Sprintf("/api/v1/admin/cert/ca?page=%d&limit=%d", page, size)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list admin CAs: %w", err)
	}
	result, err := decodeResponse[PageDTO[CACert]](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetAdminCACert gets the CA certificate PEM.
func (c *Client) GetAdminCACert(ctx context.Context, uuid string, chain, needRoot bool) (string, error) {
	path := fmt.Sprintf("/api/v1/admin/cert/ca/%s/cer?isChain=%v&needRootCa=%v", uuid, chain, needRoot)
	resp, err := c.get(ctx, path)
	if err != nil {
		return "", fmt.Errorf("get admin CA cert: %w", err)
	}
	result, err := decodeResponse[string](resp)
	if err != nil {
		return "", err
	}
	return result.Data, nil
}

// GetAdminCAPrivKey gets the CA private key.
// The API returns the private key as a base64-encoded PEM string in the data field.
func (c *Client) GetAdminCAPrivKey(ctx context.Context, uuid, password string) (string, error) {
	resp, err := c.post(ctx, fmt.Sprintf("/api/v1/admin/cert/ca/%s/privkey", uuid), GetPrivKeyRequest{Password: password})
	if err != nil {
		return "", fmt.Errorf("get admin CA privkey: %w", err)
	}
	result, err := decodeResponse[string](resp)
	if err != nil {
		return "", err
	}
	return result.Data, nil
}

// UpdateAdminCAComment updates the CA comment.
func (c *Client) UpdateAdminCAComment(ctx context.Context, uuid, comment string) error {
	resp, err := c.patch(ctx, fmt.Sprintf("/api/v1/admin/cert/ca/%s/comment", uuid), UpdateCommentRequest{Comment: comment})
	if err != nil {
		return fmt.Errorf("update CA comment: %w", err)
	}
	_, err = decodeResponse[any](resp)
	return err
}

// ToggleAdminCAAvailable toggles the CA availability.
func (c *Client) ToggleAdminCAAvailable(ctx context.Context, uuid string, available bool) error {
	resp, err := c.patch(ctx, fmt.Sprintf("/api/v1/admin/cert/ca/%s/available", uuid), ToggleAvailableRequest{Available: available})
	if err != nil {
		return fmt.Errorf("toggle CA availability: %w", err)
	}
	_, err = decodeResponse[any](resp)
	return err
}

// ImportAdminCA imports a CA certificate.
func (c *Client) ImportAdminCA(ctx context.Context, req ImportCACertRequest) (*CACert, error) {
	resp, err := c.post(ctx, "/api/v1/admin/cert/ca/import", req)
	if err != nil {
		return nil, fmt.Errorf("import CA: %w", err)
	}
	result, err := decodeResponse[CACert](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// RequestAdminCA creates a new CA certificate.
func (c *Client) RequestAdminCA(ctx context.Context, req RequestCACertRequest) (*CACert, error) {
	resp, err := c.post(ctx, "/api/v1/admin/cert/ca", req)
	if err != nil {
		return nil, fmt.Errorf("request CA: %w", err)
	}
	result, err := decodeResponse[CACert](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// RenewAdminCA renews a CA certificate.
func (c *Client) RenewAdminCA(ctx context.Context, uuid string, req RenewCACertRequest) (*CACert, error) {
	resp, err := c.put(ctx, fmt.Sprintf("/api/v1/admin/cert/ca/%s", uuid), req)
	if err != nil {
		return nil, fmt.Errorf("renew CA: %w", err)
	}
	result, err := decodeResponse[CACert](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// DeleteAdminCA deletes a CA certificate.
func (c *Client) DeleteAdminCA(ctx context.Context, uuid string) error {
	resp, err := c.delete(ctx, fmt.Sprintf("/api/v1/admin/cert/ca/%s", uuid))
	if err != nil {
		return fmt.Errorf("delete CA: %w", err)
	}
	_, err = decodeResponse[any](resp)
	return err
}

// BindUsersToCA binds a user to a CA.
// Uses POST /api/v1/admin/cert/ca/bind/create with CABindingDTO.
func (c *Client) BindUsersToCA(ctx context.Context, caUUID string, usernames []string) error {
	for _, username := range usernames {
		resp, err := c.post(ctx, "/api/v1/admin/cert/ca/bind/create", CABindingDTO{CaUUID: caUUID, Username: username})
		if err != nil {
			return fmt.Errorf("bind user %q to CA: %w", username, err)
		}
		if _, err = decodeResponse[any](resp); err != nil {
			return fmt.Errorf("bind user %q to CA: %w", username, err)
		}
	}
	return nil
}

// UnbindUsersFromCA unbinds a user from a CA.
// Uses POST /api/v1/admin/cert/ca/bind/delete with CABindingDTO.
func (c *Client) UnbindUsersFromCA(ctx context.Context, caUUID string, usernames []string) error {
	for _, username := range usernames {
		resp, err := c.post(ctx, "/api/v1/admin/cert/ca/bind/delete", CABindingDTO{CaUUID: caUUID, Username: username})
		if err != nil {
			return fmt.Errorf("unbind user %q from CA: %w", username, err)
		}
		if _, err = decodeResponse[any](resp); err != nil {
			return fmt.Errorf("unbind user %q from CA: %w", username, err)
		}
	}
	return nil
}

// GetBoundUsers gets users bound to a CA.
// Uses GET /api/v1/admin/cert/ca/{uuid}/bind.
func (c *Client) GetBoundUsers(ctx context.Context, uuid string, page, size int) (*PageDTO[AdminUser], error) {
	path := fmt.Sprintf("/api/v1/admin/cert/ca/%s/bind?page=%d&limit=%d", uuid, page, size)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("get bound users: %w", err)
	}
	result, err := decodeResponse[PageDTO[AdminUser]](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// GetUnboundUsers gets users not yet bound to a CA.
// Uses GET /api/v1/admin/cert/ca/{uuid}/bind/not.
func (c *Client) GetUnboundUsers(ctx context.Context, uuid string, page, size int) (*PageDTO[AdminUser], error) {
	path := fmt.Sprintf("/api/v1/admin/cert/ca/%s/bind/not?page=%d&limit=%d", uuid, page, size)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("get unbound users: %w", err)
	}
	result, err := decodeResponse[PageDTO[AdminUser]](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// CountAdminUsers returns the total number of users (admin+).
func (c *Client) CountAdminUsers(ctx context.Context) (int64, error) {
resp, err := c.get(ctx, "/api/v1/admin/users/count")
if err != nil {
return 0, fmt.Errorf("count admin users: %w", err)
}
result, err := decodeResponse[int64](resp)
if err != nil {
return 0, err
}
return result.Data, nil
}

// CountAdminCAs returns the total number of CA certs (admin+).
func (c *Client) CountAdminCAs(ctx context.Context) (int64, error) {
resp, err := c.get(ctx, "/api/v1/admin/cert/ca/count")
if err != nil {
return 0, fmt.Errorf("count admin CAs: %w", err)
}
result, err := decodeResponse[int64](resp)
if err != nil {
return 0, err
}
return result.Data, nil
}
