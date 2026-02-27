package api

import (
	"context"
	"fmt"
)

// Ping checks if the server is reachable.
func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.get(ctx, "/api/v1/test/ping")
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	_, err = decodeResponse[any](resp)
	return err
}
