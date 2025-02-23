package server

import (
	"net/http"

	"github.com/Amin-Abdi/ISA-Investment-project/internal/postgres"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

type Server struct {
	store *postgres.Store
}

func NewServer(store *postgres.Store) *Server {
	return &Server{
		store: store,
	}
}

func (s *Server) Start() error {
	r := gin.Default()

	// Run the server
	r.POST("/isa", s.CreateIsa)
	r.POST("/fund", s.CreateFund)

	r.PUT("/funds/:id", s.UpdateFund)
	r.PUT("/isa/:isa_id/fund/:fund_id", s.AddFundToIsa)

	r.GET("/isa/:id", s.GetIsa)
	r.GET("/funds", s.ListFunds)

	return r.Run(":8080")
}

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

	createdIsaID, err := s.store.CreateIsa(c.Request.Context(), isa)

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

func (s *Server) GetIsa(c *gin.Context) {
	isaID := c.Param("id")
	logger := logrus.New().WithContext(c.Request.Context())

	isa, err := s.store.GetIsa(c.Request.Context(), isaID)
	if err != nil {
		if err == postgres.ErrNotFound {
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

//Add money to ISA

//Deposit money into a fund, which in turn will subtract the cash balance from isa and add investments

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

	createdFundID, err := s.store.CreateFund(c.Request.Context(), fund)
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

func (s *Server) UpdateFund(c *gin.Context) {
	logger := logrus.New().WithContext(c.Request.Context())
	var req UpdateFundRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Invalid request payload for creating fund")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fundID := c.Param("id")

	updatedFund, err := s.store.UpdateFund(c.Request.Context(), fundID, req.Name, req.Description)
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

func (s *Server) ListFunds(c *gin.Context) {
	logger := logrus.New().WithContext(c.Request.Context())

	funds, err := s.store.ListFunds(c.Request.Context())
	if err != nil {
		logger.WithError(err).Error("Failed to retrieve funds")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"funds": funds,
	})
}

func (s *Server) AddFundToIsa(c *gin.Context) {
	logger := logrus.New().WithContext(c.Request.Context())
	isaID := c.Param("isa_id")
	fundID := c.Param("fund_id")

	updatedIsa, err := s.store.AddFundToISA(c.Request.Context(), isaID, fundID)
	if err != nil {
		if err == postgres.ErrNotFound {
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

//TODO:
//List Investments
//Get Investments

/*
When creating an investment, we need to make sure that the fund ID (the destination fund is related to the
ISA otherwise return an error)

*/
