package rest

import (
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wsppppp/data-aggregation/internal/domain"
	"github.com/wsppppp/data-aggregation/internal/repository"
	"github.com/wsppppp/data-aggregation/internal/service"
)

const MonthYearLayout = "01-2006" // это шаблон для парсинга даты из строки

type Handler struct {
	service *service.SubscriptionService
}

func NewHandler(service *service.SubscriptionService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.Default()

	router.GET("/swagger/openapi.yaml", func(c *gin.Context) {
		c.File("docs/openapi.yaml")
	})

	router.GET("/docs/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger/openapi.yaml"),
	))

	api := router.Group("/api/v1")
	{
		api.POST("/subscriptions", h.createSubscription)
		api.GET("/subscriptions/:id", h.getSubscription)
		api.PUT("/subscriptions/:id", h.updateSubscription)
		api.DELETE("/subscriptions/:id", h.deleteSubscription)
		api.GET("/subscriptions", h.listSubscriptions)

		api.GET("/subscriptions/total", h.totalCost)
	}

	return router
}

func (h *Handler) createSubscription(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedDate, err := time.Parse(MonthYearLayout, req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected MM-YYYY"})
		return
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	input := service.CreateSubscriptionInput{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userUUID,
		StartDate:   parsedDate,
	}

	id, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		slog.Error("failed to create subscription", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *Handler) getSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	sub, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		slog.Error("failed to get subscription", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, toSubscriptionResponse(sub)) // для правильного отображения в API
}

func (h *Handler) updateSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := time.Parse(MonthYearLayout, req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected MM-YYYY"})
		return
	}

	var endDate *time.Time
	if req.EndDate != nil {
		ed, err := time.Parse(MonthYearLayout, *req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected MM-YYYY"})
			return
		}
		endDate = &ed
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	sub := &domain.Subscription{
		ID:          id,
		UserID:      userUUID,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := h.service.Update(c.Request.Context(), sub); err != nil {
		slog.Error("failed to update subscription", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) deleteSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		slog.Error("failed to delete subscription", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) listSubscriptions(c *gin.Context) {
	var filter repository.SubscriptionFilter

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &userID
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
	}
	if sn := c.Query("service_name"); sn != "" {
		filter.ServiceName = &sn
	}

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	items, err := h.service.List(c.Request.Context(), filter, limit, offset)
	if err != nil {
		slog.Error("failed to list subscriptions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := make([]SubscriptionResponse, 0, len(items))
	for i := range items {
		s := items[i]
		resp = append(resp, toSubscriptionResponse(&s))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) totalCost(c *gin.Context) {

	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to are required (MM-YYYY)"})
		return
	}

	from, err := time.Parse(MonthYearLayout, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from format, expected MM-YYYY"})
		return
	}
	to, err := time.Parse(MonthYearLayout, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to format, expected MM-YYYY"})
		return
	}

	var filter repository.SubscriptionFilter
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &userID
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
	}
	if sn := c.Query("service_name"); sn != "" {
		filter.ServiceName = &sn
	}

	total, err := h.service.TotalCost(c.Request.Context(), filter, from, to)
	if err != nil {
		slog.Error("failed to calc total", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total})
}
