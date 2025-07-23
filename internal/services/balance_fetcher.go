package services

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"simple_api/internal/cache"
	"simple_api/internal/config"
	"simple_api/internal/models"
	"simple_api/internal/repository"
	"simple_api/pkg/logger"
)

// BalanceFetcherService handles background balance fetching
type BalanceFetcherService interface {
	Start(ctx context.Context)
	Stop()
	FetchBalancesForUser(ctx context.Context, userID uint) error
}

// balanceFetcherService implements BalanceFetcherService
type balanceFetcherService struct {
	watchlistRepo repository.WatchlistRepository
	web3Service    Web3Service
	cacheService   cache.CacheProvider
	logger         *logger.Logger
	config         *config.Config
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// NewBalanceFetcherService creates a new balance fetcher service
func NewBalanceFetcherService(
	watchlistRepo repository.WatchlistRepository,
	web3Service Web3Service,
	cacheService cache.CacheProvider,
	logger *logger.Logger,
	config *config.Config,
) BalanceFetcherService {
	return &balanceFetcherService{
		watchlistRepo: watchlistRepo,
		web3Service:    web3Service,
		cacheService:   cacheService,
		logger:         logger,
		config:         config,
		stopChan:       make(chan struct{}),
	}
}

// Start begins the background balance fetching process
func (bfs *balanceFetcherService) Start(ctx context.Context) {
	bfs.logger.Info("Starting background balance fetcher")
	
	// Start the main balance fetching goroutine
	bfs.wg.Add(1)
	go bfs.runBalanceFetcher(ctx)
	
	// Start the cleanup goroutine
	bfs.wg.Add(1)
	go bfs.runCleanup(ctx)
}

// Stop gracefully stops the balance fetcher
func (bfs *balanceFetcherService) Stop() {
	bfs.logger.Info("Stopping background balance fetcher")
	close(bfs.stopChan)
	bfs.wg.Wait()
	bfs.logger.Info("Background balance fetcher stopped")
}

// runBalanceFetcher runs the main balance fetching loop
func (bfs *balanceFetcherService) runBalanceFetcher(ctx context.Context) {
	defer bfs.wg.Done()
	
	ticker := time.NewTicker(time.Duration(bfs.config.Web3.FetchInterval) * time.Minute)
	defer ticker.Stop()
	
	// Fetch immediately on startup
	if err := bfs.fetchAllBalances(ctx); err != nil {
		bfs.logger.Error("Failed to fetch initial balances", "error", err)
	}
	
	for {
		select {
		case <-ticker.C:
			if err := bfs.fetchAllBalances(ctx); err != nil {
				bfs.logger.Error("Failed to fetch balances", "error", err)
			}
		case <-bfs.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// runCleanup runs the cleanup process for old balance records
func (bfs *balanceFetcherService) runCleanup(ctx context.Context) {
	defer bfs.wg.Done()
	
	ticker := time.NewTicker(24 * time.Hour) // Cleanup daily
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Delete balances older than 30 days
			if err := bfs.watchlistRepo.DeleteOldBalances(ctx, 30*24*time.Hour); err != nil {
				bfs.logger.Error("Failed to cleanup old balances", "error", err)
			} else {
				bfs.logger.Info("Cleaned up old balance records")
			}
		case <-bfs.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// fetchAllBalances fetches balances for all users
func (bfs *balanceFetcherService) fetchAllBalances(ctx context.Context) error {
	// Create a context with timeout for the entire operation
	fetchCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	
	// Get all wallets and tokens from the database
	wallets, err := bfs.watchlistRepo.GetAllWallets(fetchCtx)
	if err != nil {
		return fmt.Errorf("failed to get wallets: %w", err)
	}
	
	tokens, err := bfs.watchlistRepo.GetAllTokens(fetchCtx)
	if err != nil {
		return fmt.Errorf("failed to get tokens: %w", err)
	}
	
	bfs.logger.Infof("Starting balance fetch cycle - wallets: %d, tokens: %d", len(wallets), len(tokens))
	
	if len(wallets) == 0 || len(tokens) == 0 {
		bfs.logger.Info("No wallets or tokens to fetch balances for")
		return nil
	}
	
	// Use a worker pool to fetch balances concurrently
	maxWorkers := bfs.config.Web3.MaxWorkers
	taskChan := make(chan fetchTask, 100)
	resultChan := make(chan fetchResult, 100)
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go bfs.balanceWorker(fetchCtx, i, taskChan, resultChan, &wg)
	}
	
	// Send tasks to workers
	go func() {
		defer close(taskChan)
		
		for _, wallet := range wallets {
			for _, token := range tokens {
				// Only fetch if wallet and token belong to the same user
				if wallet.UserID == token.UserID {
					task := fetchTask{
						walletAddress: wallet.WalletAddress,
						tokenAddress:  token.TokenAddress,
					}
					
					select {
					case taskChan <- task:
					case <-fetchCtx.Done():
						return
					}
				}
			}
		}
	}()
	
	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Collect results and store balances
	successCount := 0
	errorCount := 0
	
	for result := range resultChan {
		if result.err != nil {
			errorCount++
			bfs.logger.Error("Failed to fetch balance", 
				"wallet", result.walletAddress, 
				"token", result.tokenAddress, 
				"error", result.err)
		} else {
			// Store the balance in the database
			if err := bfs.storeBalance(fetchCtx, result); err != nil {
				errorCount++
				bfs.logger.Error("Failed to store balance", 
					"wallet", result.walletAddress, 
					"token", result.tokenAddress, 
					"error", err)
			} else {
				successCount++
				bfs.logger.Debug("Successfully fetched and stored balance", 
					"wallet", result.walletAddress, 
					"token", result.tokenAddress, 
					"balance", result.balance)
			}
		}
	}
	
	bfs.logger.Infof("Balance fetch cycle completed - successes: %d, errors: %d", successCount, errorCount)
	
	return nil
}

// fetchTask represents a balance fetching task
type fetchTask struct {
	walletAddress string
	tokenAddress  *string // nil for ETH
}

// fetchResult represents the result of a balance fetch
type fetchResult struct {
	walletAddress string
	tokenAddress  *string
	balance       *big.Int
	err           error
}

// balanceWorker processes balance fetching tasks
func (bfs *balanceFetcherService) balanceWorker(
	ctx context.Context,
	_ int, // workerID - unused but kept for future use
	taskChan <-chan fetchTask,
	resultChan chan<- fetchResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	
	for task := range taskChan {
		select {
		case <-ctx.Done():
			return
		default:
		}
		
		var balance *big.Int
		var err error
		
		// Fetch balance based on token type
		if task.tokenAddress == nil {
			// Fetch ETH balance
			balance, err = bfs.web3Service.GetETHBalance(ctx, task.walletAddress)
		} else {
			// Fetch token balance
			balance, err = bfs.web3Service.GetTokenBalance(ctx, *task.tokenAddress, task.walletAddress)
		}
		
		resultChan <- fetchResult{
			walletAddress: task.walletAddress,
			tokenAddress:  task.tokenAddress,
			balance:       balance,
			err:           err,
		}
		
		// Small delay to avoid overwhelming the RPC
		time.Sleep(100 * time.Millisecond)
	}
}

// FetchBalancesForUser fetches balances for a specific user
func (bfs *balanceFetcherService) FetchBalancesForUser(ctx context.Context, userID uint) error {
	// Get user's wallets
	wallets, err := bfs.watchlistRepo.GetWalletsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user wallets: %w", err)
	}
	
	// Get user's tracked tokens
	tokens, err := bfs.watchlistRepo.GetTokensByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}
	
	// Create a context with timeout
	fetchCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	
	// Fetch balances for each wallet-token combination
	for _, wallet := range wallets {
		for _, token := range tokens {
			if err := bfs.fetchAndStoreBalance(fetchCtx, wallet, token); err != nil {
				bfs.logger.Error("Failed to fetch balance", 
					"wallet", wallet.WalletAddress, 
					"token", token.TokenSymbol, 
					"error", err)
			}
		}
	}
	
	// Invalidate cache for this user
	cacheKey := fmt.Sprintf("user_balances:%d", userID)
	bfs.cacheService.Delete(ctx, cacheKey)
	
	return nil
}

// fetchAndStoreBalance fetches and stores a single balance
func (bfs *balanceFetcherService) fetchAndStoreBalance(
	ctx context.Context,
	wallet *models.WatchlistWallet,
	token *models.TrackedToken,
) error {
	var balance *big.Int
	var err error
	
	// Fetch balance
	if token.TokenAddress == nil {
		// ETH balance
		balance, err = bfs.web3Service.GetETHBalance(ctx, wallet.WalletAddress)
	} else {
		// Token balance
		balance, err = bfs.web3Service.GetTokenBalance(ctx, *token.TokenAddress, wallet.WalletAddress)
	}
	
	if err != nil {
		return err
	}
	
	// Create balance record
	balanceRecord := &models.WalletBalance{
		WalletID:  wallet.ID,
		TokenID:   token.ID,
		Balance:   balance.String(),
		FetchedAt: time.Now(),
	}
	
	// Store in database
	if err := bfs.watchlistRepo.CreateBalance(ctx, balanceRecord); err != nil {
		return fmt.Errorf("failed to store balance: %w", err)
	}
	
	// Cache the balance
	cacheKey := fmt.Sprintf("balance:%d:%d", wallet.ID, token.ID)
	cacheData := map[string]interface{}{
		"balance":    balance.String(),
		"fetched_at": time.Now().Unix(),
	}
	
	if err := bfs.cacheService.Set(ctx, cacheKey, cacheData, 10*time.Minute); err != nil {
		bfs.logger.Warn("Failed to cache balance", "error", err)
	}
	
	return nil
}

// storeBalance stores a fetched balance in the database
func (bfs *balanceFetcherService) storeBalance(ctx context.Context, result fetchResult) error {
	// Find the wallet and token by their addresses
	wallets, err := bfs.watchlistRepo.GetAllWallets(ctx)
	if err != nil {
		return fmt.Errorf("failed to get wallets: %w", err)
	}
	
	tokens, err := bfs.watchlistRepo.GetAllTokens(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tokens: %w", err)
	}
	
	// Find the wallet
	var wallet *models.WatchlistWallet
	for _, w := range wallets {
		if w.WalletAddress == result.walletAddress {
			wallet = w
			break
		}
	}
	
	if wallet == nil {
		return fmt.Errorf("wallet not found: %s", result.walletAddress)
	}
	
	// Find the token
	var token *models.TrackedToken
	for _, t := range tokens {
		if result.tokenAddress == nil {
			// ETH balance - look for token with nil address
			if t.TokenAddress == nil {
				token = t
				break
			}
		} else {
			// Token balance - look for token with matching address
			if t.TokenAddress != nil && *t.TokenAddress == *result.tokenAddress {
				token = t
				break
			}
		}
	}
	
	if token == nil {
		if result.tokenAddress == nil {
			return fmt.Errorf("ETH token not found in user's tracked tokens")
		}
		return fmt.Errorf("token not found: %s", *result.tokenAddress)
	}
	
	// Create balance record
	balanceRecord := &models.WalletBalance{
		WalletID:  wallet.ID,
		TokenID:   token.ID,
		Balance:   result.balance.String(),
		FetchedAt: time.Now(),
	}
	
	// Store in database
	if err := bfs.watchlistRepo.CreateBalance(ctx, balanceRecord); err != nil {
		return fmt.Errorf("failed to store balance: %w", err)
	}
	
	// Cache the balance
	cacheKey := fmt.Sprintf("balance:%d:%d", wallet.ID, token.ID)
	cacheData := map[string]interface{}{
		"balance":    result.balance.String(),
		"fetched_at": time.Now().Unix(),
	}
	
	if err := bfs.cacheService.Set(ctx, cacheKey, cacheData, 10*time.Minute); err != nil {
		bfs.logger.Warn("Failed to cache balance", "error", err)
	}
	
	return nil
}