package datatype

type CreateAccountRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

type UpdateAccountRequest struct {
	Name   *string `json:"name,omitempty"`
	Status *string `json:"status,omitempty"` // allowed: "active", "closed"
}

type CreateAccountForUserRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
}

type UpdateUserRequest struct {
	ID       int32   `json:"id"`
	Name     *string `json:"name"`
	Nickname *string `json:"nickname"`
}

type CreateBankAccountRequest struct {
	AccountID     int32  `json:"account_id"`
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
	BranchName    string `json:"branch_name,omitempty"`
	SwiftCode     string `json:"swift_code,omitempty"`
}

type UpdateBankAccountRequest struct {
	ID            int32   `json:"id"`
	AccountNumber *string `json:"account_number,omitempty"`
	BankName      *string `json:"bank_name,omitempty"`
	BranchName    *string `json:"branch_name,omitempty"`
	SwiftCode     *string `json:"swift_code,omitempty"`
}

// CreateLedgerRequest for bookkeeping.ledger.go/CreateLedger
type CreateLedgerRequest struct {
	AccountID int32  `json:"account_id"`
	Date      string `json:"date"`
	Type      string `json:"type"`
	Amount    string `json:"amount"` // decimal string
	Note      string `json:"note"`
}

// UpdateLedgerRequest for bookkeeping.ledger.go/UpdateLedger
type UpdateLedgerRequest struct {
	ID     int32   `json:"id"`
	Date   *string `json:"date,omitempty"`
	Type   *string `json:"type,omitempty"`
	Amount *string `json:"amount,omitempty"` // decimal string
	Note   *string `json:"note,omitempty"`
}

type CreateRecurringTransactionRequest struct {
	AccountID   int32   `json:"account_id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Amount      string  `json:"amount"` // decimal as string
	Note        string  `json:"note"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
	RecurType   string  `json:"recur_type"`
	Frequency   int     `json:"frequency"`
	DayOfWeek   *int    `json:"day_of_week,omitempty"`
	DayOfMonth  *int    `json:"day_of_month,omitempty"`
	MonthOfYear *int    `json:"month_of_year,omitempty"`
}

type UpdateRecurringTransactionRequest struct {
	ID          int32   `json:"id"`
	Name        *string `json:"name,omitempty"`
	Type        *string `json:"type,omitempty"`
	Amount      *string `json:"amount,omitempty"`
	Note        *string `json:"note,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
	RecurType   *string `json:"recur_type,omitempty"`
	Status      *string `json:"status,omitempty"`
	Frequency   *int    `json:"frequency,omitempty"`
	DayOfWeek   *int    `json:"day_of_week,omitempty"`
	DayOfMonth  *int    `json:"day_of_month,omitempty"`
	MonthOfYear *int    `json:"month_of_year,omitempty"`
}
