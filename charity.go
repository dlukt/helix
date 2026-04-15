package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// CharityService provides access to Twitch charity APIs.
type CharityService struct {
	client *Client
}

// CharityAmount describes a monetary amount in a charity campaign.
type CharityAmount struct {
	Value         int    `json:"value"`
	DecimalPlaces int    `json:"decimal_places"`
	Currency      string `json:"currency"`
}

// GetCharityCampaignParams identifies the broadcaster whose active charity campaign to fetch.
type GetCharityCampaignParams struct {
	BroadcasterID string
}

// GetCharityCampaignDonationsParams filters Get Charity Campaign Donations requests.
type GetCharityCampaignDonationsParams struct {
	CursorParams
	BroadcasterID string
}

// CharityCampaign describes an active charity campaign.
type CharityCampaign struct {
	ID                 string        `json:"id"`
	BroadcasterID      string        `json:"broadcaster_id"`
	BroadcasterLogin   string        `json:"broadcaster_login"`
	BroadcasterName    string        `json:"broadcaster_name"`
	CharityName        string        `json:"charity_name"`
	CharityDescription string        `json:"charity_description"`
	CharityLogo        string        `json:"charity_logo"`
	CharityWebsite     string        `json:"charity_website"`
	CurrentAmount      CharityAmount `json:"current_amount"`
	TargetAmount       CharityAmount `json:"target_amount"`
}

// GetCharityCampaignResponse is the typed response for Get Charity Campaign.
type GetCharityCampaignResponse struct {
	Data []CharityCampaign `json:"data"`
}

// CharityCampaignDonation describes one donation made to the active charity campaign.
type CharityCampaignDonation struct {
	ID               string        `json:"id"`
	CampaignID       string        `json:"campaign_id"`
	BroadcasterID    string        `json:"broadcaster_id"`
	BroadcasterLogin string        `json:"broadcaster_login"`
	BroadcasterName  string        `json:"broadcaster_name"`
	UserID           string        `json:"user_id"`
	UserLogin        string        `json:"user_login"`
	UserName         string        `json:"user_name"`
	Amount           CharityAmount `json:"amount"`
	CreatedAt        time.Time     `json:"created_at"`
}

// GetCharityCampaignDonationsResponse is the typed response for Get Charity Campaign Donations.
type GetCharityCampaignDonationsResponse struct {
	Data       []CharityCampaignDonation `json:"data"`
	Pagination Pagination                `json:"pagination"`
}

// GetCampaign fetches the broadcaster's active charity campaign.
func (s *CharityService) GetCampaign(ctx context.Context, params GetCharityCampaignParams) (*GetCharityCampaignResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)

	var resp GetCharityCampaignResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/charity/campaigns",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetDonations fetches donations made to the broadcaster's active charity campaign.
func (s *CharityService) GetDonations(ctx context.Context, params GetCharityCampaignDonationsParams) (*GetCharityCampaignDonationsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	addCursorParams(query, params.CursorParams)

	var resp GetCharityCampaignDonationsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/charity/donations",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
