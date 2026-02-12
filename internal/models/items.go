package models

type ItemResponse struct {
	MarketHashName      string   `json:"market_hash_name"`
	Currency            string   `json:"currency"`
	MinPriceTradable    *float64 `json:"min_price_tradable"`
	MinPriceNonTradable *float64 `json:"min_price_non_tradable"`
}

type GetItemsResponse struct {
	RetryAfterSeconds uint64          `json:"-"`
	Items             []*ItemResponse `json:"items,omitempty"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Payload any    `json:"payload,omitempty"`
}
