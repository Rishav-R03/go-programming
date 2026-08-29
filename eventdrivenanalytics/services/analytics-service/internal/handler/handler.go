package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"analyticsservice/internal/service"
)

type AnalyticsHandler struct {
	service *service.AnalyticsService
}

func NewAnalyticsHandler(
	service *service.AnalyticsService,
) *AnalyticsHandler {

	return &AnalyticsHandler{
		service: service,
	}
}

func (h *AnalyticsHandler) Dashboard(
	w http.ResponseWriter,
	r *http.Request,
) {

	idStr :=
		r.URL.Query().
			Get("restaurant_id")

	restaurantID, err :=
		strconv.ParseInt(
			idStr,
			10,
			64,
		)

	if err != nil {
		http.Error(
			w,
			"invalid restaurant_id",
			http.StatusBadRequest,
		)
		return
	}

	resp, err :=
		h.service.Dashboard(
			r.Context(),
			restaurantID,
		)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().
		Set(
			"Content-Type",
			"application/json",
		)

	json.NewEncoder(w).
		Encode(resp)
}
