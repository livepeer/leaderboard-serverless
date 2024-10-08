package common

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func HandleInternalError(w http.ResponseWriter, err error) {
	RespondWithError(w, err, http.StatusInternalServerError)
}

func HandleBadRequest(w http.ResponseWriter, err error) {
	RespondWithError(w, err, http.StatusBadRequest)
}

func RespondWithError(w http.ResponseWriter, err error, code int) {
	Logger.Warn("An error occured while handling the user request: %v", err.Error())
	http.Error(w, fmt.Sprintf("{\"error\":\"%s\"}", err.Error()), code)
}

// EnvOrDefault returns the value of the environment variable if set, otherwise returns the default value.
// It supports default values of type string and int.
func EnvOrDefault(envVar string, defaultValue interface{}) interface{} {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}

	switch defaultValue.(type) {
	case string:
		return value
	case int:
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	// If the type is not supported or conversion fails, return the default value
	return defaultValue
}
