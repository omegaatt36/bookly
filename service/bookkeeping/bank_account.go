package bookkeeping

import (
	"errors"

	"github.com/omegaatt36/bookly/domain"
)

// CreateBankAccount creates a new bank account for an account
func (s *Service) CreateBankAccount(req domain.CreateBankAccountRequest) error {
	// Validate if the account exists
	account, err := s.accountRepo.GetAccountByID(req.AccountID)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("account does not exist")
	}

	// Check if the account already has a bank account
	existingBankAccount, err := s.bankAccountRepo.GetBankAccountByAccountID(req.AccountID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	if existingBankAccount != nil {
		return errors.New("account already has a bank account")
	}

	// Create the bank account
	return s.bankAccountRepo.CreateBankAccount(req)
}

// GetBankAccountByID retrieves a bank account by its ID
func (s *Service) GetBankAccountByID(id int32) (*domain.BankAccount, error) {
	return s.bankAccountRepo.GetBankAccountByID(id)
}

// GetBankAccountByAccountID retrieves a bank account by its associated account ID
func (s *Service) GetBankAccountByAccountID(accountID int32) (*domain.BankAccount, error) {
	return s.bankAccountRepo.GetBankAccountByAccountID(accountID)
}

// UpdateBankAccount updates an existing bank account
func (s *Service) UpdateBankAccount(req domain.UpdateBankAccountRequest) error {
	// Validate if the bank account exists
	bankAccount, err := s.bankAccountRepo.GetBankAccountByID(req.ID)
	if err != nil {
		return err
	}
	if bankAccount == nil {
		return errors.New("bank account does not exist")
	}

	// Update the bank account
	return s.bankAccountRepo.UpdateBankAccount(req)
}

// DeleteBankAccount deletes a bank account by its ID
func (s *Service) DeleteBankAccount(id int32) error {
	// Validate if the bank account exists
	bankAccount, err := s.bankAccountRepo.GetBankAccountByID(id)
	if err != nil {
		return err
	}
	if bankAccount == nil {
		return errors.New("bank account does not exist")
	}

	// Delete the bank account
	return s.bankAccountRepo.DeleteBankAccount(id)
}