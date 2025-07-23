package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"simple_api/internal/config"
	"simple_api/pkg/logger"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Web3Service handles blockchain interactions
type Web3Service interface {
	GetETHBalance(ctx context.Context, address string) (*big.Int, error)
	GetTokenBalance(ctx context.Context, tokenAddress, walletAddress string) (*big.Int, error)
	ValidateAddress(address string) bool
}

// web3Service implements Web3Service
type web3Service struct {
	client     *ethclient.Client
	config     *config.Config
	logger     *logger.Logger
	rateLimiter *RateLimiter
}

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	tokens chan struct{}
	ticker *time.Ticker
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, rate),
		ticker: time.NewTicker(time.Second / time.Duration(rate)),
	}
	
	go rl.refill()
	return rl
}

// refill adds tokens to the bucket
func (rl *RateLimiter) refill() {
	for range rl.ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Bucket is full
		}
	}
}

// Wait waits for a token to be available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// NewWeb3Service creates a new Web3 service
func NewWeb3Service(config *config.Config, logger *logger.Logger) (Web3Service, error) {
	// Connect to Ethereum client
	// log the rpc endpoint
	client, err := ethclient.Dial(config.Web3.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}

	// Create rate limiter (10 requests per second)
	rateLimiter := NewRateLimiter(10)

	return &web3Service{
		client:      client,
		config:      config,
		logger:      logger,
		rateLimiter: rateLimiter,
	}, nil
}

// GetETHBalance retrieves ETH balance with retry mechanism
func (s *web3Service) GetETHBalance(ctx context.Context, address string) (*big.Int, error) {
	if !s.ValidateAddress(address) {
		return nil, errors.New("invalid Ethereum address")
	}

	// Wait for rate limiter
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Retry with exponential backoff
	var balance *big.Int
	var err error
	
	for attempt := 1; attempt <= 3; attempt++ {
		balance, err = s.fetchETHBalance(ctx, address)
		if err == nil {
			return balance, nil
		}

		s.logger.Warn("Failed to fetch ETH balance", 
			"address", address, 
			"attempt", attempt, 
			"error", err)

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Exponential backoff: 1s, 2s, 4s
		if attempt < 3 {
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, fmt.Errorf("failed to fetch ETH balance after 3 attempts: %w", err)
}

// fetchETHBalance performs the actual ETH balance fetch
func (s *web3Service) fetchETHBalance(ctx context.Context, address string) (*big.Int, error) {
	addr := common.HexToAddress(address)
	balance, err := s.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

// GetTokenBalance retrieves ERC-20 token balance with retry mechanism
func (s *web3Service) GetTokenBalance(ctx context.Context, tokenAddress, walletAddress string) (*big.Int, error) {
	if !s.ValidateAddress(tokenAddress) || !s.ValidateAddress(walletAddress) {
		return nil, errors.New("invalid address")
	}

	// Wait for rate limiter
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Retry with exponential backoff
	var balance *big.Int
	var err error
	
	for attempt := 1; attempt <= 3; attempt++ {
		balance, err = s.fetchTokenBalance(ctx, tokenAddress, walletAddress)
		if err == nil {
			return balance, nil
		}

		s.logger.Warn("Failed to fetch token balance", 
			"token", tokenAddress, 
			"wallet", walletAddress, 
			"attempt", attempt, 
			"error", err)

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Exponential backoff: 1s, 2s, 4s
		if attempt < 3 {
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, fmt.Errorf("failed to fetch token balance after 3 attempts: %w", err)
}

// fetchTokenBalance performs the actual token balance fetch
func (s *web3Service) fetchTokenBalance(ctx context.Context, tokenAddress, walletAddress string) (*big.Int, error) {
	// ERC-20 balanceOf function signature
	balanceOfSignature := []byte("balanceOf(address)")
	hash := crypto.Keccak256(balanceOfSignature)
	methodID := hash[:4]

	// Pack the address parameter
	addr := common.HexToAddress(walletAddress)
	paddedAddress := common.LeftPadBytes(addr.Bytes(), 32)

	// Create the call data
	data := append(methodID, paddedAddress...)

	// Make the call
	tokenAddr := common.HexToAddress(tokenAddress)
	
	result, err := s.client.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddr,
		Data: data,
	}, nil)
	
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	// Parse the result
	balance := new(big.Int).SetBytes(result)
	return balance, nil
}

// ValidateAddress validates Ethereum address format
func (s *web3Service) ValidateAddress(address string) bool {
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	
	if len(address) != 42 {
		return false
	}
	
	// Check if it's a valid hex string
	_ = common.HexToAddress(address)
	return true
}

// Close closes the Web3 service
func (s *web3Service) Close() {
	if s.client != nil {
		s.client.Close()
	}
	if s.rateLimiter != nil && s.rateLimiter.ticker != nil {
		s.rateLimiter.ticker.Stop()
	}
} 