package api

import (
	"context"
	"fmt"
)

// ListAllSessions lists all sessions across all users (superadmin only).
func (c *Client) ListAllSessions(ctx context.Context, page, size int) (*PageDTO[AllSession], error) {
	path := fmt.Sprintf("/api/v1/superadmin/user/session?page=%d&size=%d", page, size)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list all sessions: %w", err)
	}
	result, err := decodeResponse[PageDTO[AllSession]](resp)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// ListUserSessionsBySuperadmin lists sessions for a specific user.
func (c *Client) ListUserSessionsBySuperadmin(ctx context.Context, username string, page, size int) (*PageDTO[AllSession], error) {
	path := fmt.Sprintf("/api/v1/superadmin/user/session/%s?page=%d&size=%d", username, page, size)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list user sessions: %w", err)
	}
	result, err := decodeResponse[PageDTO[AllSession]](resp)
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
func (c *Client) BatchCreateUsers(ctx context.Context, req BatchCreateUsersRequest) error {
	resp, err := c.post(ctx, "/api/v1/superadmin/users/create", req)
	if err != nil {
		return fmt.Errorf("batch create users: %w", err)
	}
	_, err = decodeResponse[any](resp)
	return err
}

// BatchDeleteUsers deletes multiple users at once.
func (c *Client) BatchDeleteUsers(ctx context.Context, req BatchDeleteUsersRequest) error {
	resp, err := c.post(ctx, "/api/v1/superadmin/users/delete", req)
	if err != nil {
		return fmt.Errorf("batch delete users: %w", err)
	}
	_, err = decodeResponse[any](resp)
	return err
}

// UpdateUser updates a user's info (superadmin only).
func (c *Client) UpdateUser(ctx context.Context, username string, req UpdateUserRequest) error {
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
