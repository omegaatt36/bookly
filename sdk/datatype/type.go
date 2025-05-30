package datatype

const TimeFormat = "2006-01-02 15:04:05"

type User struct {
	ID        int32  `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Name      string `json:"name"`
	Nickname  string `json:"nickname"`
	Disabled  bool   `json:"disabled"`
}

type Account struct {
	ID        int32   `json:"id"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	UserID    int32   `json:"user_id"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	Currency  string  `json:"currency"`
	Balance   string  `json:"balance"`
	DeletedAt *string `json:"deleted_at,omitempty"`
}

type BankAccount struct {
	ID            int32  `json:"id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	AccountID     int32  `json:"account_id"`
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
	BranchName    string `json:"branch_name,omitempty"`
	SwiftCode     string `json:"swift_code,omitempty"`
}

type Ledger struct {
	ID           int32   `json:"id"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	AccountID    int32   `json:"account_id"`
	Date         string  `json:"date"`
	Type         string  `json:"type"`
	Currency     string  `json:"currency"`
	Amount       string  `json:"amount"`
	Note         string  `json:"note"`
	Adjustable   bool    `json:"adjustable"`
	IsAdjustment bool    `json:"is_adjustment"`
	AdjustedFrom *int32  `json:"adjusted_from,omitempty"`
	IsVoided     bool    `json:"is_voided"`
	VoidedAt     *string `json:"voided_at,omitempty"`
}

type RecurringTransaction struct {
	ID             int32   `json:"id"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	UserID         int32   `json:"user_id"`
	AccountID      int32   `json:"account_id"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Amount         string  `json:"amount"`
	Note           string  `json:"note"`
	StartDate      string  `json:"start_date"`
	EndDate        *string `json:"end_date"`
	RecurrenceType string  `json:"recurrence_type"`
	Status         string  `json:"status"`
	Frequency      int     `json:"frequency"`
	DayOfWeek      *int    `json:"day_of_week,omitempty"`
	DayOfMonth     *int    `json:"day_of_month,omitempty"`
	MonthOfYear    *int    `json:"month_of_year,omitempty"`
	LastExecuted   *string `json:"last_executed"`
	NextDue        string  `json:"next_due"`
}

type Reminder struct {
	ID                     int32   `json:"id"`
	CreatedAt              string  `json:"created_at"`
	RecurringTransactionID int32   `json:"recurring_transaction_id"`
	ReminderDate           string  `json:"reminder_date"`
	IsRead                 bool    `json:"is_read"`
	ReadAt                 *string `json:"read_at,omitempty"`
}
