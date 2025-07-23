package handlers

import (
	"net/http"
	"strconv"

	"cryptoportfolio/internal/services"
	"cryptoportfolio/pkg/logger"

	"github.com/gin-gonic/gin"
)

// SuccessResponse represents a successful operation response
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// WatchlistHandler handles watchlist-related HTTP requests
type WatchlistHandler struct {
	watchlistService services.WatchlistService
	logger           *logger.Logger
}

// NewWatchlistHandler creates a new watchlist handler
func NewWatchlistHandler(watchlistService services.WatchlistService, logger *logger.Logger) *WatchlistHandler {
	return &WatchlistHandler{
		watchlistService: watchlistService,
		logger:           logger,
	}
}

// AddWallet godoc
// @Summary Add wallet to watchlist
// @Description Add a new wallet address to the user's watchlist
// @Tags Watchlist
// @Accept json
// @Produce json
// @Param wallet body services.AddWalletRequest true "Wallet information"
// @Security BearerAuth
// @Success 201 {object} services.WalletResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/watchlist/wallets [post]
func (h *WatchlistHandler) AddWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.AddWalletRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
			return
		}

		userID := c.GetUint("user_id")
		wallet, err := h.watchlistService.AddWallet(c.Request.Context(), userID, &req)
		if err != nil {
			switch err {
			case services.ErrInvalidAddress:
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wallet address"})
			case services.ErrWalletAlreadyExists:
				c.JSON(http.StatusConflict, ErrorResponse{Error: "Wallet already exists in watchlist"})
			default:
				h.logger.Error("Failed to add wallet", "error", err, "user_id", userID)
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add wallet"})
			}
			return
		}

		c.JSON(http.StatusCreated, wallet)
	}
}

// GetWallets godoc
// @Summary Get user's watchlist wallets
// @Description Retrieve all wallet addresses in the user's watchlist
// @Tags Watchlist
// @Produce json
// @Security BearerAuth
// @Success 200 {array} services.WalletResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/watchlist/wallets [get]
func (h *WatchlistHandler) GetWallets() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		wallets, err := h.watchlistService.GetWallets(c.Request.Context(), userID)
		if err != nil {
			h.logger.Error("Failed to get wallets", "error", err, "user_id", userID)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get wallets"})
			return
		}

		c.JSON(http.StatusOK, wallets)
	}
}

// DeleteWallet godoc
// @Summary Remove wallet from watchlist
// @Description Remove a wallet address from the user's watchlist
// @Tags Watchlist
// @Produce json
// @Param id path int true "Wallet ID"
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/watchlist/wallets/{id} [delete]
func (h *WatchlistHandler) DeleteWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		walletIDStr := c.Param("id")
		walletID, err := strconv.ParseUint(walletIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wallet ID"})
			return
		}

		userID := c.GetUint("user_id")
		err = h.watchlistService.DeleteWallet(c.Request.Context(), userID, uint(walletID))
		if err != nil {
			h.logger.Error("Failed to delete wallet", "error", err, "user_id", userID, "wallet_id", walletID)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete wallet"})
			return
		}

		c.JSON(http.StatusOK, SuccessResponse{Message: "Wallet removed from watchlist"})
	}
}

// AddToken godoc
// @Summary Add token to watchlist
// @Description Add a new token to the user's tracked tokens
// @Tags Watchlist
// @Accept json
// @Produce json
// @Param token body services.AddTokenRequest true "Token information"
// @Security BearerAuth
// @Success 201 {object} services.TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/watchlist/tokens [post]
func (h *WatchlistHandler) AddToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.AddTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
			return
		}

		userID := c.GetUint("user_id")
		token, err := h.watchlistService.AddToken(c.Request.Context(), userID, &req)
		if err != nil {
			switch err {
			case services.ErrInvalidAddress:
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid token address"})
			case services.ErrTokenAlreadyExists:
				c.JSON(http.StatusConflict, ErrorResponse{Error: "Token already exists in watchlist"})
			default:
				h.logger.Error("Failed to add token", "error", err, "user_id", userID)
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add token"})
			}
			return
		}

		c.JSON(http.StatusCreated, token)
	}
}

