package handlers

import (
	"net/http"

	"gh-ts/internal/utils"
)

func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
