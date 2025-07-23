package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"simple_api/internal/cache"
	"simple_api/internal/models"
	"simple_api/internal/repository"
	"simple_api/pkg/logger"
)

// Common errors
var (
	ErrWalletNotFound     = errors.New("wallet not found")
	ErrTokenNotFound      = errors.New("token not found")
	ErrInvalidAddress     = errors.New("invalid wallet address")
	ErrWalletAlreadyExists = errors.New("wallet already exists in watchlist")
	ErrTokenAlreadyExists  = errors.New("token already exists in watchlist")
)

// Request/Response types
type AddWalletRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Label         string `json:"label"`
}

type AddTokenRequest struct {
	TokenAddress *string `json:"token_address"` // nil for ETH
	TokenSymbol  string  `json:"token_symbol" binding:"required"`
	TokenName    string  `json:"token_name" binding:"required"`
}

type WalletResponse struct {
	ID            uint      `json:"id"`
	WalletAddress string    `json:"wallet_address"`
	Label         string    `json:"label"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TokenResponse struct {
	ID           uint      `json:"id"`
	TokenAddress *string   `json:"token_address"`
	TokenSymbol  string    `json:"token_symbol"`
	TokenName    string    `json:"token_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type BalanceResponse struct {
	WalletID     uint      `json:"wallet_id"`
	WalletAddress string   `json:"wallet_address"`
	TokenID      uint      `json:"token_id"`
	TokenSymbol  string    `json:"token_symbol"`
	Balance      string    `json:"balance"`
	BalanceUSD   *string   `json:"balance_usd,omitempty"`
	FetchedAt    time.Time `json:"fetched_at"`
}

