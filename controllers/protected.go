package controllers

import (
	"jwt-app/utils"
	"net/http"
)

type Controller struct{}

func (c Controller) ProtectedEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.ResponseJSON(w, http.StatusOK, "Yes")
	}
}
