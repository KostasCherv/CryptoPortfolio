package repository

import (
	"context"
	"time"

	"cryptoportfolio/internal/models"

	"gorm.io/gorm"
)

// WatchlistRepository defines the interface for watchlist operations
type WatchlistRepository interface {
	// Wallet operations
	CreateWallet(ctx context.Context, wallet *models.WatchlistWallet) error
	GetWalletsByUserID(ctx context.Context, userID uint) ([]*models.WatchlistWallet, error)
	GetAllWallets(ctx context.Context) ([]*models.WatchlistWallet, error)
	GetWalletByID(ctx context.Context, walletID uint) (*models.WatchlistWallet, error)
	DeleteWallet(ctx context.Context, walletID uint, userID uint) error
	
	// Token operations
	CreateToken(ctx context.Context, token *models.TrackedToken) error
	GetTokensByUserID(ctx context.Context, userID uint) ([]*models.TrackedToken, error)
	GetAllTokens(ctx context.Context) ([]*models.TrackedToken, error)
	GetTokenByID(ctx context.Context, tokenID uint) (*models.TrackedToken, error)
	DeleteToken(ctx context.Context, tokenID uint, userID uint) error
	
	// Balance operations
	CreateBalance(ctx context.Context, balance *models.WalletBalance) error
	GetLatestBalances(ctx context.Context, userID uint) ([]*models.WalletBalance, error)
	GetBalanceHistory(ctx context.Context, walletID, tokenID uint, limit int) ([]*models.WalletBalance, error)
	DeleteOldBalances(ctx context.Context, olderThan time.Duration) error
}

// watchlistRepository implements WatchlistRepository
type watchlistRepository struct {
	db *gorm.DB
}

// NewWatchlistRepository creates a new watchlist repository
func NewWatchlistRepository(db *gorm.DB) WatchlistRepository {
	return &watchlistRepository{db: db}
}

// CreateWallet creates a new wallet in the watchlist
func (r *watchlistRepository) CreateWallet(ctx context.Context, wallet *models.WatchlistWallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

// GetWalletsByUserID retrieves all wallets for a user
func (r *watchlistRepository) GetWalletsByUserID(ctx context.Context, userID uint) ([]*models.WatchlistWallet, error) {
	var wallets []*models.WatchlistWallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&wallets).Error
	return wallets, err
}

// GetAllWallets retrieves all wallets from all users
func (r *watchlistRepository) GetAllWallets(ctx context.Context) ([]*models.WatchlistWallet, error) {
	var wallets []*models.WatchlistWallet
	err := r.db.WithContext(ctx).Find(&wallets).Error
	return wallets, err
}

// GetWalletByID retrieves a wallet by ID
func (r *watchlistRepository) GetWalletByID(ctx context.Context, walletID uint) (*models.WatchlistWallet, error) {
	var wallet models.WatchlistWallet
	err := r.db.WithContext(ctx).Where("id = ?", walletID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// DeleteWallet deletes a wallet from the watchlist
func (r *watchlistRepository) DeleteWallet(ctx context.Context, walletID uint, userID uint) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", walletID, userID).Delete(&models.WatchlistWallet{}).Error
}

// CreateToken creates a new tracked token
func (r *watchlistRepository) CreateToken(ctx context.Context, token *models.TrackedToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetTokensByUserID retrieves all tracked tokens for a user
func (r *watchlistRepository) GetTokensByUserID(ctx context.Context, userID uint) ([]*models.TrackedToken, error) {
	var tokens []*models.TrackedToken
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

// GetAllTokens retrieves all tracked tokens from all users
func (r *watchlistRepository) GetAllTokens(ctx context.Context) ([]*models.TrackedToken, error) {
	var tokens []*models.TrackedToken
	err := r.db.WithContext(ctx).Find(&tokens).Error
	return tokens, err
}

// GetTokenByID retrieves a token by ID
func (r *watchlistRepository) GetTokenByID(ctx context.Context, tokenID uint) (*models.TrackedToken, error) {
	var token models.TrackedToken
	err := r.db.WithContext(ctx).Where("id = ?", tokenID).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteToken deletes a tracked token
func (r *watchlistRepository) DeleteToken(ctx context.Context, tokenID uint, userID uint) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", tokenID, userID).Delete(&models.TrackedToken{}).Error
}

// CreateBalance creates a new balance record
func (r *watchlistRepository) CreateBalance(ctx context.Context, balance *models.WalletBalance) error {
	return r.db.WithContext(ctx).Create(balance).Error
}

// GetLatestBalances retrieves the latest balance for each wallet-token combination for a user
func (r *watchlistRepository) GetLatestBalances(ctx context.Context, userID uint) ([]*models.WalletBalance, error) {
	var balances []*models.WalletBalance
	
	// Subquery to get the latest balance for each wallet-token combination
	subquery := r.db.Model(&models.WalletBalance{}).
		Select("wallet_id, token_id, MAX(fetched_at) as max_fetched_at").
		Joins("JOIN watchlist_wallets ON wallet_balances.wallet_id = watchlist_wallets.id").
		Where("watchlist_wallets.user_id = ?", userID).
		Group("wallet_id, token_id")
	
	err := r.db.WithContext(ctx).
		Joins("JOIN (?) as latest ON wallet_balances.wallet_id = latest.wallet_id AND wallet_balances.token_id = latest.token_id AND wallet_balances.fetched_at = latest.max_fetched_at", subquery).
		Preload("Wallet").
		Preload("Token").
		Find(&balances).Error
	
	return balances, err
}

// GetBalanceHistory retrieves balance history for a wallet-token combination
func (r *watchlistRepository) GetBalanceHistory(ctx context.Context, walletID, tokenID uint, limit int) ([]*models.WalletBalance, error) {
	var balances []*models.WalletBalance
	err := r.db.WithContext(ctx).
		Where("wallet_id = ? AND token_id = ?", walletID, tokenID).
		Order("fetched_at DESC").
		Limit(limit).
		Find(&balances).Error
	return balances, err
}

// DeleteOldBalances removes balance records older than the specified duration
func (r *watchlistRepository) DeleteOldBalances(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return r.db.WithContext(ctx).Where("fetched_at < ?", cutoff).Delete(&models.WalletBalance{}).Error
} 