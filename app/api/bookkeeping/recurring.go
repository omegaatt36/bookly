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
)

// RecurringTransactionResponse is the response for a recurring transaction
type RecurringTransactionResponse struct {
	ID           int32           `json:"id"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Amount       decimal.Decimal `json:"amount"`
	Note         string          `json:"note"`
	StartDate    time.Time       `json:"start_date"`
	EndDate      *time.Time      `json:"end_date,omitempty"`
	RecurType    string          `json:"recur_type"`
	Status       string          `json:"status"`
	Frequency    int             `json:"frequency"`
	DayOfWeek    *int            `json:"day_of_week,omitempty"`
	DayOfMonth   *int            `json:"day_of_month,omitempty"`
	MonthOfYear  *int            `json:"month_of_year,omitempty"`
	LastExecuted *time.Time      `json:"last_executed,omitempty"`
	NextDue      time.Time       `json:"next_due"`
}

// ReminderResponse is the response for a reminder
type ReminderResponse struct {
	ID                     int32      `json:"id"`
	CreatedAt              time.Time  `json:"created_at"`
	RecurringTransactionID int32      `json:"recurring_transaction_id"`
	ReminderDate           time.Time  `json:"reminder_date"`
	IsRead                 bool       `json:"is_read"`
	ReadAt                 *time.Time `json:"read_at,omitempty"`
}

// CreateRecurringTransaction creates a new recurring transaction
func (x *Controller) CreateRecurringTransaction() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			AccountID   int32           `json:"account_id"`
			Name        string          `json:"name"`
			Type        string          `json:"type"`
			Amount      decimal.Decimal `json:"amount"`
			Note        string          `json:"note"`
			StartDate   time.Time       `json:"start_date"`
			EndDate     *time.Time      `json:"end_date,omitempty"`
			RecurType   string          `json:"recur_type"`
			Frequency   int             `json:"frequency"`
			DayOfWeek   *int            `json:"day_of_week,omitempty"`
			DayOfMonth  *int            `json:"day_of_month,omitempty"`
			MonthOfYear *int            `json:"month_of_year,omitempty"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*RecurringTransactionResponse, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			if req.Name == "" || req.AccountID == 0 {
				return nil, app.ParamError(errors.New("name and account_id are required"))
			}

			if req.Amount.LessThanOrEqual(decimal.Zero) {
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

			serviceReq := domain.CreateRecurringTransactionRequest{
				UserID:      userID,
				AccountID:   req.AccountID,
				Name:        req.Name,
				Type:        ledgerType,
				Amount:      req.Amount,
				Note:        req.Note,
				StartDate:   req.StartDate,
				EndDate:     req.EndDate,
				RecurType:   recurType,
				Frequency:   req.Frequency,
				DayOfWeek:   req.DayOfWeek,
				DayOfMonth:  req.DayOfMonth,
				MonthOfYear: req.MonthOfYear,
			}

			transaction, err := x.service.CreateRecurringTransaction(r.Context(), serviceReq)
			if err != nil {
				slog.Error("Failed to create recurring transaction", "error", err)
				return nil, err
			}

			response := mapToRecurringTransactionResponse(transaction)
			return &response, nil
		}).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// GetRecurringTransactions gets all recurring transactions for the current user
func (x *Controller) GetRecurringTransactions() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]RecurringTransactionResponse, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			transactions, err := x.service.GetRecurringTransactionsByUserID(r.Context(), userID)
			if err != nil {
				slog.Error("Failed to get recurring transactions", "error", err)
				return nil, err
			}

			response := make([]RecurringTransactionResponse, len(transactions))
			for i, transaction := range transactions {
				response[i] = mapToRecurringTransactionResponse(transaction)
			}

			return response, nil
		}).Call(&engine.Empty{}).ResponseJSON()
	}
}

// GetRecurringTransaction gets a recurring transaction by ID
func (x *Controller) GetRecurringTransaction() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32

		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*RecurringTransactionResponse, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			transaction, err := x.service.GetRecurringTransaction(r.Context(), id)
			if err != nil {
				slog.Error("Failed to get recurring transaction", "id", id, "error", err)
				return nil, err
			}

			if transaction.UserID != userID {
				return nil, app.NotFoundError()
			}

			response := mapToRecurringTransactionResponse(transaction)
			return &response, nil
		}).Param("id", &id).Call(nil).ResponseJSON()
	}
}

