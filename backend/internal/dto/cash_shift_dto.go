package dto

// OpenCashShiftRequest defines request DTO for opening a cash drawer shift.
type OpenCashShiftRequest struct {
	OpeningCash float64 `json:"opening_cash" binding:"required,gte=0"`
	Notes       string  `json:"notes"`
}

// OpenCashShiftResponse defines response DTO when a shift is opened.
type OpenCashShiftResponse struct {
	Message string `json:"message"`
	ShiftID string `json:"shift_id"`
}

// CloseCashShiftRequest defines request DTO for closing a cash drawer shift.
type CloseCashShiftRequest struct {
	ClosingCash float64 `json:"closing_cash" binding:"required,gte=0"`
	Notes       string  `json:"notes"`
}

// CloseCashShiftResponse defines structured response DTO when a shift is closed.
type CloseCashShiftResponse struct {
	Message      string  `json:"message"`
	OpeningCash  float64 `json:"opening_cash"`
	TotalSales   float64 `json:"total_sales"`
	ExpectedCash float64 `json:"expected_cash"`
	ClosingCash  float64 `json:"closing_cash"`
	Difference   float64 `json:"difference"`
}
