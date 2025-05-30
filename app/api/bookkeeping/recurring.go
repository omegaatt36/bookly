package bookkeeping

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/sdk/datatype"
)

// CreateRecurringTransaction creates a new recurring transaction
func (x *Controller) CreateRecurringTransaction() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req datatype.CreateRecurringTransactionRequest
		engine.Chain(r, w, func(ctx *engine.Context, req datatype.CreateRecurringTransactionRequest) (*datatype.RecurringTransaction, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Verify account ownership
			account, err := x.service.GetAccountByID(req.AccountID)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return nil, app.NotFoundError()
				}
				slog.Error("Failed to get account", "account_id", req.AccountID, "error", err)
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: account does not belong to user"))
			}

			if req.Name == "" {
				return nil, app.ParamError(errors.New("name is required"))
			}

			amount, err := decimal.NewFromString(req.Amount)
			if err != nil {
				return nil, app.ParamError(errors.New("invalid amount format"))
			}
			if amount.LessThanOrEqual(decimal.Zero) {
				return nil, app.ParamError(errors.New("amount must be greater than zero"))
			}

			ledgerType, err := domain.ParseLedgerType(req.Type)
			if err != nil {
				return nil, app.ParamError(err)
			}

			recurType, err := domain.ParseRecurrenceType(req.RecurType)
			if err != nil {
				return nil, app.ParamError(err)
			}

			startDate, err := time.Parse(datatype.TimeFormat, req.StartDate)
			if err != nil {
				return nil, app.ParamError(errors.New("invalid start_date format"))
			}

			var endDate *time.Time
			if req.EndDate != nil && *req.EndDate != "" {
				t, err := time.Parse(datatype.TimeFormat, *req.EndDate)
				if err != nil {
					return nil, app.ParamError(errors.New("invalid end_date format"))
				}
				endDate = &t
			}

			serviceReq := domain.CreateRecurringTransactionRequest{
				UserID:      userID,
				AccountID:   req.AccountID,
				Name:        req.Name,
				Type:        ledgerType,
				Amount:      amount,
				Note:        req.Note,
				StartDate:   startDate,
				EndDate:     endDate,
				RecurType:   recurType,
				Frequency:   req.Frequency,
				DayOfWeek:   req.DayOfWeek,
				DayOfMonth:  req.DayOfMonth,
				MonthOfYear: req.MonthOfYear,
			}

			transaction, err := x.service.CreateRecurringTransaction(r.Context(), serviceReq)
			if err != nil {
				slog.Error("Failed to create recurring transaction", "error", err, "request", serviceReq)
				return nil, err
			}

			response := convertRecurringTransactionFromDomain(transaction)
			return &response, nil
		}).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// GetRecurringTransactions gets all recurring transactions for the current user
func (x *Controller) GetRecurringTransactions() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]datatype.RecurringTransaction, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			transactions, err := x.service.GetRecurringTransactionsByUserID(r.Context(), userID)
			if err != nil {
				slog.Error("Failed to get recurring transactions", "user_id", userID, "error", err)
				return nil, err
			}

			response := make([]datatype.RecurringTransaction, len(transactions))
			for i, transaction := range transactions {
				response[i] = convertRecurringTransactionFromDomain(transaction)
			}

			return response, nil
		}).Call(&engine.Empty{}).ResponseJSON()
	}
}

// GetRecurringTransaction gets a recurring transaction by ID
func (x *Controller) GetRecurringTransaction() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32

		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*datatype.RecurringTransaction, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			transaction, err := x.service.GetRecurringTransaction(r.Context(), id)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return nil, app.NotFoundError()
				}
				slog.Error("Failed to get recurring transaction", "id", id, "error", err)
				return nil, err
			}

			if transaction.UserID != userID {
				// This case should ideally not happen if service layer correctly filters by user ID or if ID is globally unique and checked
				// However, as a safeguard:
				return nil, app.Forbidden(errors.New("access denied: recurring transaction does not belong to user"))
			}

			response := convertRecurringTransactionFromDomain(transaction)
			return &response, nil
		}).Param("id", &id).Call(nil).ResponseJSON()
	}
}

