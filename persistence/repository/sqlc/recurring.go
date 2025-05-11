package sqlc

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/sqlcgen"
)

// CreateRecurringTransaction creates a new recurring transaction
func (r *Repository) CreateRecurringTransaction(ctx context.Context, req domain.CreateRecurringTransactionRequest) (*domain.RecurringTransaction, error) {
	var dayOfWeek, dayOfMonth, monthOfYear pgtype.Int4

	if req.DayOfWeek != nil {
		dayOfWeek.Int32 = int32(*req.DayOfWeek)
		dayOfWeek.Valid = true
	}

	if req.DayOfMonth != nil {
		dayOfMonth.Int32 = int32(*req.DayOfMonth)
		dayOfMonth.Valid = true
	}

	if req.MonthOfYear != nil {
		monthOfYear.Int32 = int32(*req.MonthOfYear)
		monthOfYear.Valid = true
	}

	var endDate pgtype.Timestamptz
	if req.EndDate != nil {
		endDate.Time = *req.EndDate
		endDate.Valid = true
	}

	nextDue := calculateNextDueDate(req.StartDate, req.RecurType, req.Frequency, req.DayOfWeek, req.DayOfMonth, req.MonthOfYear)

	params := sqlcgen.CreateRecurringTransactionParams{
		UserID:      req.UserID,
		AccountID:   req.AccountID,
		Name:        req.Name,
		Type:        string(req.Type),
		Amount:      req.Amount,
		Note:        pgtype.Text{String: req.Note, Valid: true},
		StartDate:   pgtype.Timestamptz{Time: req.StartDate, Valid: true},
		EndDate:     endDate,
		RecurType:   string(req.RecurType),
		Status:      string(domain.RecurrenceStatusActive),
		Frequency:   int32(req.Frequency),
		DayOfWeek:   dayOfWeek,
		DayOfMonth:  dayOfMonth,
		MonthOfYear: monthOfYear,
		NextDue:     pgtype.Timestamptz{Time: nextDue, Valid: true},
	}

	result, err := r.querier.CreateRecurringTransaction(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapToRecurringTransaction(result), nil
}

// GetRecurringTransactionByID gets a recurring transaction by ID
func (r *Repository) GetRecurringTransactionByID(ctx context.Context, id string) (*domain.RecurringTransaction, error) {
	result, err := r.querier.GetRecurringTransactionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapToRecurringTransaction(result), nil
}

// GetRecurringTransactionsByUserID gets all recurring transactions for a user
func (r *Repository) GetRecurringTransactionsByUserID(ctx context.Context, userID string) ([]*domain.RecurringTransaction, error) {
	results, err := r.querier.GetRecurringTransactionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	transactions := make([]*domain.RecurringTransaction, len(results))
	for i, result := range results {
		transactions[i] = mapToRecurringTransaction(result)
	}

	return transactions, nil
}

// GetActiveRecurringTransactionsDue gets all active recurring transactions due before a specified time
func (r *Repository) GetActiveRecurringTransactionsDue(ctx context.Context, before time.Time) ([]*domain.RecurringTransaction, error) {
	results, err := r.querier.GetActiveRecurringTransactionsDue(ctx, pgtype.Timestamptz{Time: before, Valid: true})
	if err != nil {
		return nil, err
	}

	transactions := make([]*domain.RecurringTransaction, len(results))
	for i, result := range results {
		transactions[i] = mapToRecurringTransaction(result)
	}

	return transactions, nil
}

// UpdateRecurringTransaction updates a recurring transaction
func (r *Repository) UpdateRecurringTransaction(ctx context.Context, req domain.UpdateRecurringTransactionRequest) (*domain.RecurringTransaction, error) {
	params := sqlcgen.UpdateRecurringTransactionParams{
		ID: req.ID,
	}

	if req.Name != nil {
		params.Name = pgtype.Text{
			String: *req.Name,
			Valid:  true,
		}
	}

	if req.Type != nil {
		params.Type = pgtype.Text{
			String: string(*req.Type),
			Valid:  true,
		}
	}

	if req.Amount != nil {
		params.Amount = pgtype.Numeric{Valid: true}
		params.Amount.InfinityModifier = pgtype.Finite
		params.Amount.NaN = false
		params.Amount.Int = req.Amount.Coefficient()
		params.Amount.Exp = req.Amount.Exponent()
	}

	if req.Note != nil {
		params.Note = pgtype.Text{
			String: *req.Note,
			Valid:  true,
		}
	}

	if req.EndDate != nil {
		params.EndDate = pgtype.Timestamptz{
			Time:  *req.EndDate,
			Valid: true,
		}
	}

	if req.RecurType != nil {
		params.RecurType = pgtype.Text{
			String: string(*req.RecurType),
			Valid:  true,
		}
	}

	if req.Status != nil {
		params.Status = pgtype.Text{
			String: string(*req.Status),
			Valid:  true,
		}
	}

	if req.Frequency != nil {
		params.Frequency = pgtype.Int4{
			Int32: int32(*req.Frequency),
			Valid: true,
		}
	}

	if req.DayOfWeek != nil {
		params.DayOfWeek = pgtype.Int4{
			Int32: int32(*req.DayOfWeek),
			Valid: true,
		}
	}

	if req.DayOfMonth != nil {
		params.DayOfMonth = pgtype.Int4{
			Int32: int32(*req.DayOfMonth),
			Valid: true,
		}
	}

	if req.MonthOfYear != nil {
		params.MonthOfYear = pgtype.Int4{
			Int32: int32(*req.MonthOfYear),
			Valid: true,
		}
	}

	result, err := r.querier.UpdateRecurringTransaction(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return mapToRecurringTransaction(result), nil
}

// UpdateRecurringTransactionExecution updates the last executed and next due dates
func (r *Repository) UpdateRecurringTransactionExecution(ctx context.Context, id string, lastExecuted, nextDue time.Time) (*domain.RecurringTransaction, error) {
	params := sqlcgen.UpdateRecurringTransactionExecutionParams{
		ID:           id,
		LastExecuted: pgtype.Timestamptz{Time: lastExecuted, Valid: true},
		NextDue:      pgtype.Timestamptz{Time: nextDue, Valid: true},
	}

	result, err := r.querier.UpdateRecurringTransactionExecution(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return mapToRecurringTransaction(result), nil
}

// DeleteRecurringTransaction implements the domain.RecurringTransactionRepository interface
// This method performs a soft delete by setting the deleted_at timestamp and status to cancelled.
func (r *Repository) DeleteRecurringTransaction(ctx context.Context, id string) error {
	// The SQL query now sets deleted_at and status = 'cancelled'.
	err := r.querier.DeleteRecurringTransaction(ctx, id)
	if err != nil {
		// It's good practice to wrap errors for context
		return fmt.Errorf("failed to soft delete recurring transaction: %w", err)
	}
	return nil
}

// Helper functions
func mapToRecurringTransaction(rt sqlcgen.RecurringTransaction) *domain.RecurringTransaction {
	var endDate *time.Time
	if rt.EndDate.Valid {
		endDate = &rt.EndDate.Time
	}

	var lastExecuted *time.Time
	if rt.LastExecuted.Valid {
		lastExecuted = &rt.LastExecuted.Time
	}

	var dayOfWeek, dayOfMonth, monthOfYear *int
	if rt.DayOfWeek.Valid {
		dow := int(rt.DayOfWeek.Int32)
		dayOfWeek = &dow
	}

	if rt.DayOfMonth.Valid {
		dom := int(rt.DayOfMonth.Int32)
		dayOfMonth = &dom
	}

	if rt.MonthOfYear.Valid {
		moy := int(rt.MonthOfYear.Int32)
		monthOfYear = &moy
	}

	return &domain.RecurringTransaction{
		ID:           rt.ID,
		CreatedAt:    rt.CreatedAt.Time,
		UpdatedAt:    rt.UpdatedAt.Time,
		UserID:       rt.UserID,
		AccountID:    rt.AccountID,
		Name:         rt.Name,
		Type:         domain.LedgerType(rt.Type),
		Amount:       rt.Amount,
		Note:         rt.Note.String,
		StartDate:    rt.StartDate.Time,
		EndDate:      endDate,
		RecurType:    domain.RecurrenceType(rt.RecurType),
		Status:       domain.RecurrenceStatus(rt.Status),
		Frequency:    int(rt.Frequency),
		DayOfWeek:    dayOfWeek,
		DayOfMonth:   dayOfMonth,
		MonthOfYear:  monthOfYear,
		LastExecuted: lastExecuted,
		NextDue:      rt.NextDue.Time,
	}
}

func calculateNextDueDate(startDate time.Time, recurType domain.RecurrenceType, frequency int, dayOfWeek, dayOfMonth, monthOfYear *int) time.Time {
	now := time.Now()

	if startDate.After(now) {
		return startDate
	}

	switch recurType {
	case domain.RecurrenceTypeDaily:
		return now.AddDate(0, 0, frequency)

	case domain.RecurrenceTypeWeekly:
		daysToAdd := frequency * 7
		if dayOfWeek != nil {
			currentDOW := int(now.Weekday())
			targetDOW := *dayOfWeek

			daysToAdd = (targetDOW - currentDOW + 7) % 7
			if daysToAdd == 0 && frequency > 0 {
				daysToAdd = 7 * frequency
			}
		}
		return now.AddDate(0, 0, daysToAdd)

	case domain.RecurrenceTypeBiweekly:
		return now.AddDate(0, 0, 14*frequency)

	case domain.RecurrenceTypeMonthly:
		nextDate := now.AddDate(0, frequency, 0)

		if dayOfMonth != nil {
			targetDay := *dayOfMonth
			if targetDay > 28 {
				daysInMonth := daysInMonth(nextDate.Year(), int(nextDate.Month()))
				if targetDay > daysInMonth {
					targetDay = daysInMonth
				}
			}

			nextDate = time.Date(nextDate.Year(), nextDate.Month(), targetDay,
				nextDate.Hour(), nextDate.Minute(), nextDate.Second(),
				nextDate.Nanosecond(), nextDate.Location())
		}

		return nextDate

	case domain.RecurrenceTypeQuarterly:
		nextDate := now.AddDate(0, 3*frequency, 0)

		if dayOfMonth != nil {
			targetDay := *dayOfMonth
			if targetDay > 28 {
				daysInMonth := daysInMonth(nextDate.Year(), int(nextDate.Month()))
				if targetDay > daysInMonth {
					targetDay = daysInMonth
				}
			}

			nextDate = time.Date(nextDate.Year(), nextDate.Month(), targetDay,
				nextDate.Hour(), nextDate.Minute(), nextDate.Second(),
				nextDate.Nanosecond(), nextDate.Location())
		}

		return nextDate

	case domain.RecurrenceTypeYearly:
		nextDate := now.AddDate(frequency, 0, 0)

		if monthOfYear != nil && dayOfMonth != nil {
			targetMonth := time.Month(*monthOfYear)
			targetDay := *dayOfMonth

			if targetDay > 28 {
				daysInMonth := daysInMonth(nextDate.Year(), *monthOfYear)
				if targetDay > daysInMonth {
					targetDay = daysInMonth
				}
			}

			nextDate = time.Date(nextDate.Year(), targetMonth, targetDay,
				nextDate.Hour(), nextDate.Minute(), nextDate.Second(),
				nextDate.Nanosecond(), nextDate.Location())
		}

		return nextDate

	case domain.RecurrenceTypeCustom:
		return now.AddDate(0, 0, frequency)

	default:
		return now.AddDate(0, 1, 0)
	}
}

func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
}