// UpdateRecurringTransaction updates a recurring transaction
func (x *Controller) UpdateRecurringTransaction() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id          int32
			Name        *string          `json:"name,omitempty"`
			Type        *string          `json:"type,omitempty"`
			Amount      *decimal.Decimal `json:"amount,omitempty"`
			Note        *string          `json:"note,omitempty"`
			EndDate     *time.Time       `json:"end_date,omitempty"`
			RecurType   *string          `json:"recur_type,omitempty"`
			Status      *string          `json:"status,omitempty"`
			Frequency   *int             `json:"frequency,omitempty"`
			DayOfWeek   *int             `json:"day_of_week,omitempty"`
			DayOfMonth  *int             `json:"day_of_month,omitempty"`
			MonthOfYear *int             `json:"month_of_year,omitempty"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*RecurringTransactionResponse, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			existingTransaction, err := x.service.GetRecurringTransaction(r.Context(), req.id)
			if err != nil {
				slog.Error("Failed to get recurring transaction", "id", req.id, "error", err)
				return nil, err
			}

			if existingTransaction.UserID != userID {
				return nil, app.NotFoundError()
			}

			var transactionType *domain.LedgerType
			if req.Type != nil {
				t, err := domain.ParseLedgerType(*req.Type)
				if err != nil {
					return nil, app.ParamError(err)
				}
				transactionType = &t
			}

			var recurType *domain.RecurrenceType
			if req.RecurType != nil {
				rt, err := domain.ParseRecurrenceType(*req.RecurType)
				if err != nil {
					return nil, app.ParamError(err)
				}
				recurType = &rt
			}

			var status *domain.RecurrenceStatus
			if req.Status != nil {
				s, err := domain.ParseRecurrenceStatus(*req.Status)
				if err != nil {
					return nil, app.ParamError(err)
				}
				status = &s
			}

			serviceReq := domain.UpdateRecurringTransactionRequest{
				ID:          req.id,
				Name:        req.Name,
				Type:        transactionType,
				Amount:      req.Amount,
				Note:        req.Note,
				EndDate:     req.EndDate,
				RecurType:   recurType,
				Status:      status,
				Frequency:   req.Frequency,
				DayOfWeek:   req.DayOfWeek,
				DayOfMonth:  req.DayOfMonth,
				MonthOfYear: req.MonthOfYear,
			}

			transaction, err := x.service.UpdateRecurringTransaction(r.Context(), serviceReq)
			if err != nil {
				slog.Error("Failed to update recurring transaction", "id", req.id, "error", err)
				return nil, err
			}

			response := mapToRecurringTransactionResponse(transaction)
			return &response, nil
		}).Param("id", &req.id).BindJSON(&req).Call(req).ResponseJSON()
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
				slog.Error("Failed to get recurring transaction", "id", id, "error", err)
				return nil, err
			}

			if existingTransaction.UserID != userID {
				return nil, app.NotFoundError()
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
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]ReminderResponse, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			reminders, err := x.service.GetActiveRemindersByUserID(r.Context(), userID)
			if err != nil {
				slog.Error("Failed to get reminders", "error", err)
				return nil, err
			}

			response := make([]ReminderResponse, len(reminders))
			for i, reminder := range reminders {
				response[i] = mapToReminderResponse(reminder)
			}

			return response, nil
		}).Call(&engine.Empty{}).ResponseJSON()
	}
}

// MarkReminderAsRead marks a reminder as read
func (x *Controller) MarkReminderAsRead() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*ReminderResponse, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			reminder, err := x.service.GetReminderByID(r.Context(), id)
			if err != nil {
				slog.Error("Failed to get reminder", "id", id, "error", err)
				return nil, err
			}

			transaction, err := x.service.GetRecurringTransaction(r.Context(), reminder.RecurringTransactionID)
			if err != nil {
				slog.Error("Failed to get transaction for reminder", "id", reminder.RecurringTransactionID, "error", err)
				return nil, err
			}

			if transaction.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: reminder does not belong to user"))
			}

			// Mark reminder as read
			reminder, err = x.service.MarkReminderAsRead(r.Context(), id)
			if err != nil {
				slog.Error("Failed to mark reminder as read", "id", id, "error", err)
				return nil, err
			}

			response := mapToReminderResponse(reminder)
			return &response, nil
		}).Param("id", &id).Call(&engine.Empty{}).ResponseJSON()
	}
}

// DTO
func mapToRecurringTransactionResponse(t *domain.RecurringTransaction) RecurringTransactionResponse {
	return RecurringTransactionResponse{
		ID:           t.ID,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
		Name:         t.Name,
		Type:         string(t.Type),
		Amount:       t.Amount,
		Note:         t.Note,
		StartDate:    t.StartDate,
		EndDate:      t.EndDate,
		RecurType:    string(t.RecurType),
		Status:       string(t.Status),
		Frequency:    t.Frequency,
		DayOfWeek:    t.DayOfWeek,
		DayOfMonth:   t.DayOfMonth,
		MonthOfYear:  t.MonthOfYear,
		LastExecuted: t.LastExecuted,
		NextDue:      t.NextDue,
	}
}

// DTO
func mapToReminderResponse(r *domain.Reminder) ReminderResponse {
	return ReminderResponse{
		ID:                     r.ID,
		CreatedAt:              r.CreatedAt,
		RecurringTransactionID: r.RecurringTransactionID,
		ReminderDate:           r.ReminderDate,
		IsRead:                 r.IsRead,
		ReadAt:                 r.ReadAt,
	}
}
