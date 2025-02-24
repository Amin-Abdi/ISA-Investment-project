package server

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/Amin-Abdi/ISA-Investment-project/internal/postgres"
)

type StoreInterface interface {
	CreateIsa(ctx context.Context, isa postgres.ISA) (string, error)
	GetIsa(ctx context.Context, id string) (*postgres.ISA, error)
	UpdateIsa(ctx context.Context, isaID string, cashBalance, investmentAmount float64) (*postgres.ISA, error)
	AddFundToISA(ctx context.Context, isaID, fundID string) (*postgres.ISA, error)
	CreateFund(ctx context.Context, fund postgres.Fund) (string, error)
	GetFund(ctx context.Context, id string) (*postgres.Fund, error)
	UpdateFund(ctx context.Context, id, name, description string) (*postgres.Fund, error)
	UpdateFundTotalAmount(ctx context.Context, fundID string, totalAmount float64) (*postgres.Fund, error)
	ListFunds(ctx context.Context) ([]postgres.Fund, error)
	CreateInvestment(ctx context.Context, investment postgres.Investment) (string, error)
	GetInvestment(ctx context.Context, investmentID string) (*postgres.Investment, error)
	ListInvestments(ctx context.Context, isaID string) ([]postgres.Investment, error)
}

type Server struct {
	Store StoreInterface
}

func NewServer(store *postgres.Store) *Server {
	return &Server{
		Store: store,
	}
}

func (s *Server) Start() error {
	r := gin.Default()

	// Run the server
	r.POST("/isa", s.CreateIsa)
	r.POST("/fund", s.CreateFund)
	r.POST("/isa/:id/invest", s.InvestIntoFund)

	r.PUT("/funds/:id", s.UpdateFund)
	r.PUT("/isa/:isa_id/fund/:fund_id", s.AddFundToIsa)

	r.GET("/isa/:id", s.GetIsa)
	r.GET("/funds", s.ListFunds)
	r.GET("/investments/:isa_id", s.ListInvestments)

	return r.Run(":8080")
}

// CreateIsa Creates an isa
func (s *Server) CreateIsa(c *gin.Context) {
	logger := logrus.New().WithContext(c.Request.Context())
	var req CreateISARequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Invalid request payload for creating ISA")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Generate a new UUID for the ISA
	isaID := uuid.New().String()
	isa := postgres.ISA{
		ID:               isaID,
		UserID:           req.UserID,
		CashBalance:      req.CashBalance,
		InvestmentAmount: 0, //Opening a new ISA, the invested amount will be 0.
	}

	createdIsaID, err := s.Store.CreateIsa(c.Request.Context(), isa)

	if err != nil {
		logger.WithError(err).Error("Failed to create ISA")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithField("created_isa_id", createdIsaID).Info("Isa has been successfully created")

	c.JSON(http.StatusCreated, gin.H{
		"message": "Isa successfully created",
		"isa_id":  createdIsaID,
	})
}

// GetIsa fetches an Isa
func (s *Server) GetIsa(c *gin.Context) {
	isaID := c.Param("id")
	logger := logrus.New().WithContext(c.Request.Context())

	isa, err := s.Store.GetIsa(c.Request.Context(), isaID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			logger.WithError(err).Error("Failed to find Isa")
			c.JSON(http.StatusNotFound, gin.H{"error": "Isa not found. Please check the id and try again."})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"isa": isa,
	})
}

