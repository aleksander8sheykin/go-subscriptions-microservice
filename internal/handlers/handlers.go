package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"subscriptions-service/internal/models"
	"subscriptions-service/internal/repository"
	"time"

	"subscriptions-service/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	repo repository.SubscriptionRepository
}

func NewHandler(repo repository.SubscriptionRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/subscriptions", h.CreateSubscription)
	r.GET("/subscriptions/:id", h.GetSubscription)
	r.PUT("/subscriptions/:id", h.UpdateSubscription)
	r.DELETE("/subscriptions/:id", h.DeleteSubscription)
	r.GET("/subscriptions/list", h.GetSubscriptionList)
	r.GET("/subscriptions/sum", h.GetSubscriptionSum)
}

// CreateSubscription godoc
// @Summary Создать подписку
// @Description Создает новую запись подписки
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.Subscription true "Subscription data"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (h *Handler) CreateSubscription(c *gin.Context) {
	ctx := c.Request.Context()

	var sub models.Subscription
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Create(ctx, &sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	traceID, _ := c.Get("trace_id")
	logger.Log.Info("Подписка создана", "trace_id", traceID, "subscription", sub)

	c.JSON(http.StatusCreated, sub)
}

func (h *Handler) fetchSubscription(c *gin.Context, id uuid.UUID) (*models.Subscription, bool) {
	ctx := c.Request.Context()
	sub, err := h.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return nil, false // false означает, что объект не получен
	}
	return sub, true
}

// GetSubscription godoc
// @Summary Получить подписку по ID
// @Description Получаем подписку по уникальному ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "UUID подписки"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *Handler) GetSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	sub, ok := h.fetchSubscription(c, id)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, sub)
}

// UpdateSubscription godoc
// @Summary Обновить подписку
// @Description Обновляет данные подписки по ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "UUID подписки"
// @Param subscription body models.Subscription true "Updated subscription"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [put]
func (h *Handler) UpdateSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var sub models.Subscription
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.repo.Update(ctx, id, &sub); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sub_updated, ok := h.fetchSubscription(c, id)
	if !ok {
		return
	}
	traceID, _ := c.Get("trace_id")
	logger.Log.Info("Подписка обновлена", "trace_id", traceID, "subscription", sub_updated)

	c.JSON(http.StatusOK, sub_updated)
}

// DeleteSubscription godoc
// @Summary Удалить подписку
// @Description Удаляет подписку по ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "UUID подписки"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	sub, ok := h.fetchSubscription(c, id)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	if err := h.repo.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	traceID, _ := c.Get("trace_id")
	logger.Log.Info("Подписка удалена", "trace_id", traceID, "subscription", sub)

	c.JSON(http.StatusOK, gin.H{"message": "subscription deleted"})
}

type QueryParams struct {
	UserID      uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	ServiceName *string
}

func parseQueryParams(c *gin.Context) (*QueryParams, error) {
	userID, err := uuid.Parse(c.Query("user_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid user_id")
	}

	var startDate *time.Time
	if v := c.Query("start_date"); v != "" {
		t, err := time.Parse("01-2006", v)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date")
		}
		startDate = &t
	}

	var endDate *time.Time
	if v := c.Query("end_date"); v != "" {
		t, err := time.Parse("01-2006", v)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date")
		}
		endDate = &t
	}

	var serviceName *string
	if v := c.Query("service_name"); v != "" {
		serviceName = &v
	}

	return &QueryParams{
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
		ServiceName: serviceName,
	}, nil
}

// GetSubscriptionList godoc
// @Summary Получить список подписок пользователя
// @Description Список подписок пользователя за период с фильтрацией по сервису
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id query string true "UUID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param start_date query string false "Начало периода (MM-YYYY)"
// @Param end_date query string false "Конец периода (MM-YYYY)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 10)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {array} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/list [get]
func (h *Handler) GetSubscriptionList(c *gin.Context) {
	params, err := parseQueryParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	type PagesParams struct {
		Limit  int `form:"limit,default=10"`
		Offset int `form:"offset,default=0"`
	}
	var pages_params PagesParams
	if err := c.ShouldBindQuery(&pages_params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if pages_params.Limit < 1 || pages_params.Limit > 100 {
		pages_params.Limit = 10
	}
	if pages_params.Offset < 1 {
		pages_params.Offset = 0
	}

	ctx := c.Request.Context()
	subs, err := h.repo.ListByUser(ctx, params.UserID, params.ServiceName, params.StartDate, params.EndDate, pages_params.Limit, pages_params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subs)
}

// GetSubscriptionSum godoc
// @Summary Получить сумму подписок
// @Description Считает общую стоимость подписок пользователя по сервису за период
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id query string true "UUID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param start_date query string false "Начало периода (MM-YYYY)"
// @Param end_date query string false "Конец периода (MM-YYYY)"
// @Success 200 {object} map[string]int
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/sum [get]
func (h *Handler) GetSubscriptionSum(c *gin.Context) {
	params, err := parseQueryParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	sum, err := h.repo.SumByUserAndService(ctx, params.UserID, params.ServiceName, params.StartDate, params.EndDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sum": sum})
}
