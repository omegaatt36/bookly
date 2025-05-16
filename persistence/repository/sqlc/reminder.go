package sqlc

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/sqlcgen"

	"github.com/jackc/pgx/v5"
)

// CreateReminder creates a new reminder for a recurring transaction
func (r *Repository) CreateReminder(ctx context.Context, recurringTransactionID int32, reminderDate time.Time) (*domain.Reminder, error) {
	params := sqlcgen.CreateReminderParams{
		RecurringTransactionID: recurringTransactionID,
		ReminderDate:           pgtype.Timestamptz{Time: reminderDate, Valid: true},
	}

	result, err := r.querier.CreateReminder(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapToReminder(result), nil
}

// GetRemindersByRecurringTransactionID gets all reminders for a recurring transaction
func (r *Repository) GetRemindersByRecurringTransactionID(ctx context.Context, recurringTransactionID int32) ([]*domain.Reminder, error) {
	results, err := r.querier.GetRemindersByRecurringTransactionID(ctx, recurringTransactionID)
	if err != nil {
		return nil, err
	}

	reminders := make([]*domain.Reminder, len(results))
	for i, result := range results {
		reminders[i] = mapToReminder(result)
	}

	return reminders, nil
}

// GetActiveRemindersByUserID gets all active (unread) reminders for a user
func (r *Repository) GetActiveRemindersByUserID(ctx context.Context, userID int32, before time.Time) ([]*domain.Reminder, error) {
	results, err := r.querier.GetActiveRemindersByUserID(ctx, sqlcgen.GetActiveRemindersByUserIDParams{
		UserID:       userID,
		ReminderDate: pgtype.Timestamptz{Time: before, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	reminders := make([]*domain.Reminder, len(results))
	for i, result := range results {
		reminders[i] = mapToReminder(result)
	}

	return reminders, nil
}

// GetReminderByID gets a reminder by its ID
func (r *Repository) GetReminderByID(ctx context.Context, id int32) (*domain.Reminder, error) {
	result, err := r.querier.GetReminderByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return mapToReminder(result), nil
}

// MarkReminderAsRead marks a reminder as read
func (r *Repository) MarkReminderAsRead(ctx context.Context, id int32) (*domain.Reminder, error) {
	result, err := r.querier.MarkReminderAsRead(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapToReminder(result), nil
}

// DeleteReminder implements the domain.ReminderRepository interface for soft delete
func (r *Repository) DeleteReminder(ctx context.Context, id int32) error {
	if _, err := r.querier.DeleteReminder(ctx, id); err != nil {
		return fmt.Errorf("failed to soft delete reminder: %w", err)
	}
	return nil
}

// GetUpcomingReminders gets upcoming reminders for a user within a date range
func (r *Repository) GetUpcomingReminders(ctx context.Context, userID int32, start, end time.Time) ([]*domain.Reminder, error) {
	results, err := r.querier.GetUpcomingReminders(ctx, sqlcgen.GetUpcomingRemindersParams{
		UserID:         userID,
		ReminderDate:   pgtype.Timestamptz{Time: start, Valid: true},
		ReminderDate_2: pgtype.Timestamptz{Time: end, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	reminders := make([]*domain.Reminder, len(results))
	for i, result := range results {
		var readAt *time.Time
		if result.ReadAt.Valid {
			readAt = &result.ReadAt.Time
		}

		reminders[i] = &domain.Reminder{
			ID:                     result.ID,
			CreatedAt:              result.CreatedAt.Time,
			UpdatedAt:              result.UpdatedAt.Time,
			RecurringTransactionID: result.RecurringTransactionID,
			ReminderDate:           result.ReminderDate.Time,
			IsRead:                 result.IsRead,
			ReadAt:                 readAt,
		}
	}

	return reminders, nil
}

// Helper functions
func mapToReminder(r sqlcgen.Reminder) *domain.Reminder {
	var readAt *time.Time
	if r.ReadAt.Valid {
		readAt = &r.ReadAt.Time
	}

	return &domain.Reminder{
		ID:                     r.ID,
		CreatedAt:              r.CreatedAt.Time,
		UpdatedAt:              r.UpdatedAt.Time,
		RecurringTransactionID: r.RecurringTransactionID,
		ReminderDate:           r.ReminderDate.Time,
		IsRead:                 r.IsRead,
		ReadAt:                 readAt,
	}
}
