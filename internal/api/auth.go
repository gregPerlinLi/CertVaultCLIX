package api

import (
	"context"
	"fmt"
)

// Login authenticates with the server and stores the session.
func (c *Client) Login(ctx context.Context, username, password string) error {
	resp, err := c.post(ctx, "/api/v1/auth/login", LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	result, err := decodeResponse[any](resp)
	if err != nil {
		return err
	}
	_ = result
	return nil
}

// Logout logs out the current session.
func (c *Client) Logout(ctx context.Context) error {
	resp, err := c.delete(ctx, "/api/v1/auth/logout")
	if err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	_, err = decodeResponse[any](resp)
	if err != nil {
		return err
	}
	c.session = ""
	return nil
}

// GetOIDCAuthURL returns the OIDC authorization URL.
func (c *Client) GetOIDCAuthURL(ctx context.Context) (string, error) {
	resp, err := c.get(ctx, "/api/v1/auth/oidc/authorization")
	if err != nil {
		return "", fmt.Errorf("oidc auth url: %w", err)
	}
	result, err := decodeResponse[string](resp)
	if err != nil {
		return "", err
	}
	return result.Data, nil
}