// CreateFund creates a new fund
func (s *Server) CreateFund(c *gin.Context) {
	var req CreateFundRequest
	logger := logrus.New().WithContext(c.Request.Context())

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Invalid request payload for creating fund")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fundID := uuid.New().String()

	fund := postgres.Fund{
		ID:          fundID,
		Name:        req.Name,
		Description: req.Description,
		Type:        postgres.FundType(req.Type),
		RiskLevel:   postgres.RiskLevel(req.RiskLevel),
		Performance: req.Performance,
		TotalAmount: req.TotalAmount,
	}

	createdFundID, err := s.Store.CreateFund(c.Request.Context(), fund)
	if err != nil {
		logger.WithError(err).Error("Failed to create fund")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithField("created_fund_id", createdFundID).Info("Fund has been successfully created")
	c.JSON(http.StatusCreated, gin.H{
		"message": "Fund successfully created",
		"isa_id":  createdFundID,
	})
}

// UpdateFund updating the name/description of the fund
func (s *Server) UpdateFund(c *gin.Context) {
	logger := logrus.New().WithContext(c.Request.Context())
	var req UpdateFundRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Invalid request payload for creating fund")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fundID := c.Param("id")

	updatedFund, err := s.Store.UpdateFund(c.Request.Context(), fundID, req.Name, req.Description)
	if err != nil {
		if err == postgres.ErrNotFound {
			logger.WithError(err).Error("Failed to find fund")
			c.JSON(http.StatusNotFound, gin.H{"error": "Fund not found"})
		} else {
			logger.WithError(err).Error("Internal server error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	logger.WithField("updated_fund_id", updatedFund.ID).Info("Fund has been successfully updated")

	c.JSON(http.StatusOK, gin.H{
		"message": "Fund successfully updated",
		"fund":    updatedFund,
	})
}

// ListFunds lists all the avalaible funds
func (s *Server) ListFunds(c *gin.Context) {
	logger := logrus.New().WithContext(c.Request.Context())

	funds, err := s.Store.ListFunds(c.Request.Context())
	if err != nil {
		logger.WithError(err).Error("Failed to retrieve funds")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"funds": funds,
	})
}

// AddFundToIsa: Adds a fund to an Isa
func (s *Server) AddFundToIsa(c *gin.Context) {
	logger := logrus.New().WithContext(c.Request.Context())
	isaID := c.Param("isa_id")
	fundID := c.Param("fund_id")

	logger = logger.WithFields(logrus.Fields{
		"isa_id":  isaID,
		"fund_id": fundID,
	})

	getIsa, err := s.Store.GetIsa(c.Request.Context(), isaID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			logger.WithError(err).Error("Failed to find Isa")
			c.JSON(http.StatusNotFound, gin.H{"error": "Isa not found. Please check the id and try again."})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if len(getIsa.FundIDs) > 0 {
		logger.Error("ISA already has a fund associated with it. Cannot add another fund.")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "This ISA already has a fund associated with it. Only one fund can be added.",
		})
		return
	}

	updatedIsa, err := s.Store.AddFundToISA(c.Request.Context(), isaID, fundID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			logger.WithError(err).Error("Failed to find Isa")
			c.JSON(http.StatusNotFound, gin.H{"error": "Isa not found. Please check the id and try again."})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Fund successfully added to ISA",
		"updated_isa": updatedIsa,
	})
}

// InvestIntoFund adds the investment money to the fund from the isa
func (s *Server) InvestIntoFund(c *gin.Context) {
	logger := logrus.New().WithContext(c.Request.Context())
	var req InvestIntoFundRequest
	isaID := c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Invalid investment request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. Fund ID and amount are required."})
		return
	}

	isa, err := s.Store.GetIsa(c.Request.Context(), isaID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			logger.WithError(err).Error("Failed to find Isa")
			c.JSON(http.StatusNotFound, gin.H{"error": "Isa not found. Please check the id and try again."})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if req.Amount > isa.CashBalance {
		logger.Warn("Insufficient cash balance to make this investment")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance for this investment. Please add funds to your account and try again"})
		return
	}

	// Check if fund is selected in the isa.
	fundExists := slices.Contains(isa.FundIDs, req.FundID)
	if !fundExists {
		logger.Warn("Fund not associated with ISA")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fund not found in your ISA. Please add it before investing."})
		return
	}

	investment := postgres.Investment{
		ID:     uuid.NewString(),
		ISAID:  isaID,
		FundID: req.FundID,
		Amount: req.Amount,
	}

	//update the isa: Deduct cash balance and increase investment amount
	newCashBalance := isa.CashBalance - req.Amount
	newInvestmentAmount := isa.InvestmentAmount + req.Amount

	_, err = s.Store.UpdateIsa(c.Request.Context(), isaID, newCashBalance, newInvestmentAmount)
	if err != nil {
		logger.WithError(err).Error("Failed to update ISA")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fund, err := s.Store.GetFund(c.Request.Context(), req.FundID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			logger.WithError(err).Error("Failed to find fund")
			c.JSON(http.StatusNotFound, gin.H{"error": "Fund not found. Please check the id and try again."})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	totalAmount := fund.TotalAmount + req.Amount
	_, err = s.Store.UpdateFundTotalAmount(c.Request.Context(), req.FundID, totalAmount)
	if err != nil {
		logger.WithError(err).Error("Failed to update Fund total amount")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	investmentID, err := s.Store.CreateInvestment(c.Request.Context(), investment)
	if err != nil {
		logger.WithError(err).Error("Failed to create investment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"isa_id":  investment.ISAID,
		"fund_id": investment.FundID,
	})

	logger.Info("Investment has been successfully made")
	c.JSON(http.StatusOK, gin.H{
		"investment_id": investmentID,
	})
}

// ListInvestments lists the investments made in an isa
func (s *Server) ListInvestments(c *gin.Context) {
	logger := logrus.New().WithContext(c.Request.Context())
	isaID := c.Param("isa_id")

	investments, err := s.Store.ListInvestments(c.Request.Context(), isaID)
	if err != nil {
		logger.WithError(err).Error("Failed to list investments")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"investments": investments,
	})
}
