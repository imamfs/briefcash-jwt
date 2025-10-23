package controller

import (
	jsonHelper "briefcash-jwt/internal/helper/jsonhelper"
	logger "briefcash-jwt/internal/helper/loghelper"
	service "briefcash-jwt/internal/service"
	"encoding/json"
	"net/http"
	"time"
)

type MerchantController struct {
	svc service.MerchantService
}

func NewMerchantController(svc service.MerchantService) *MerchantController {
	return &MerchantController{svc: svc}
}

func (s *MerchantController) SyncMerchantCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	logs := logger.Logger.WithFields(map[string]interface{}{
		"endpoint": "/api/v1/merchant/sync",
		"method":   r.Method,
		"trace_id": r.Header.Get("X-Request-ID"),
	})

	logs.Info("Start syncing merchant code from db to redis")

	if err := s.svc.LoadActiveMerchantCodeToRedis(ctx); err != nil {
		logs.WithFields(map[string]interface{}{
			"duration_ms": time.Since(start).Milliseconds(),
			"error":       err.Error(),
		}).Error("Failed to sync merchant code")
		jsonHelper.WriteJsonError(w, http.StatusInternalServerError, "Failed to sync merchant code list")
		return
	}

	logs.WithFields(map[string]interface{}{
		"duration_ms": time.Since(start).Milliseconds(),
	}).Info("Merchant code sync completed successfully")

	jsonHelper.WriteJson(w, http.StatusOK, "Succesfully synced merchant code")
}

func (s *MerchantController) AddMerchantCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	decoder := json.NewDecoder(r.Body)

	start := time.Now()

	logs := logger.Logger.WithFields(map[string]interface{}{
		"endpoint": "/api/v1/merchant/add",
		"method":   r.Method,
		"trace_id": r.Header.Get("X-Request-ID"),
	})

	var request struct {
		MerchantCode string `json:"merchant_code"`
	}

	if err := decoder.Decode(&request); err != nil {
		logs.WithFields(map[string]interface{}{
			"duration_ms": time.Since(start).Milliseconds(),
			"error":       err.Error(),
		}).Warn("Invalid body request while adding merchant code")
		jsonHelper.WriteJsonError(w, http.StatusBadRequest, "Invalid body request")
		return
	}

	if request.MerchantCode == "" {
		logs.WithFields(map[string]interface{}{
			"duration_ms": time.Since(start).Milliseconds(),
		}).Warn("Merchant code is empty")
		jsonHelper.WriteJsonError(w, http.StatusBadRequest, "merchant_code is required")
		return
	}

	logs.WithFields(map[string]interface{}{
		"merchant_code": request.MerchantCode,
	}).Info("Adding merchant code to redis")

	if err := s.svc.AddMerchantCode(ctx, request.MerchantCode); err != nil {
		logs.WithFields(map[string]interface{}{
			"duration_ms": time.Since(start).Milliseconds(),
			"error":       err.Error(),
		}).Error("Failed to add merchant code to redis")
		jsonHelper.WriteJsonError(w, http.StatusInternalServerError, "Failed to add merchant code")
		return
	}

	logs.WithFields(map[string]interface{}{
		"duration_ms": time.Since(start).Milliseconds(),
	}).Info("Successfully added merchant code to redis")

	jsonHelper.WriteJson(w, http.StatusOK, "Merchant code successfully added to redis")
}

func (s *MerchantController) RemoveMerchantCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	decoder := json.NewDecoder(r.Body)
	start := time.Now()

	logs := logger.Logger.WithFields(map[string]interface{}{
		"endpoint": "/api/v1/merchant/remove",
		"method":   r.Method,
		"trace_id": r.Header.Get("X-Request-ID"),
	})

	var request struct {
		MerchantCode string `json:"merchant_code"`
	}

	if err := decoder.Decode(&request); err != nil {
		logs.WithFields(map[string]interface{}{
			"duration_ms": time.Since(start).Milliseconds(),
		}).Warn("Invalid request body while removing merchant code")
		jsonHelper.WriteJsonError(w, http.StatusBadRequest, "Invalid body request")
		return
	}

	if request.MerchantCode == "" {
		logs.WithFields(map[string]interface{}{
			"duration_ms": time.Since(start).Milliseconds(),
		}).Warn("Missing merchant code field in request body")
		jsonHelper.WriteJsonError(w, http.StatusBadRequest, "merchant_code is required")
		return
	}

	logs.WithFields(map[string]interface{}{
		"merchant_code": request.MerchantCode,
	}).Info("Removing merchant code from redis")

	if err := s.svc.RemoveMerchantCode(ctx, request.MerchantCode); err != nil {
		logs.WithFields(map[string]interface{}{
			"duration_ms": time.Since(start).Milliseconds(),
			"error":       err.Error(),
		}).Error("Failed to remove merchant code from redis")
		jsonHelper.WriteJsonError(w, http.StatusInternalServerError, "Failed to remove merchant code")
		return
	}

	logs.WithFields(map[string]interface{}{
		"duration_ms": time.Since(start).Milliseconds(),
	}).Info("Successfully removed merchant code from redis")

	jsonHelper.WriteJson(w, http.StatusOK, "Merchant code successfully removed from redis")
}
