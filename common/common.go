package common

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// HandleInternalError handles an internal server error for the Lambda function
func HandleLambdaInternalError(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       err.Error(),
		StatusCode: 502,
	}, err
}

// HandleBadRequest handles a bad request for the Lambda function
func HandleLambdaBadRequest(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       err.Error(),
		StatusCode: 400,
	}, err
}

func HandleInternalError(w http.ResponseWriter, err error) {
	RespondWithError(w, err, http.StatusInternalServerError)
}

func HandleBadRequest(w http.ResponseWriter, err error) {
	RespondWithError(w, err, http.StatusBadRequest)
}

func RespondWithError(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
}
