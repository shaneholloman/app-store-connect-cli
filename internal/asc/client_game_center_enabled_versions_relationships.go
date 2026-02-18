package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// GetGameCenterEnabledVersionCompatibleVersionsRelationships retrieves compatible enabled version linkages.
func (c *Client) GetGameCenterEnabledVersionCompatibleVersionsRelationships(ctx context.Context, enabledVersionID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	return c.getGameCenterEnabledVersionLinkages(ctx, enabledVersionID, "compatibleVersions", opts...)
}

// AddGameCenterEnabledVersionCompatibleVersions adds compatible enabled versions.
func (c *Client) AddGameCenterEnabledVersionCompatibleVersions(ctx context.Context, enabledVersionID string, compatibleIDs []string) error {
	payload := RelationshipRequest{
		Data: buildRelationshipData(ResourceTypeGameCenterEnabledVersions, compatibleIDs),
	}
	body, err := BuildRequestBody(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v1/gameCenterEnabledVersions/%s/relationships/compatibleVersions", strings.TrimSpace(enabledVersionID))
	_, err = c.do(ctx, http.MethodPost, path, body)
	return err
}

// RemoveGameCenterEnabledVersionCompatibleVersions removes compatible enabled versions.
func (c *Client) RemoveGameCenterEnabledVersionCompatibleVersions(ctx context.Context, enabledVersionID string, compatibleIDs []string) error {
	payload := RelationshipRequest{
		Data: buildRelationshipData(ResourceTypeGameCenterEnabledVersions, compatibleIDs),
	}
	body, err := BuildRequestBody(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v1/gameCenterEnabledVersions/%s/relationships/compatibleVersions", strings.TrimSpace(enabledVersionID))
	_, err = c.do(ctx, http.MethodDelete, path, body)
	return err
}

// UpdateGameCenterEnabledVersionCompatibleVersionsRelationship replaces the compatibleVersions relationship.
func (c *Client) UpdateGameCenterEnabledVersionCompatibleVersionsRelationship(ctx context.Context, enabledVersionID string, compatibleIDs []string) error {
	payload := RelationshipRequest{
		Data: buildRelationshipData(ResourceTypeGameCenterEnabledVersions, compatibleIDs),
	}
	body, err := BuildRequestBody(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v1/gameCenterEnabledVersions/%s/relationships/compatibleVersions", strings.TrimSpace(enabledVersionID))
	_, err = c.do(ctx, http.MethodPatch, path, body)
	return err
}

func (c *Client) getGameCenterEnabledVersionLinkages(ctx context.Context, enabledVersionID, relationship string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	query := &linkagesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	enabledVersionID = strings.TrimSpace(enabledVersionID)
	if query.nextURL == "" && enabledVersionID == "" {
		return nil, fmt.Errorf("enabledVersionID is required")
	}

	path := fmt.Sprintf("/v1/gameCenterEnabledVersions/%s/relationships/%s", enabledVersionID, relationship)
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("gameCenterEnabledVersionRelationships: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildLinkagesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response LinkagesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}
