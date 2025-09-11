package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"subscriptions-service/internal/logger"
	"time"

	"github.com/google/uuid"
)

type MonthYearDate time.Time

func (myd *MonthYearDate) UnmarshalJSON(b []byte) error {
	var dateStr string
	if err := json.Unmarshal(b, &dateStr); err != nil {
		logger.Log.Error("Failed to unmarshal date string", "error", err)
		return err
	}
	t, err := time.Parse("01-2006", dateStr)
	*myd = MonthYearDate(t)
	return err
}

func (myd MonthYearDate) MarshalJSON() ([]byte, error) {
	t := time.Time(myd)
	return json.Marshal(t.Format("01-2006"))
}

func (myd *MonthYearDate) Scan(value interface{}) error {
	if t, ok := value.(time.Time); ok {
		*myd = MonthYearDate(t)
		return nil
	}
	return fmt.Errorf("failed to scan MonthYearDate")
}

func (myd MonthYearDate) Value() (driver.Value, error) {
	return time.Time(myd), nil
}

type Subscription struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ServiceName string         `gorm:"not null" json:"service_name"`
	Price       int            `gorm:"not null" json:"price"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	StartDate   MonthYearDate  `gorm:"not null" json:"start_date"`
	EndDate     *MonthYearDate `json:"end_date"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}
