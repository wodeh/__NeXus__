package domain

import "time"

type Property struct {
	ID        string                 `json:"id"`
	TenantID  string                 `json:"tenant_id"`
	Name      string                 `json:"name"`
	Code      string                 `json:"code"`
	Address   string                 `json:"address"`
	City      string                 `json:"city"`
	Country   string                 `json:"country"`
	Timezone  string                 `json:"timezone"`
	Status    string                 `json:"status"`
	Settings  map[string]interface{} `json:"settings"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type PropertySummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Status   string `json:"status"`
	City     string `json:"city"`
	RoomCount int   `json:"room_count"`
}
