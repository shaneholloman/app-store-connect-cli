package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// GetGameCenterMatchmakingRuleSetQueuesRelationships retrieves matchmaking queue linkages for a rule set.
func (c *Client) GetGameCenterMatchmakingRuleSetQueuesRelationships(ctx context.Context, ruleSetID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	return c.getGameCenterMatchmakingRuleSetLinkages(ctx, ruleSetID, "matchmakingQueues", opts...)
}

// GetGameCenterMatchmakingRuleSetRulesRelationships retrieves rule linkages for a rule set.
func (c *Client) GetGameCenterMatchmakingRuleSetRulesRelationships(ctx context.Context, ruleSetID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	return c.getGameCenterMatchmakingRuleSetLinkages(ctx, ruleSetID, "rules", opts...)
}

// GetGameCenterMatchmakingRuleSetTeamsRelationships retrieves team linkages for a rule set.
func (c *Client) GetGameCenterMatchmakingRuleSetTeamsRelationships(ctx context.Context, ruleSetID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	return c.getGameCenterMatchmakingRuleSetLinkages(ctx, ruleSetID, "teams", opts...)
}

func (c *Client) getGameCenterMatchmakingRuleSetLinkages(ctx context.Context, ruleSetID, relationship string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	query := &linkagesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	ruleSetID = strings.TrimSpace(ruleSetID)
	if query.nextURL == "" && ruleSetID == "" {
		return nil, fmt.Errorf("ruleSetID is required")
	}

	path := fmt.Sprintf("/v1/gameCenterMatchmakingRuleSets/%s/relationships/%s", ruleSetID, relationship)
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("gameCenterMatchmakingRuleSetRelationships: %w", err)
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
