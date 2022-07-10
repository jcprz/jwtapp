package controllers

import (
	"net/http"

	"github.com/jcprz/jwtapp/utils"
)

func (c Controller) HealthZ() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.ResponseJSON(w, http.StatusOK, "Ok")
	}
}
