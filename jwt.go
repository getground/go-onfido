package onfido

import (
	"bytes"
	"context"
	"encoding/json"
)

// SdkToken represents the response for a request for a JWT token
type SdkToken struct {
	ApplicantID string `json:"applicant_id,omitempty"`
	Token       string `json:"token,omitempty"`
}

// NewSdkToken returns a JWT token to used by the Javascript SDK
func (c *Client) NewSdkToken(ctx context.Context, id string) (*SdkToken, error) {
	t := &SdkToken{
		ApplicantID: id,
	}
	jsonStr, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest("POST", "/sdk_token", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	var resp SdkToken
	if _, err := c.do(ctx, req, &resp); err != nil {
		return nil, err
	}
	t.Token = resp.Token
	return t, err
}
