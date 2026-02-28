package api

import (
"context"
"fmt"
)

// ListAllSessions lists all sessions across all users (superadmin only).
func (c *Client) ListAllSessions(ctx context.Context, page, limit int) (*PageDTO[LoginRecord], error) {
path := fmt.Sprintf("/api/v1/superadmin/user/session?page=%d&limit=%d", page, limit)
resp, err := c.get(ctx, path)
if err != nil {
return nil, fmt.Errorf("list all sessions: %w", err)
}
result, err := decodeResponse[PageDTO[LoginRecord]](resp)
if err != nil {
return nil, err
}
return &result.Data, nil
}

// ListUserSessionsBySuperadmin lists sessions for a specific user (superadmin only).
func (c *Client) ListUserSessionsBySuperadmin(ctx context.Context, username string, page, limit int) (*PageDTO[LoginRecord], error) {
path := fmt.Sprintf("/api/v1/superadmin/user/session/%s?page=%d&limit=%d", username, page, limit)
resp, err := c.get(ctx, path)
if err != nil {
return nil, fmt.Errorf("list user sessions: %w", err)
}
result, err := decodeResponse[PageDTO[LoginRecord]](resp)
if err != nil {
return nil, err
}
return &result.Data, nil
}

// ForceLogoutUser force-logs out a user (superadmin only).
func (c *Client) ForceLogoutUser(ctx context.Context, username string) error {
resp, err := c.delete(ctx, fmt.Sprintf("/api/v1/superadmin/user/%s/logout", username))
if err != nil {
return fmt.Errorf("force logout user: %w", err)
}
_, err = decodeResponse[any](resp)
return err
}

// CreateUser creates a new user (superadmin only).
func (c *Client) CreateUser(ctx context.Context, req CreateUserRequest) (*AdminUser, error) {
resp, err := c.post(ctx, "/api/v1/superadmin/user", req)
if err != nil {
return nil, fmt.Errorf("create user: %w", err)
}
result, err := decodeResponse[AdminUser](resp)
if err != nil {
return nil, err
}
return &result.Data, nil
}

// BatchCreateUsers creates multiple users at once.
func (c *Client) BatchCreateUsers(ctx context.Context, users []CreateUserRequest) error {
resp, err := c.post(ctx, "/api/v1/superadmin/users/create", users)
if err != nil {
return fmt.Errorf("batch create users: %w", err)
}
_, err = decodeResponse[any](resp)
return err
}

// BatchDeleteUsers deletes multiple users at once.
func (c *Client) BatchDeleteUsers(ctx context.Context, usernames []string) error {
resp, err := c.post(ctx, "/api/v1/superadmin/users/delete", usernames)
if err != nil {
return fmt.Errorf("batch delete users: %w", err)
}
_, err = decodeResponse[any](resp)
return err
}

// UpdateSuperadminUser updates a user's info (superadmin only).
func (c *Client) UpdateSuperadminUser(ctx context.Context, username string, req UpdateSuperadminUserRequest) error {
resp, err := c.patch(ctx, fmt.Sprintf("/api/v1/superadmin/user/%s", username), req)
if err != nil {
return fmt.Errorf("update user: %w", err)
}
_, err = decodeResponse[any](resp)
return err
}

// UpdateUserRole updates a user's role (superadmin only).
func (c *Client) UpdateUserRole(ctx context.Context, req UpdateUserRoleRequest) error {
resp, err := c.patch(ctx, "/api/v1/superadmin/user/role", req)
if err != nil {
return fmt.Errorf("update user role: %w", err)
}
_, err = decodeResponse[any](resp)
return err
}

// DeleteSuperadminUser deletes a user (superadmin only).
func (c *Client) DeleteSuperadminUser(ctx context.Context, username string) error {
resp, err := c.delete(ctx, fmt.Sprintf("/api/v1/superadmin/user/%s", username))
if err != nil {
return fmt.Errorf("delete user: %w", err)
}
_, err = decodeResponse[any](resp)
return err
}

// CountAllSSLCerts returns the total number of SSL certs (superadmin only).
func (c *Client) CountAllSSLCerts(ctx context.Context) (int64, error) {
resp, err := c.get(ctx, "/api/v1/superadmin/cert/ssl/count")
if err != nil {
return 0, fmt.Errorf("count all SSL certs: %w", err)
}
result, err := decodeResponse[int64](resp)
if err != nil {
return 0, err
}
return result.Data, nil
}

// CountAllCAs returns the total number of CA certs (superadmin only).
func (c *Client) CountAllCAs(ctx context.Context) (int64, error) {
resp, err := c.get(ctx, "/api/v1/superadmin/cert/ca/count")
if err != nil {
return 0, fmt.Errorf("count all CAs: %w", err)
}
result, err := decodeResponse[int64](resp)
if err != nil {
return 0, err
}
return result.Data, nil
}