// UpdateRecurringTransaction updates a recurring transaction
func (x *Controller) UpdateRecurringTransaction() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type requestPayload struct {
			datatype.UpdateRecurringTransactionRequest
		}
		var payload requestPayload
		var recurringTransactionID int32

		engine.Chain(r, w, func(ctx *engine.Context, p requestPayload) (*datatype.RecurringTransaction, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			existingTransaction, err := x.service.GetRecurringTransaction(r.Context(), recurringTransactionID)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return nil, app.NotFoundError()
				}
				slog.Error("Failed to get recurring transaction for update", "id", recurringTransactionID, "error", err)
				return nil, err
			}

			if existingTransaction.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: recurring transaction does not belong to user"))
			}

			serviceReq := domain.UpdateRecurringTransactionRequest{
				ID:          recurringTransactionID,
				Name:        p.Name,
				Note:        p.Note,
				Frequency:   p.Frequency,
				DayOfWeek:   p.DayOfWeek,
				DayOfMonth:  p.DayOfMonth,
				MonthOfYear: p.MonthOfYear,
			}

			if p.Type != nil && *p.Type != "" {
				t, err := domain.ParseLedgerType(*p.Type)
				if err != nil {
					return nil, app.ParamError(err)
				}
				serviceReq.Type = &t
			}

			if p.Amount != nil && *p.Amount != "" {
				amount, err := decimal.NewFromString(*p.Amount)
				if err != nil {
					return nil, app.ParamError(errors.New("invalid amount format"))
				}
				if amount.LessThanOrEqual(decimal.Zero) {
					return nil, app.ParamError(errors.New("amount must be greater than zero"))
				}
				serviceReq.Amount = &amount
			}

			if p.EndDate != nil && *p.EndDate != "" {
				t, err := time.Parse(datatype.TimeFormat, *p.EndDate)
				if err != nil {
					return nil, app.ParamError(errors.New("invalid end_date format"))
				}
				serviceReq.EndDate = &t
			}

			if p.RecurType != nil && *p.RecurType != "" {
				rt, err := domain.ParseRecurrenceType(*p.RecurType)
				if err != nil {
					return nil, app.ParamError(err)
				}
				serviceReq.RecurType = &rt
			}

			if p.Status != nil && *p.Status != "" {
				s, err := domain.ParseRecurrenceStatus(*p.Status)
				if err != nil {
					return nil, app.ParamError(err)
				}
				serviceReq.Status = &s
			}

			transaction, err := x.service.UpdateRecurringTransaction(r.Context(), serviceReq)
			if err != nil {
				slog.Error("Failed to update recurring transaction", "id", recurringTransactionID, "error", err, "request", serviceReq)
				return nil, err
			}

			response := convertRecurringTransactionFromDomain(transaction)
			return &response, nil
		}).Param("id", &recurringTransactionID).BindJSON(&payload).Call(payload).ResponseJSON()
	}
}

// DeleteRecurringTransaction deletes a recurring transaction
func (x *Controller) DeleteRecurringTransaction() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*engine.Empty, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			existingTransaction, err := x.service.GetRecurringTransaction(r.Context(), id)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return nil, app.NotFoundError()
				}
				slog.Error("Failed to get recurring transaction for delete", "id", id, "error", err)
				return nil, err
			}

			if existingTransaction.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: recurring transaction does not belong to user"))
			}

			if err := x.service.DeleteRecurringTransaction(r.Context(), id); err != nil {
				slog.Error("Failed to delete recurring transaction", "id", id, "error", err)
				return nil, err
			}

			return nil, nil
		}).Param("id", &id).Call(&engine.Empty{}).ResponseJSON()
	}
}

