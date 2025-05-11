//go:generate go-enum

package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// RecurrenceType represents the type of recurrence for a transaction
// ENUM(daily, weekly, biweekly, monthly, quarterly, yearly, custom)
type RecurrenceType string

// RecurrenceStatus represents the status of a recurring transaction
// ENUM(active, paused, completed, cancelled)
type RecurrenceStatus string

// RecurringTransaction represents a recurring transaction configuration
type RecurringTransaction struct {
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	UserID       string
	AccountID    string
	Name         string
	Type         LedgerType
	Amount       decimal.Decimal
	Note         string
	StartDate    time.Time
	EndDate      *time.Time
	RecurType    RecurrenceType
	Status       RecurrenceStatus
	Frequency    int        // How often the recurrence happens (e.g., every 2 weeks)
	DayOfWeek    *int       // 0-6 (Sunday-Saturday) for weekly recurrences
	DayOfMonth   *int       // 1-31 for monthly recurrences
	MonthOfYear  *int       // 1-12 for yearly recurrences
	LastExecuted *time.Time // When the transaction was last created
	NextDue      time.Time  // When the next transaction is due
}

// Reminder represents a reminder for a recurring transaction
type Reminder struct {
	ID                     string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	RecurringTransactionID string
	ReminderDate           time.Time
	IsRead                 bool
	ReadAt                 *time.Time
}

// CreateRecurringTransactionRequest defines the request to create a recurring transaction
type CreateRecurringTransactionRequest struct {
	UserID      string
	AccountID   string
	Name        string
	Type        LedgerType
	Amount      decimal.Decimal
	Note        string
	StartDate   time.Time
	EndDate     *time.Time
	RecurType   RecurrenceType
	Frequency   int
	DayOfWeek   *int
	DayOfMonth  *int
	MonthOfYear *int
}

// UpdateRecurringTransactionRequest defines the request to update a recurring transaction
type UpdateRecurringTransactionRequest struct {
	ID          string
	Name        *string
	Type        *LedgerType
	Amount      *decimal.Decimal
	Note        *string
	EndDate     *time.Time
	RecurType   *RecurrenceType
	Status      *RecurrenceStatus
	Frequency   *int
	DayOfWeek   *int
	DayOfMonth  *int
	MonthOfYear *int
}

// RecurringTransactionRepository represents a recurring transaction repository
type RecurringTransactionRepository interface {
	CreateRecurringTransaction(ctx context.Context, req CreateRecurringTransactionRequest) (*RecurringTransaction, error)
	GetRecurringTransactionByID(ctx context.Context, id string) (*RecurringTransaction, error)
	GetRecurringTransactionsByUserID(ctx context.Context, userID string) ([]*RecurringTransaction, error)
	GetActiveRecurringTransactionsDue(ctx context.Context, before time.Time) ([]*RecurringTransaction, error)
	UpdateRecurringTransaction(ctx context.Context, req UpdateRecurringTransactionRequest) (*RecurringTransaction, error)
	UpdateRecurringTransactionExecution(ctx context.Context, id string, lastExecuted, nextDue time.Time) (*RecurringTransaction, error)
	DeleteRecurringTransaction(ctx context.Context, id string) error
}

// ReminderRepository represents a reminder repository
type ReminderRepository interface {
	CreateReminder(ctx context.Context, recurringTransactionID string, reminderDate time.Time) (*Reminder, error)
	GetRemindersByRecurringTransactionID(ctx context.Context, recurringTransactionID string) ([]*Reminder, error)
	GetActiveRemindersByUserID(ctx context.Context, userID string, before time.Time) ([]*Reminder, error)
	GetReminderByID(ctx context.Context, id string) (*Reminder, error)
	MarkReminderAsRead(ctx context.Context, id string) (*Reminder, error)
}
