package controllers

import (
	"net/http"

	"github.com/jcprz/jwtapp/utils"
)

type Controller struct{}

func (c Controller) ProtectedEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.ResponseJSON(w, http.StatusOK, "Yes")
	}
}