// GetReminders gets all active reminders for the current user
func (x *Controller) GetReminders() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]datatype.Reminder, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			reminders, err := x.service.GetActiveRemindersByUserID(r.Context(), userID)
			if err != nil {
				slog.Error("Failed to get reminders", "user_id", userID, "error", err)
				return nil, err
			}

			response := make([]datatype.Reminder, len(reminders))
			for i, reminder := range reminders {
				response[i] = convertReminderFromDomain(reminder)
			}

			return response, nil
		}).Call(&engine.Empty{}).ResponseJSON()
	}
}

// MarkReminderAsRead marks a reminder as read
func (x *Controller) MarkReminderAsRead() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*datatype.Reminder, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			reminder, err := x.service.GetReminderByID(r.Context(), id)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					slog.Error("Reminder not found", "id", id)
					return nil, app.NotFoundError()
				}
				slog.Error("Failed to get reminder", "id", id, "error", err)
				return nil, err
			}

			// Verify ownership by checking the associated recurring transaction's user ID
			transaction, err := x.service.GetRecurringTransaction(r.Context(), reminder.RecurringTransactionID)
			if err != nil {
				// If the transaction is not found, it's an internal data integrity issue or the transaction was deleted.
				// Log it and return an error.
				slog.Error("Failed to get recurring transaction for reminder",
					"reminder_id", id,
					"recurring_transaction_id", reminder.RecurringTransactionID,
					"error", err)
				if errors.Is(err, domain.ErrNotFound) {
					slog.Error("Recurring transaction not found for reminder", "reminder_id", id)
					return nil, app.NotFoundError()
				}
				return nil, err // Or a more generic server error
			}

			if transaction.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: reminder does not belong to user"))
			}

			// Mark reminder as read
			updatedReminder, err := x.service.MarkReminderAsRead(r.Context(), id)
			if err != nil {
				slog.Error("Failed to mark reminder as read", "id", id, "error", err)
				return nil, err
			}

			response := convertReminderFromDomain(updatedReminder)
			return &response, nil
		}).Param("id", &id).Call(&engine.Empty{}).ResponseJSON()
	}
}

func convertRecurringTransactionFromDomain(t *domain.RecurringTransaction) datatype.RecurringTransaction {
	resp := datatype.RecurringTransaction{
		ID:             t.ID,
		CreatedAt:      t.CreatedAt.Format(datatype.TimeFormat),
		UpdatedAt:      t.UpdatedAt.Format(datatype.TimeFormat),
		UserID:         t.UserID,
		AccountID:      t.AccountID,
		Name:           t.Name,
		Type:           string(t.Type),
		Amount:         t.Amount.String(),
		Note:           t.Note,
		StartDate:      t.StartDate.Format(datatype.TimeFormat),
		RecurrenceType: string(t.RecurrenceType),
		Status:         string(t.Status),
		Frequency:      t.Frequency,
		DayOfWeek:      t.DayOfWeek,
		DayOfMonth:     t.DayOfMonth,
		MonthOfYear:    t.MonthOfYear,
		NextDue:        t.NextDue.Format(datatype.TimeFormat),
	}
	if t.EndDate != nil {
		endDateStr := t.EndDate.Format(datatype.TimeFormat)
		resp.EndDate = &endDateStr
	}
	if t.LastExecuted != nil {
		lastExecutedStr := t.LastExecuted.Format(datatype.TimeFormat)
		resp.LastExecuted = &lastExecutedStr
	}
	return resp
}

func convertReminderFromDomain(r *domain.Reminder) datatype.Reminder {
	resp := datatype.Reminder{
		ID:                     r.ID,
		CreatedAt:              r.CreatedAt.Format(datatype.TimeFormat),
		RecurringTransactionID: r.RecurringTransactionID,
		ReminderDate:           r.ReminderDate.Format(datatype.TimeFormat),
		IsRead:                 r.IsRead,
	}
	if r.ReadAt != nil {
		readAtStr := r.ReadAt.Format(datatype.TimeFormat)
		resp.ReadAt = &readAtStr
	}
	return resp
}
