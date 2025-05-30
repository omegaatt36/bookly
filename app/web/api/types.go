package api

import "time"

type User struct {
	ID        int32  `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Name      string `json:"name"`
	Nickname  string `json:"nickname"`
	Disabled  bool   `json:"disabled"`
}

type Account struct {
	ID        int32      `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	UserID    int32      `json:"user_id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	Currency  string     `json:"currency"`
	Balance   string     `json:"balance"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type Ledger struct {
	ID           int32      `json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	AccountID    int32      `json:"account_id"`
	Date         time.Time  `json:"date"`
	Type         string     `json:"type"`
	Currency     string     `json:"currency"`
	Amount       string     `json:"amount"`
	Note         string     `json:"note"`
	IsAdjustment bool       `json:"is_adjustment"`
	AdjustedFrom *int32     `json:"adjusted_from,omitempty"`
	IsVoided     bool       `json:"is_voided"`
	VoidedAt     *time.Time `json:"voided_at,omitempty"`
}

type RecurringTransaction struct {
	ID             int32      `json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	UserID         int32      `json:"user_id"`
	AccountID      int32      `json:"account_id"`
	Name           string     `json:"name"`
	Type           string     `json:"type"`
	Amount         string     `json:"amount"`
	Note           string     `json:"note"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	RecurrenceType string     `json:"recurrence_type"`
	Status         string     `json:"status"`
	Frequency      int        `json:"frequency"`
	DayOfWeek      *int       `json:"day_of_week,omitempty"`
	DayOfMonth     *int       `json:"day_of_month,omitempty"`
	MonthOfYear    *int       `json:"month_of_year,omitempty"`
	LastExecuted   *time.Time `json:"last_executed"`
	NextDue        time.Time  `json:"next_due"`
}

type Reminder struct {
	ID                     int32      `json:"id"`
	CreatedAt              time.Time  `json:"created_at"`
	RecurringTransactionID int32      `json:"recurring_transaction_id"`
	ReminderDate           time.Time  `json:"reminder_date"`
	IsRead                 bool       `json:"is_read"`
	ReadAt                 *time.Time `json:"read_at,omitempty"`
}
