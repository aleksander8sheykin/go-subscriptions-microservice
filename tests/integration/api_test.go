package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"subscriptions-service/internal/config"
	"subscriptions-service/internal/database"
	"subscriptions-service/internal/handlers"
	_ "subscriptions-service/internal/logger"
	"subscriptions-service/internal/models"
	"subscriptions-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	r  *gin.Engine
	db *gorm.DB
)

func TestMain(m *testing.M) {
	var err error

	cfg := config.LoadConfig()
	db, err = database.Connect(cfg)
	if err != nil {
		panic("Ошибка подключения к БД" + err.Error())
	}

	_ = db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")

	err = db.AutoMigrate(&models.Subscription{})
	if err != nil {
		panic("не удалось выполнить миграцию: " + err.Error())
	}

	repo := repository.NewSubscriptionRepository(db)
	h := handlers.NewHandler(repo)
	r = gin.Default()
	h.RegisterRoutes(r)

	code := m.Run()

	os.Exit(code)
}

func clearDB(db *gorm.DB) error {
	return db.Exec("DELETE FROM subscriptions").Error
}

func TestCreateSubscription(t *testing.T) {
	clearDB(db)

	body := models.Subscription{
		ServiceName: "Netflix",
		Price:       499,
		UserID:      uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		StartDate:   models.MonthYearDate(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
		EndDate:     nil,
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/subscriptions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)

	var created models.Subscription
	err := json.Unmarshal(resp.Body.Bytes(), &created)
	assert.NoError(t, err)
	assert.Equal(t, body.ServiceName, created.ServiceName)
	assert.Equal(t, body.Price, created.Price)
	assert.Equal(t, body.UserID, created.UserID)
}

func TestGetSubscription(t *testing.T) {
	clearDB(db)
	sub := models.Subscription{
		ID:          uuid.New(),
		ServiceName: "Spotify",
		Price:       299,
		UserID:      uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		StartDate:   models.MonthYearDate(time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)),
	}
	err := db.Create(&sub).Error
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/subscriptions/"+sub.ID.String(), nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var fetched models.Subscription
	err = json.Unmarshal(resp.Body.Bytes(), &fetched)
	assert.NoError(t, err)
	assert.Equal(t, sub.ID, fetched.ID)
	assert.Equal(t, sub.ServiceName, fetched.ServiceName)
}

func TestUpdateSubscription(t *testing.T) {
	clearDB(db)
	sub := models.Subscription{
		ID:          uuid.New(),
		ServiceName: "YouTube Premium",
		Price:       179,
		UserID:      uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		StartDate:   models.MonthYearDate(time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)),
	}
	err := db.Create(&sub).Error
	assert.NoError(t, err)

	updated := models.Subscription{
		ServiceName: "YouTube Premium",
		Price:       249,
		UserID:      sub.UserID,
		StartDate:   models.MonthYearDate(time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)),
	}

	jsonBody, _ := json.Marshal(updated)
	req, _ := http.NewRequest("PUT", "/subscriptions/"+sub.ID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestDeleteSubscription(t *testing.T) {
	clearDB(db)
	sub := models.Subscription{
		ID:          uuid.New(),
		ServiceName: "ToBeDeleted",
		Price:       123,
		UserID:      uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		StartDate:   models.MonthYearDate(time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)),
	}

	err := db.Create(&sub).Error
	assert.NoError(t, err)

	req, _ := http.NewRequest("DELETE", "/subscriptions/"+sub.ID.String(), nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	getReq, _ := http.NewRequest("GET", "/subscriptions/"+sub.ID.String(), nil)
	getResp := httptest.NewRecorder()
	r.ServeHTTP(getResp, getReq)
	assert.Equal(t, http.StatusNotFound, getResp.Code)
}

func TestGetSubscriptionList(t *testing.T) {
	clearDB(db)

	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	start := models.MonthYearDate(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	end := models.MonthYearDate(time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC))

	db.Create(&models.Subscription{
		ServiceName: "Netflix",
		Price:       500,
		UserID:      userID,
		StartDate:   start,
	})
	db.Create(&models.Subscription{
		ServiceName: "Spotify",
		Price:       300,
		UserID:      userID,
		StartDate:   start,
		EndDate:     &end,
	})

	url := "/subscriptions/list?user_id=" + userID.String() +
		"&start_date=01-2025&end_date=12-2025"

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var subs []models.Subscription
	err := json.Unmarshal(resp.Body.Bytes(), &subs)
	assert.NoError(t, err)
	assert.Len(t, subs, 2)

	// one service
	url = "/subscriptions/list?user_id=" + userID.String() +
		"&service_name=Netflix&start_date=01-2025&end_date=12-2025"

	req, _ = http.NewRequest("GET", url, nil)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	err = json.Unmarshal(resp.Body.Bytes(), &subs)
	assert.NoError(t, err)
	assert.Len(t, subs, 1)

	// limit
	url = "/subscriptions/list?user_id=" + userID.String() +
		"&start_date=01-2025&end_date=12-2025&limit=1"

	req, _ = http.NewRequest("GET", url, nil)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	err = json.Unmarshal(resp.Body.Bytes(), &subs)
	assert.NoError(t, err)
	assert.Equal(t, "Netflix", subs[0].ServiceName)

	url = "/subscriptions/list?user_id=" + userID.String() +
		"&start_date=01-2025&end_date=12-2025&limit=1&offset=1"

	req, _ = http.NewRequest("GET", url, nil)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	err = json.Unmarshal(resp.Body.Bytes(), &subs)
	assert.NoError(t, err)
	assert.Equal(t, "Spotify", subs[0].ServiceName)
}

func TestGetSubscriptionSum(t *testing.T) {
	clearDB(db)

	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	start := models.MonthYearDate(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	end := models.MonthYearDate(time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC))

	db.Create(&models.Subscription{
		ServiceName: "Netflix",
		Price:       500,
		UserID:      userID,
		StartDate:   start,
		EndDate:     &end,
	})

	db.Create(&models.Subscription{
		ServiceName: "Spotify",
		Price:       200,
		UserID:      userID,
		StartDate:   start,
		EndDate:     &end,
	})

	url := "/subscriptions/sum?user_id=" + userID.String() +
		"&start_date=02-2025&end_date=07-2025"

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var result map[string]int
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, (500+200)*6, result["sum"])
}

func TestGetSubscriptionSumWithSameService(t *testing.T) {
	clearDB(db)

	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	start := models.MonthYearDate(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	end := models.MonthYearDate(time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC))
	start_in := models.MonthYearDate(time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))
	end_in := models.MonthYearDate(time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC))

	db.Create(&models.Subscription{
		ServiceName: "Spotify",
		Price:       200,
		UserID:      userID,
		StartDate:   start,
		EndDate:     &end,
	})

	db.Create(&models.Subscription{
		ServiceName: "Spotify",
		Price:       100,
		UserID:      userID,
		StartDate:   start_in,
		EndDate:     &end_in,
	})

	url := "/subscriptions/sum?user_id=" + userID.String() +
		"&start_date=02-2025&end_date=07-2025"

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var result map[string]int
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, (200)*6, result["sum"])
}

func TestGetSubscriptionSumWithEmptyEnd(t *testing.T) {
	clearDB(db)

	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	start := models.MonthYearDate(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	db.Create(&models.Subscription{
		ServiceName: "Spotify",
		Price:       200,
		UserID:      userID,
		StartDate:   start,
	})

	url := "/subscriptions/sum?user_id=" + userID.String() +
		"&start_date=01-2025&end_date=09-2025"

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var result map[string]int
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, (200)*9, result["sum"])
}
