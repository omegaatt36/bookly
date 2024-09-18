package bookkeeping

import "github.com/omegaatt36/bookly/domain"

// CreateAccount creates a new account based on the provided CreateAccountRequest.
func (s *Service) CreateAccount(req domain.CreateAccountRequest) error {
	return s.accountRepo.CreateAccount(req)
}

// GetAccountByID retrieves an account by its ID.
func (s *Service) GetAccountByID(id string) (*domain.Account, error) {
	return s.accountRepo.GetAccountByID(id)
}

// UpdateAccount updates an existing account based on the provided UpdateAccountRequest.
func (s *Service) UpdateAccount(req domain.UpdateAccountRequest) error {
	return s.accountRepo.UpdateAccount(req)
}

// DeactivateAccountByID deactivates an account by its ID.
func (s *Service) DeactivateAccountByID(id string) error {
	return s.accountRepo.DeactivateAccountByID(id)
}

// GetAllAccounts retrieves all accounts.
func (s *Service) GetAllAccounts() ([]*domain.Account, error) {
	return s.accountRepo.GetAllAccounts()
}

// GetAccountsByUserID retrieves all accounts by userID.
func (s *Service) GetAccountsByUserID(userID string) ([]*domain.Account, error) {
	return s.accountRepo.GetAccountsByUserID(userID)
}
