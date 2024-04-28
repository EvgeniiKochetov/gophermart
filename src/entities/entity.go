package entities

import "time"

type Order struct {
	Number     int       `json:"number,omitempty"`
	Status     string    `json:"status,omitempty"`
	Accrual    int       `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploadedAt,omitempty"`
}

type Balance struct {
	Current   int `json:"balance"`
	WithDrawn int `json:"withDrawn"`
}

type WithDraw struct {
	Order      int       `json:"order"`
	Sum        int       `json:"sum"`
	UploadedAt time.Time `json:"processed_at,omitempty"`
}
