package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// ContentClassificationLabelsService provides access to content classification labels.
type ContentClassificationLabelsService struct {
	client *Client
}

// ContentClassificationLabel describes one Twitch content classification label.
type ContentClassificationLabel struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

// GetContentClassificationLabelsResponse is the typed response for Get Content Classification Labels.
type GetContentClassificationLabelsResponse struct {
	Data []ContentClassificationLabel `json:"data"`
}

// GetContentClassificationLabelsParams controls optional localization for label responses.
type GetContentClassificationLabelsParams struct {
	Locale string
}

// UnmarshalJSON accepts both the flat documented data array and a nested content_classification_labels wrapper.
func (r *GetContentClassificationLabelsResponse) UnmarshalJSON(data []byte) error {
	type alias GetContentClassificationLabelsResponse

	var direct alias
	if err := json.Unmarshal(data, &direct); err == nil {
		if len(direct.Data) == 0 {
			*r = GetContentClassificationLabelsResponse(direct)
			return nil
		}
		for _, label := range direct.Data {
			if label.ID != "" || label.Description != "" || label.Name != "" {
				*r = GetContentClassificationLabelsResponse(direct)
				return nil
			}
		}
	}

	var wrapped struct {
		Data []struct {
			ContentClassificationLabels []ContentClassificationLabel `json:"content_classification_labels"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	for _, item := range wrapped.Data {
		r.Data = append(r.Data, item.ContentClassificationLabels...)
	}
	return nil
}

// Get fetches the supported content classification labels.
func (s *ContentClassificationLabelsService) Get(ctx context.Context, params ...GetContentClassificationLabelsParams) (*GetContentClassificationLabelsResponse, *Response, error) {
	query := url.Values{}
	if len(params) > 0 && params[0].Locale != "" {
		query.Set("locale", params[0].Locale)
	}

	var resp GetContentClassificationLabelsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/content_classification_labels",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
