package repository

import (
	"subscriptions-service/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(sub *models.Subscription) error {
	return r.db.Create(sub).Error
}

func (r *SubscriptionRepository) Get(id uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	if err := r.db.First(&sub, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriptionRepository) Update(id uuid.UUID, s *models.Subscription) error {
	result := r.db.
		Model(&models.Subscription{}).
		Where("id = ?", id).
		Omit("id").
		Updates(s)

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *SubscriptionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Subscription{}, "id = ?", id).Error
}

func (r *SubscriptionRepository) baseQuery(userID uuid.UUID, serviceName string, start, end time.Time) *gorm.DB {
	q := r.db.Model(&models.Subscription{}).Where("user_id = ?", userID)
	if serviceName != "" {
		q = q.Where("service_name = ?", serviceName)
	}
	return q.Where("start_date BETWEEN ? AND ?", start, end)
}

func (r *SubscriptionRepository) ListByUser(userID uuid.UUID, serviceName string, start, end time.Time) ([]models.Subscription, error) {
	var subs []models.Subscription
	if err := r.baseQuery(userID, serviceName, start, end).Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *SubscriptionRepository) SumByUserAndService(userID uuid.UUID, serviceName string, start, end time.Time) (int, error) {
	var sum int64
	err := r.baseQuery(userID, serviceName, start, end).Select("SUM(price)").Scan(&sum).Error
	return int(sum), err
}
