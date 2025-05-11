package bookkeeping

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/omegaatt36/bookly/domain"
)

// CreateRecurringTransaction creates a new recurring transaction
func (s *Service) CreateRecurringTransaction(ctx context.Context, request domain.CreateRecurringTransactionRequest) (*domain.RecurringTransaction, error) {
	if s.recurringTransactionRepo == nil || s.reminderRepo == nil {
		return nil, ErrRecurringRepositoriesNotSet
	}

	transaction, err := s.recurringTransactionRepo.CreateRecurringTransaction(ctx, request)
	if err != nil {
		return nil, err
	}

	// Create initial reminder if needed (3 days before first due date)
	reminderDate := calculateReminderDate(transaction.NextDue)
	if reminderDate.After(time.Now()) {
		_, err = s.reminderRepo.CreateReminder(ctx, transaction.ID, reminderDate)
		if err != nil {
			slog.Warn("failed to create reminder for recurring transaction",
				"transaction_id", transaction.ID,
				"error", err)
		}
	}

	return transaction, nil
}

// GetRecurringTransaction gets a recurring transaction by ID
func (s *Service) GetRecurringTransaction(ctx context.Context, id string) (*domain.RecurringTransaction, error) {
	return s.recurringTransactionRepo.GetRecurringTransactionByID(ctx, id)
}

// GetRecurringTransactionsByUserID gets all recurring transactions for a user
func (s *Service) GetRecurringTransactionsByUserID(ctx context.Context, userID string) ([]*domain.RecurringTransaction, error) {
	return s.recurringTransactionRepo.GetRecurringTransactionsByUserID(ctx, userID)
}

// UpdateRecurringTransaction updates a recurring transaction
func (s *Service) UpdateRecurringTransaction(ctx context.Context, request domain.UpdateRecurringTransactionRequest) (*domain.RecurringTransaction, error) {
	return s.recurringTransactionRepo.UpdateRecurringTransaction(ctx, request)
}

// DeleteRecurringTransaction deletes a recurring transaction
func (s *Service) DeleteRecurringTransaction(ctx context.Context, id string) error {
	if s.recurringTransactionRepo == nil {
		return ErrRecurringRepositoriesNotSet
	}
	return s.recurringTransactionRepo.DeleteRecurringTransaction(ctx, id)
}

// GetReminders gets reminders for a recurring transaction
func (s *Service) GetReminders(ctx context.Context, recurringTransactionID string) ([]*domain.Reminder, error) {
	return s.reminderRepo.GetRemindersByRecurringTransactionID(ctx, recurringTransactionID)
}

// GetActiveRemindersByUserID gets all active reminders for a user
func (s *Service) GetActiveRemindersByUserID(ctx context.Context, userID string) ([]*domain.Reminder, error) {
	return s.reminderRepo.GetActiveRemindersByUserID(ctx, userID, time.Now())
}

// GetReminderByID gets a reminder by ID
func (s *Service) GetReminderByID(ctx context.Context, id string) (*domain.Reminder, error) {
	return s.reminderRepo.GetReminderByID(ctx, id)
}

// MarkReminderAsRead marks a reminder as read
func (s *Service) MarkReminderAsRead(ctx context.Context, id string) (*domain.Reminder, error) {
	return s.reminderRepo.MarkReminderAsRead(ctx, id)
}