// GetTokens godoc
// @Summary Get user's tracked tokens
// @Description Retrieve all tokens in the user's watchlist
// @Tags Watchlist
// @Produce json
// @Security BearerAuth
// @Success 200 {array} services.TokenResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/watchlist/tokens [get]
func (h *WatchlistHandler) GetTokens() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		tokens, err := h.watchlistService.GetTokens(c.Request.Context(), userID)
		if err != nil {
			h.logger.Error("Failed to get tokens", "error", err, "user_id", userID)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get tokens"})
			return
		}

		c.JSON(http.StatusOK, tokens)
	}
}

// DeleteToken godoc
// @Summary Remove token from watchlist
// @Description Remove a token from the user's tracked tokens
// @Tags Watchlist
// @Produce json
// @Param id path int true "Token ID"
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/watchlist/tokens/{id} [delete]
func (h *WatchlistHandler) DeleteToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenIDStr := c.Param("id")
		tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid token ID"})
			return
		}

		userID := c.GetUint("user_id")
		err = h.watchlistService.DeleteToken(c.Request.Context(), userID, uint(tokenID))
		if err != nil {
			h.logger.Error("Failed to delete token", "error", err, "user_id", userID, "token_id", tokenID)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete token"})
			return
		}

		c.JSON(http.StatusOK, SuccessResponse{Message: "Token removed from watchlist"})
	}
}

// GetBalances godoc
// @Summary Get wallet balances
// @Description Retrieve current balances for all wallets and tokens in the user's watchlist
// @Tags Watchlist
// @Produce json
// @Security BearerAuth
// @Success 200 {array} services.BalanceResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/watchlist/balances [get]
func (h *WatchlistHandler) GetBalances() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		balances, err := h.watchlistService.GetBalances(c.Request.Context(), userID)
		if err != nil {
			h.logger.Error("Failed to get balances", "error", err, "user_id", userID)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get balances"})
			return
		}

		c.JSON(http.StatusOK, balances)
	}
}

// GetBalanceHistory godoc
// @Summary Get wallet balance history
// @Description Retrieve balance history for a specific wallet and token
// @Tags Watchlist
// @Produce json
// @Param wallet_id path int true "Wallet ID"
// @Param token_id path int true "Token ID"
// @Param limit query int false "Number of records to return (default: 50, max: 100)"
// @Security BearerAuth
// @Success 200 {array} services.BalanceHistoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/watchlist/wallets/{wallet_id}/tokens/{token_id}/history [get]
func (h *WatchlistHandler) GetBalanceHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		walletIDStr := c.Param("wallet_id")
		tokenIDStr := c.Param("token_id")
		limitStr := c.DefaultQuery("limit", "50")
		
		walletID, err := strconv.ParseUint(walletIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wallet ID"})
			return
		}
		
		tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid token ID"})
			return
		}
		
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 50
		}
		if limit > 100 {
			limit = 100
		}
		
		userID := c.GetUint("user_id")
		history, err := h.watchlistService.GetBalanceHistory(c.Request.Context(), userID, uint(walletID), uint(tokenID), limit)
		if err != nil {
			h.logger.Error("Failed to get balance history", "error", err, "user_id", userID, "wallet_id", walletID, "token_id", tokenID)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get balance history"})
			return
		}
		
		c.JSON(http.StatusOK, history)
	}
}

// RefreshBalances godoc
// @Summary Refresh wallet balances
// @Description Trigger a manual refresh of wallet balances from the blockchain
// @Tags Watchlist
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/watchlist/balances/refresh [post]
func (h *WatchlistHandler) RefreshBalances() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		err := h.watchlistService.RefreshBalances(c.Request.Context(), userID)
		if err != nil {
			h.logger.Error("Failed to refresh balances", "error", err, "user_id", userID)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to refresh balances"})
			return
		}

		c.JSON(http.StatusOK, SuccessResponse{Message: "Balance refresh initiated"})
	}
} 