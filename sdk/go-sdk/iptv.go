// sdk/go-sdk/iptv.go
package nexus

import (
	"context"
	"fmt"
)

type IPTVService struct {
	client *Client
}

func (s *IPTVService) GetChannels(ctx context.Context, propertyID string) ([]*Channel, *Response, error) {
	req, err := s.client.newRequest(ctx, "GET", fmt.Sprintf("/iptv/channels?property_id=%s", propertyID), nil)
	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Data []*Channel `json:"data"`
	}
	resp, err := s.client.do(req, &result)
	if err != nil {
		return nil, resp, err
	}

	return result.Data, resp, nil
}

func (s *IPTVService) StartStream(ctx context.Context, req *StreamRequest) (*StreamSession, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", "/iptv/streams", req)
	if err != nil {
		return nil, nil, err
	}

	var session StreamSession
	resp, err := s.client.do(r, &session)
	if err != nil {
		return nil, resp, err
	}

	return &session, resp, nil
}

func (s *IPTVService) StopStream(ctx context.Context, streamID string) (*Response, error) {
	req, err := s.client.newRequest(ctx, "DELETE", fmt.Sprintf("/iptv/streams/%s", streamID), nil)
	if err != nil {
		return nil, err
	}

	return s.client.do(req, nil)
}

func (s *IPTVService) GetEPG(ctx context.Context, channelID string, date string) (*EPG, *Response, error) {
	req, err := s.client.newRequest(ctx, "GET", fmt.Sprintf("/iptv/epg/%s?date=%s", channelID, date), nil)
	if err != nil {
		return nil, nil, err
	}

	var epg EPG
	resp, err := s.client.do(req, &epg)
	if err != nil {
		return nil, resp, err
	}

	return &epg, resp, nil
}

type Channel struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Number      int    `json:"number"`
	Category    string `json:"category"`
	Language    string `json:"language"`
	HD          bool   `json:"hd"`
	LogoURL     string `json:"logo_url,omitempty"`
}

type StreamRequest struct {
	RoomID      string `json:"room_id"`
	ChannelID   int    `json:"channel_id"`
	DeviceType  string `json:"device_type,omitempty"`
	Profile     string `json:"profile,omitempty"` // 1080p, 720p, 480p, 360p
}

type StreamSession struct {
	ID          string `json:"id"`
	RoomID      string `json:"room_id"`
	ChannelID   int    `json:"channel_id"`
	ManifestURL string `json:"manifest_url"`
	Token       string `json:"token"`
	ExpiresAt   string `json:"expires_at"`
}

type EPG struct {
	ChannelID   int         `json:"channel_id"`
	Date        string      `json:"date"`
	Programs    []*Program  `json:"programs"`
}

type Program struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Genre       string `json:"genre,omitempty"`
	Rating      string `json:"rating,omitempty"`
}
