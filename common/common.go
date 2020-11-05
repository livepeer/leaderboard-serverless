package common

import (
	"net/http"
)

func HandleInternalError(w http.ResponseWriter, err error) {
	RespondWithError(w, err, http.StatusInternalServerError)
}

func HandleBadRequest(w http.ResponseWriter, err error) {
	RespondWithError(w, err, http.StatusBadRequest)
}

func RespondWithError(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
}
