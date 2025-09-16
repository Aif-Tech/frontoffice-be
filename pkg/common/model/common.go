package model

type AifResponse struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data"`
}

type Meta struct {
	Message   string `json:"message"`
	Total     any    `json:"total,omitempty"`
	Page      any    `json:"page,omitempty"`
	TotalPage any    `json:"total_page,omitempty"`
	Visible   any    `json:"visible,omitempty"`
	StartData any    `json:"start_data,omitempty"`
	EndData   any    `json:"end_data,omitempty"`
	Size      any    `json:"size,omitempty"`
}

type AifcoreAPIResponse[T any] struct {
	Success    bool   `json:"success"`
	Data       T      `json:"data"`
	Message    string `json:"message"`
	Meta       *Meta  `json:"meta,omitempty"`
	StatusCode int    `json:"-"`
}

type ProCatAPIResponse[T any] struct {
	Success         bool        `json:"success"`
	Data            T           `json:"data"`
	Input           interface{} `json:"input"`
	Message         string      `json:"message"`
	StatusCode      int         `json:"-"`
	PricingStrategy string      `json:"pricing_strategy"`
	TransactionId   string      `json:"transaction_id"`
	Date            string      `json:"datetime"`
}

type ScoreezyAPIResponse[T any] struct {
	Success      bool   `json:"success"`
	Data         *T     `json:"data"`
	Message      string `json:"message"`
	ErrorMessage string `json:"error_message,omitempty"`
	StatusCode   int    `json:"-"`
}