type BalanceHistoryResponse struct {
	ID           uint      `json:"id"`
	WalletID     uint      `json:"wallet_id"`
	WalletAddress string   `json:"wallet_address"`
	TokenID      uint      `json:"token_id"`
	TokenSymbol  string    `json:"token_symbol"`
	Balance      string    `json:"balance"`
	BalanceUSD   *string   `json:"balance_usd,omitempty"`
	FetchedAt    time.Time `json:"fetched_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// WatchlistService interface defines the contract for watchlist operations
type WatchlistService interface {
	// Wallet operations
	AddWallet(ctx context.Context, userID uint, req *AddWalletRequest) (*WalletResponse, error)
	GetWallets(ctx context.Context, userID uint) ([]*WalletResponse, error)
	DeleteWallet(ctx context.Context, userID uint, walletID uint) error
	
	// Token operations
	AddToken(ctx context.Context, userID uint, req *AddTokenRequest) (*TokenResponse, error)
	GetTokens(ctx context.Context, userID uint) ([]*TokenResponse, error)
	DeleteToken(ctx context.Context, userID uint, tokenID uint) error
	
	// Balance operations
	GetBalances(ctx context.Context, userID uint) ([]*BalanceResponse, error)
	GetBalanceHistory(ctx context.Context, userID uint, walletID uint, tokenID uint, limit int) ([]*BalanceHistoryResponse, error)
	RefreshBalances(ctx context.Context, userID uint) error
}

// watchlistService implements WatchlistService
type watchlistService struct {
	watchlistRepo     repository.WatchlistRepository
	web3Service       Web3Service
	balanceFetcher    BalanceFetcherService
	cacheService      cache.CacheProvider
	logger            *logger.Logger
}

// NewWatchlistService creates a new watchlist service
func NewWatchlistService(
	watchlistRepo repository.WatchlistRepository,
	web3Service Web3Service,
	balanceFetcher BalanceFetcherService,
	cacheService cache.CacheProvider,
	logger *logger.Logger,
) WatchlistService {
	return &watchlistService{
		watchlistRepo:  watchlistRepo,
		web3Service:    web3Service,
		balanceFetcher: balanceFetcher,
		cacheService:   cacheService,
		logger:         logger,
	}
}

// AddWallet adds a wallet to user's watchlist
func (s *watchlistService) AddWallet(ctx context.Context, userID uint, req *AddWalletRequest) (*WalletResponse, error) {
	// Validate wallet address
	if !s.web3Service.ValidateAddress(req.WalletAddress) {
		return nil, ErrInvalidAddress
	}
	
	// Check if wallet already exists for this user
	wallets, err := s.watchlistRepo.GetWalletsByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user wallets", "error", err, "user_id", userID)
		return nil, err
	}
	
	for _, wallet := range wallets {
		if wallet.WalletAddress == req.WalletAddress {
			return nil, ErrWalletAlreadyExists
		}
	}
	
	// Create wallet
	wallet := &models.WatchlistWallet{
		UserID:        userID,
		WalletAddress: req.WalletAddress,
		Label:         req.Label,
	}
	
	if err := s.watchlistRepo.CreateWallet(ctx, wallet); err != nil {
		s.logger.Error("Failed to create wallet", "error", err, "user_id", userID, "address", req.WalletAddress)
		return nil, err
	}
	
	// Invalidate cache
	s.invalidateUserCache(ctx, userID)
	
	s.logger.Info("Wallet added to watchlist", "user_id", userID, "wallet_id", wallet.ID, "address", req.WalletAddress)
	
	return &WalletResponse{
		ID:            wallet.ID,
		WalletAddress: wallet.WalletAddress,
		Label:         wallet.Label,
		CreatedAt:     wallet.CreatedAt,
		UpdatedAt:     wallet.UpdatedAt,
	}, nil
}

// GetWallets retrieves user's watchlist wallets
func (s *watchlistService) GetWallets(ctx context.Context, userID uint) ([]*WalletResponse, error) {
	wallets, err := s.watchlistRepo.GetWalletsByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user wallets", "error", err, "user_id", userID)
		return nil, err
	}
	
	responses := make([]*WalletResponse, len(wallets))
	for i, wallet := range wallets {
		responses[i] = &WalletResponse{
			ID:            wallet.ID,
			WalletAddress: wallet.WalletAddress,
			Label:         wallet.Label,
			CreatedAt:     wallet.CreatedAt,
			UpdatedAt:     wallet.UpdatedAt,
		}
	}
	
	return responses, nil
}

// DeleteWallet removes a wallet from user's watchlist
func (s *watchlistService) DeleteWallet(ctx context.Context, userID uint, walletID uint) error {
	if err := s.watchlistRepo.DeleteWallet(ctx, walletID, userID); err != nil {
		s.logger.Error("Failed to delete wallet", "error", err, "user_id", userID, "wallet_id", walletID)
		return err
	}
	
	// Invalidate cache
	s.invalidateUserCache(ctx, userID)
	
	s.logger.Info("Wallet removed from watchlist", "user_id", userID, "wallet_id", walletID)
	return nil
}

// AddToken adds a token to user's tracked tokens
func (s *watchlistService) AddToken(ctx context.Context, userID uint, req *AddTokenRequest) (*TokenResponse, error) {
	// Validate token address if provided
	if req.TokenAddress != nil && !s.web3Service.ValidateAddress(*req.TokenAddress) {
		return nil, ErrInvalidAddress
	}
	
	// Check if token already exists for this user
	tokens, err := s.watchlistRepo.GetTokensByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user tokens", "error", err, "user_id", userID)
		return nil, err
	}
	
	for _, token := range tokens {
		if token.TokenSymbol == req.TokenSymbol {
			return nil, ErrTokenAlreadyExists
		}
	}
	
	// Create token
	token := &models.TrackedToken{
		UserID:       userID,
		TokenAddress: req.TokenAddress,
		TokenSymbol:  req.TokenSymbol,
		TokenName:    req.TokenName,
	}
	
	if err := s.watchlistRepo.CreateToken(ctx, token); err != nil {
		s.logger.Error("Failed to create token", "error", err, "user_id", userID, "symbol", req.TokenSymbol)
		return nil, err
	}
	
	// Invalidate cache
	s.invalidateUserCache(ctx, userID)
	
	s.logger.Info("Token added to watchlist", "user_id", userID, "token_id", token.ID, "symbol", req.TokenSymbol)
	
	return &TokenResponse{
		ID:           token.ID,
		TokenAddress: token.TokenAddress,
		TokenSymbol:  token.TokenSymbol,
		TokenName:    token.TokenName,
		CreatedAt:    token.CreatedAt,
		UpdatedAt:    token.UpdatedAt,
	}, nil
}

// GetTokens retrieves user's tracked tokens
func (s *watchlistService) GetTokens(ctx context.Context, userID uint) ([]*TokenResponse, error) {
	tokens, err := s.watchlistRepo.GetTokensByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user tokens", "error", err, "user_id", userID)
		return nil, err
	}
	
	responses := make([]*TokenResponse, len(tokens))
	for i, token := range tokens {
		responses[i] = &TokenResponse{
			ID:           token.ID,
			TokenAddress: token.TokenAddress,
			TokenSymbol:  token.TokenSymbol,
			TokenName:    token.TokenName,
			CreatedAt:    token.CreatedAt,
			UpdatedAt:    token.UpdatedAt,
		}
	}
	
	return responses, nil
}

// DeleteToken removes a token from user's tracked tokens
func (s *watchlistService) DeleteToken(ctx context.Context, userID uint, tokenID uint) error {
	if err := s.watchlistRepo.DeleteToken(ctx, tokenID, userID); err != nil {
		s.logger.Error("Failed to delete token", "error", err, "user_id", userID, "token_id", tokenID)
		return err
	}
	
	// Invalidate cache
	s.invalidateUserCache(ctx, userID)
	
	s.logger.Info("Token removed from watchlist", "user_id", userID, "token_id", tokenID)
	return nil
}

// GetBalances retrieves user's wallet balances with caching
func (s *watchlistService) GetBalances(ctx context.Context, userID uint) ([]*BalanceResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user_balances:%d", userID)
	var cachedBalances []*BalanceResponse
	
	if err := s.cacheService.Get(ctx, cacheKey, &cachedBalances); err == nil {
		s.logger.Debug("Balances found in cache", "user_id", userID)
		return cachedBalances, nil
	}
	
	// Cache miss, get from database
	balances, err := s.watchlistRepo.GetLatestBalances(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get balances", "error", err, "user_id", userID)
		return nil, err
	}
	
	// Convert to response format
	responses := make([]*BalanceResponse, len(balances))
	for i, balance := range balances {
		responses[i] = &BalanceResponse{
			WalletID:      balance.WalletID,
			WalletAddress: balance.Wallet.WalletAddress,
			TokenID:       balance.TokenID,
			TokenSymbol:   balance.Token.TokenSymbol,
			Balance:       balance.Balance,
			BalanceUSD:    balance.BalanceUSD,
			FetchedAt:     balance.FetchedAt,
		}
	}
	
	// Cache the results
	if err := s.cacheService.Set(ctx, cacheKey, responses, 5*time.Minute); err != nil {
		s.logger.Warn("Failed to cache balances", "error", err, "user_id", userID)
	}
	
	return responses, nil
}

// RefreshBalances triggers a balance refresh for a user
func (s *watchlistService) RefreshBalances(ctx context.Context, userID uint) error {
	// Invalidate cache first
	s.invalidateUserCache(ctx, userID)
	
	// Trigger balance fetch
	if err := s.balanceFetcher.FetchBalancesForUser(ctx, userID); err != nil {
		s.logger.Error("Failed to refresh balances", "error", err, "user_id", userID)
		return err
	}
	
	s.logger.Info("Balances refreshed", "user_id", userID)
	return nil
}

// invalidateUserCache invalidates all cache entries for a user
func (s *watchlistService) invalidateUserCache(ctx context.Context, userID uint) {
	patterns := []string{
		fmt.Sprintf("user_balances:%d", userID),
		fmt.Sprintf("user_wallets:%d", userID),
		fmt.Sprintf("user_tokens:%d", userID),
	}
	
	for _, pattern := range patterns {
		if err := s.cacheService.Delete(ctx, pattern); err != nil {
			s.logger.Warn("Failed to invalidate cache", "pattern", pattern, "error", err)
		}
	}
}

// GetBalanceHistory retrieves balance history for a specific wallet and token
func (s *watchlistService) GetBalanceHistory(ctx context.Context, userID uint, walletID uint, tokenID uint, limit int) ([]*BalanceHistoryResponse, error) {
	// Verify the wallet belongs to the user
	wallet, err := s.watchlistRepo.GetWalletByID(ctx, walletID)
	if err != nil {
		s.logger.Error("Failed to get wallet", "error", err, "wallet_id", walletID)
		return nil, err
	}
	
	if wallet.UserID != userID {
		return nil, fmt.Errorf("wallet not found")
	}
	
	// Get balance history from repository
	balances, err := s.watchlistRepo.GetBalanceHistory(ctx, walletID, tokenID, limit)
	if err != nil {
		s.logger.Error("Failed to get balance history", "error", err, "wallet_id", walletID, "token_id", tokenID)
		return nil, err
	}
	
	// Get token info for symbol
	token, err := s.watchlistRepo.GetTokenByID(ctx, tokenID)
	if err != nil {
		s.logger.Error("Failed to get token", "error", err, "token_id", tokenID)
		return nil, err
	}
	
	// Convert to response format
	var history []*BalanceHistoryResponse
	for _, balance := range balances {
		history = append(history, &BalanceHistoryResponse{
			ID:            balance.ID,
			WalletID:      balance.WalletID,
			WalletAddress: wallet.WalletAddress,
			TokenID:       balance.TokenID,
			TokenSymbol:   token.TokenSymbol,
			Balance:       balance.Balance,
			BalanceUSD:    balance.BalanceUSD,
			FetchedAt:     balance.FetchedAt,
			CreatedAt:     balance.CreatedAt,
		})
	}
	
	return history, nil
} 