package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/omegaatt36/bookly/domain"
)

var _ domain.LedgerRepository = (*GORMRepository)(nil)

// Ledger represents the database model for a ledger entry
type Ledger struct {
	ID           string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	AccountID    string          `gorm:"type:uuid;not null"`
	Date         time.Time       `gorm:"not null"`
	Type         string          `gorm:"type:varchar(20);not null"`
	Amount       decimal.Decimal `gorm:"type:decimal(20,2);not null"`
	Note         string          `gorm:"type:text"`
	IsAdjustment bool            `gorm:"not null"`
	AdjustedFrom *string         `gorm:"type:uuid"`
	IsVoided     bool            `gorm:"not null"`
	VoidedAt     *time.Time
}

// toDomainLedger converts repository Ledger to domain.Ledger
func (l *Ledger) toDomainLedger() *domain.Ledger {
	return &domain.Ledger{
		ID:           l.ID,
		CreatedAt:    l.CreatedAt,
		UpdatedAt:    l.UpdatedAt,
		AccountID:    l.AccountID,
		Date:         l.Date,
		Type:         domain.LedgerType(l.Type),
		Amount:       l.Amount,
		Note:         l.Note,
		IsAdjustment: l.IsAdjustment,
		AdjustedFrom: l.AdjustedFrom,
		IsVoided:     l.IsVoided,
		VoidedAt:     l.VoidedAt,
	}
}

// CreateLedger creates a new ledger entry
func (r *GORMRepository) CreateLedger(req domain.CreateLedgerRequest) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var account Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&account, "id = ?", req.AccountID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("account not found: %s", req.AccountID)
			}
			return fmt.Errorf("failed to check account existence: %w", err)
		}

		ledger := Ledger{
			AccountID: req.AccountID,
			Date:      req.Date,
			Type:      string(req.Type),
			Amount:    req.Amount,
			Note:      req.Note,
		}
		if err := tx.Create(&ledger).Error; err != nil {
			return fmt.Errorf("failed to create ledger: %w", err)
		}

		account.Balance = account.Balance.Add(req.Amount)
		if err := tx.Save(&account).Error; err != nil {
			return fmt.Errorf("failed to update account balance: %w", err)
		}

		return nil
	})
}

// GetLedgerByID retrieves a ledger entry by its ID
func (r *GORMRepository) GetLedgerByID(id string) (*domain.Ledger, error) {
	var ledger Ledger
	if err := r.db.First(&ledger, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ledger not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get ledger: %w", err)
	}

	return ledger.toDomainLedger(), nil
}

// GetLedgersByAccountID retrieves all ledger entries for a given account ID
func (r *GORMRepository) GetLedgersByAccountID(accountID string) ([]*domain.Ledger, error) {
	var ledgers []Ledger
	if err := r.db.Where("account_id = ?", accountID).Order("date desc").Find(&ledgers).Error; err != nil {
		return nil, fmt.Errorf("failed to get ledgers for account: %w", err)
	}

	domainLedgers := make([]*domain.Ledger, len(ledgers))
	for i, ledger := range ledgers {
		domainLedgers[i] = ledger.toDomainLedger()
	}

	return domainLedgers, nil
}

// UpdateLedger updates an existing ledger entry
func (r *GORMRepository) UpdateLedger(req domain.UpdateLedgerRequest) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var ledger Ledger
		if err := tx.First(&ledger, "id = ?", req.ID).Error; err != nil {
			return fmt.Errorf("failed to find ledger: %w", err)
		}

		oldAmount := ledger.Amount

		if req.Date != nil {
			ledger.Date = *req.Date
		}
		if req.Type != nil {
			ledger.Type = string(*req.Type)
		}
		if req.Amount != nil {
			ledger.Amount = *req.Amount
		}
		if req.Note != nil {
			ledger.Note = *req.Note
		}
		if err := tx.Save(&ledger).Error; err != nil {
			return fmt.Errorf("failed to update ledger: %w", err)
		}

		if req.Amount != nil {
			var account Account
			if err := tx.First(&account, "id = ?", ledger.AccountID).Error; err != nil {
				return fmt.Errorf("failed to find account: %w", err)
			}

			account.Balance = account.Balance.Sub(oldAmount).Add(ledger.Amount)
			if err := tx.Save(&account).Error; err != nil {
				return fmt.Errorf("failed to update account balance: %w", err)
			}
		}

		return nil
	})
}

// VoidLedger marks a ledger entry as voided
func (r *GORMRepository) VoidLedger(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var ledger Ledger
		if err := tx.First(&ledger, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("ledger not found: %s", id)
			}
			return fmt.Errorf("failed to find ledger: %w", err)
		}

		now := time.Now()
		ledger.IsVoided = true
		ledger.VoidedAt = &now

		if err := tx.Save(&ledger).Error; err != nil {
			return fmt.Errorf("failed to void ledger: %w", err)
		}

		var account Account
		if err := tx.First(&account, "id = ?", ledger.AccountID).Error; err != nil {
			return fmt.Errorf("failed to find account: %w", err)
		}

		account.Balance = account.Balance.Sub(ledger.Amount)
		if err := tx.Save(&account).Error; err != nil {
			return fmt.Errorf("failed to update account balance: %w", err)
		}

		return nil
	})
}

// AdjustLedger creates a new adjusted ledger entry based on an existing one
func (r *GORMRepository) AdjustLedger(originalID string, adjustment domain.CreateLedgerRequest) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var originalLedger Ledger
		if err := tx.First(&originalLedger, "id = ?", originalID).Error; err != nil {
			return fmt.Errorf("failed to find original ledger: %w", err)
		}

		adjustedLedger := Ledger{
			AccountID:    adjustment.AccountID,
			Date:         adjustment.Date,
			Type:         string(adjustment.Type),
			Amount:       adjustment.Amount,
			Note:         adjustment.Note,
			IsAdjustment: true,
			AdjustedFrom: &originalID,
		}
		if err := tx.Create(&adjustedLedger).Error; err != nil {
			return fmt.Errorf("failed to create adjusted ledger: %w", err)
		}

		var account Account
		if err := tx.First(&account, "id = ?", adjustment.AccountID).Error; err != nil {
			return fmt.Errorf("failed to find account: %w", err)
		}

		account.Balance = account.Balance.Add(adjustment.Amount)
		if err := tx.Save(&account).Error; err != nil {
			return fmt.Errorf("failed to update account balance: %w", err)
		}

		return nil
	})
}
