package repository

import (
	"subscriptions-service/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"context"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) error
	Get(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, s *models.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID, serviceName *string, start, end *time.Time, limit, offset int) ([]models.Subscription, error)
	SumByUserAndService(ctx context.Context, userID uuid.UUID, serviceName *string, start, end *time.Time) (int, error)
}

type gormSubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &gormSubscriptionRepository{db: db}
}

func (r *gormSubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *gormSubscriptionRepository) Get(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	if err := r.db.WithContext(ctx).First(&sub, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *gormSubscriptionRepository) Update(ctx context.Context, id uuid.UUID, s *models.Subscription) error {
	result := r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("id = ?", id).
		Omit("id").
		Updates(s)

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *gormSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id).Error
}

func (r *gormSubscriptionRepository) baseQuery(ctx context.Context, userID uuid.UUID, serviceName *string, start, end *time.Time) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("user_id = ?", userID)
	if serviceName != nil {
		q = q.Where("service_name = ?", serviceName)
	}
	if start != nil {
		q = q.Where("end_date >= ? OR end_date IS NULL", start)
	}
	if end != nil {
		q = q.Where("start_date <= ?", end)
	}
	return q
}

func (r *gormSubscriptionRepository) ListByUser(ctx context.Context, userID uuid.UUID, serviceName *string, start, end *time.Time, limit, offset int) ([]models.Subscription, error) {
	var subs []models.Subscription
	if err := r.baseQuery(ctx, userID, serviceName, start, end).Limit(limit).Offset(offset).Find(&subs).Order("start, service_name").Error; err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *gormSubscriptionRepository) SumByUserAndService(ctx context.Context, userID uuid.UUID, serviceName *string, start, end *time.Time) (int, error) {
	var sum int64

	err := r.db.WithContext(ctx).Raw(`
		SELECT SUM(service_max_price) AS total_sum
		FROM (
			SELECT month, SUM(max_price) AS service_max_price
			FROM (
				SELECT month, service_name, MAX(price) AS max_price
				FROM (
					SELECT date_trunc('month', generate_series) AS month,
						service_name,
						price
					FROM (
						SELECT service_name,
							price,
							generate_series(
								date_trunc('month', start_date),
								date_trunc('month', COALESCE(end_date, CURRENT_DATE)),
								interval '1 month'
							) AS generate_series
						FROM (?) AS filtered_subs
					) AS expanded_rows
				) AS per_service_month
				WHERE month BETWEEN
					date_trunc('month', COALESCE(?, '2000-01-01'::timestamp)) AND
					date_trunc('month', COALESCE(?, CURRENT_DATE))
				GROUP BY month, service_name
			) AS max_per_service
			GROUP BY month
    	) AS month_sums
	`, r.baseQuery(ctx, userID, serviceName, start, end), start, end).Scan(&sum).Error

	return int(sum), err
}
