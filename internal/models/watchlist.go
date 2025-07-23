package models

import (
	"time"

	"gorm.io/gorm"
)

// WatchlistWallet represents a wallet address that a user wants to track
type WatchlistWallet struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null;index"`
	WalletAddress string         `json:"wallet_address" gorm:"not null;size:42;index"`
	Label         string         `json:"label" gorm:"size:100"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	User     User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Balances []WalletBalance  `json:"balances,omitempty" gorm:"foreignKey:WalletID"`
}

// TrackedToken represents a token that a user wants to track
type TrackedToken struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserID       uint           `json:"user_id" gorm:"not null;index"`
	TokenAddress *string        `json:"token_address" gorm:"size:42;index"` // null for native token (ETH)
	TokenSymbol  string         `json:"token_symbol" gorm:"not null;size:10"`
	TokenName    string         `json:"token_name" gorm:"not null;size:100"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	User     User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Balances []WalletBalance  `json:"balances,omitempty" gorm:"foreignKey:TokenID"`
}

// WalletBalance represents a balance snapshot for a wallet and token
type WalletBalance struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	WalletID     uint           `json:"wallet_id" gorm:"not null;index"`
	TokenID      uint           `json:"token_id" gorm:"not null;index"`
	Balance      string         `json:"balance" gorm:"not null;size:100"` // Store as string for precision
	BalanceUSD   *string        `json:"balance_usd" gorm:"size:100"`      // Optional USD value
	FetchedAt    time.Time      `json:"fetched_at" gorm:"not null;index"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Wallet WatchlistWallet `json:"wallet,omitempty" gorm:"foreignKey:WalletID"`
	Token  TrackedToken    `json:"token,omitempty" gorm:"foreignKey:TokenID"`
}

// TableName specifies the table name for WatchlistWallet
func (WatchlistWallet) TableName() string {
	return "watchlist_wallets"
}

// TableName specifies the table name for TrackedToken
func (TrackedToken) TableName() string {
	return "tracked_tokens"
}

// TableName specifies the table name for WalletBalance
func (WalletBalance) TableName() string {
	return "wallet_balances"
} 