// ProcessDueTransactions processes all due recurring transactions
func (s *Service) ProcessDueTransactions(ctx context.Context) error {
	if s.recurringTransactionRepo == nil || s.reminderRepo == nil || s.ledgerRepo == nil {
		return ErrRecurringRepositoriesNotSet
	}

	now := time.Now()

	// Get all active and due transactions
	dueTransactions, err := s.recurringTransactionRepo.GetActiveRecurringTransactionsDue(ctx, now)
	if err != nil {
		return err
	}

	for _, transaction := range dueTransactions {
		// Create ledger entry based on recurring transaction
		ledgerReq := domain.CreateLedgerRequest{
			AccountID: transaction.AccountID,
			Date:      now,
			Type:      transaction.Type,
			Amount:    transaction.Amount,
			Note:      transaction.Note + " (Recurring: " + transaction.Name + ")",
		}

		_, err := s.ledgerRepo.CreateLedger(ledgerReq)
		if err != nil {
			slog.Error("failed to create ledger entry for recurring transaction",
				"transaction_id", transaction.ID,
				"error", err)
			continue
		}

		// Calculate next due date
		nextDue := calculateNextDueDate(
			transaction.NextDue,
			transaction.RecurType,
			transaction.Frequency,
			transaction.DayOfWeek,
			transaction.DayOfMonth,
			transaction.MonthOfYear,
		)

		// Check if this was the last occurrence
		if transaction.EndDate != nil && nextDue.After(*transaction.EndDate) {
			// Mark as completed
			completedStatus := domain.RecurrenceStatusCompleted
			updateReq := domain.UpdateRecurringTransactionRequest{
				ID:     transaction.ID,
				Status: &completedStatus,
			}
			_, err := s.recurringTransactionRepo.UpdateRecurringTransaction(ctx, updateReq)
			if err != nil {
				slog.Error("failed to mark recurring transaction as completed",
					"transaction_id", transaction.ID,
					"error", err)
			}
		} else {
			// Update last executed and next due date
			_, err := s.recurringTransactionRepo.UpdateRecurringTransactionExecution(ctx, transaction.ID, now, nextDue)
			if err != nil {
				slog.Error("failed to update recurring transaction execution",
					"transaction_id", transaction.ID,
					"error", err)
				continue
			}

			// Create next reminder
			reminderDate := calculateReminderDate(nextDue)
			_, err = s.reminderRepo.CreateReminder(ctx, transaction.ID, reminderDate)
			if err != nil {
				slog.Error("failed to create reminder for next execution",
					"transaction_id", transaction.ID,
					"error", err)
			}
		}
	}

	return nil
}

// GetUpcomingReminders gets upcoming reminders for a user within the next week
func (s *Service) GetUpcomingReminders(ctx context.Context, userID string) ([]*domain.Reminder, error) {

	now := time.Now()
	oneWeekLater := now.AddDate(0, 0, 7)

	// Use repository that supports date range
	if repo, ok := s.reminderRepo.(interface {
		GetUpcomingReminders(ctx context.Context, userID string, start, end time.Time) ([]*domain.Reminder, error)
	}); ok {
		return repo.GetUpcomingReminders(ctx, userID, now, oneWeekLater)
	}

	// Fallback to standard method
	return s.reminderRepo.GetActiveRemindersByUserID(ctx, userID, oneWeekLater)
}

// ErrRecurringRepositoriesNotSet is returned when trying to use recurring features without setting up repositories
var ErrRecurringRepositoriesNotSet = errors.New("recurring repositories not set")

// Helper functions
func calculateReminderDate(dueDate time.Time) time.Time {
	// Default: remind 3 days before due date
	return dueDate.AddDate(0, 0, -3)
}

func calculateNextDueDate(
	lastDue time.Time,
	recurType domain.RecurrenceType,
	frequency int,
	_ *int,
	_ *int,
	_ *int,
) time.Time {
	// Implementation in RecurringTransactionRepository
	// This is a simplified version for future calculations
	switch recurType {
	case domain.RecurrenceTypeDaily:
		return lastDue.AddDate(0, 0, frequency)

	case domain.RecurrenceTypeWeekly:
		return lastDue.AddDate(0, 0, 7*frequency)

	case domain.RecurrenceTypeBiweekly:
		return lastDue.AddDate(0, 0, 14*frequency)

	case domain.RecurrenceTypeMonthly:
		return lastDue.AddDate(0, frequency, 0)

	case domain.RecurrenceTypeQuarterly:
		return lastDue.AddDate(0, 3*frequency, 0)

	case domain.RecurrenceTypeYearly:
		return lastDue.AddDate(frequency, 0, 0)

	default:
		return lastDue.AddDate(0, 1, 0)
	}
}
