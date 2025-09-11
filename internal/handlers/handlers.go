package handlers

import (
	"net/http"
	"subscriptions-service/internal/models"
	"subscriptions-service/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	repo *repository.SubscriptionRepository
}

func NewHandler(repo *repository.SubscriptionRepository) *Handler {
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
	var sub models.Subscription
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Create(&sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sub)
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

	sub, err := h.repo.Get(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	if err := h.repo.Update(id, &sub); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sub)
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

	if err := h.repo.Delete(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription deleted"})
}

// GetSubscriptionList godoc
// @Summary Получить список подписок пользователя
// @Description Список подписок пользователя за период с фильтрацией по сервису
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id query string true "UUID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param start_date query string true "Начало периода (MM-YYYY)"
// @Param end_date query string true "Конец периода (MM-YYYY)"
// @Success 200 {array} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/list [get]
func (h *Handler) GetSubscriptionList(c *gin.Context) {
	userID, err := uuid.Parse(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	start, err := time.Parse("01-2006", c.Query("start_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date"})
		return
	}

	end, err := time.Parse("01-2006", c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date"})
		return
	}

	serviceName := c.Query("service_name")
	subs, err := h.repo.ListByUser(userID, serviceName, start, end)
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
// @Param start_date query string true "Начало периода (MM-YYYY)"
// @Param end_date query string true "Конец периода (MM-YYYY)"
// @Success 200 {object} map[string]int
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/sum [get]
func (h *Handler) GetSubscriptionSum(c *gin.Context) {
	userID, err := uuid.Parse(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	start, err := time.Parse("01-2006", c.Query("start_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date"})
		return
	}

	end, err := time.Parse("01-2006", c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date"})
		return
	}

	serviceName := c.Query("service_name")
	sum, err := h.repo.SumByUserAndService(userID, serviceName, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sum": sum})
}
