package server

type CreateISARequest struct {
	UserID      string  `json:"user_id" binding:"required"`
	CashBalance float64 `json:"cash_balance" binding:"required"`
}

type CreateFundRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Type        string  `json:"type" binding:"required,oneof=Equity Bond Index Mixed"`
	RiskLevel   string  `json:"risk_level" binding:"required,oneof=Low Medium High"`
	Performance float64 `json:"performance" binding:"omitempty"`
	TotalAmount float64 `json:"total_amount" binding:"omitempty"`
}

type UpdateFundRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}
