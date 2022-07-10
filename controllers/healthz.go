package controllers

import (
	"jwt-app/utils"
	"net/http"
)

func (c Controller) HealthZ() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.ResponseJSON(w, http.StatusOK, "Ok")
	}
}
