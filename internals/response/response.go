package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Data             interface{} `json:"data,omitempty"`
	Message          string      `json:"message"`
	StatusCode       int         `json:"statusCode"`
	IsError          bool        `json:"isError"`
	ErrorMessage     string      `json:"errorMessage"`
	DeveloperMessage string      `json:"developerMessage,omitempty"`
	UserMessage      string      `json:"userMessage,omitempty"`
}

func CreateResponse(data interface{}, statusCode int, msg string, options ...interface{}) Response {
	response := Response{
		Data:       data,
		Message:    msg,
		StatusCode: statusCode,
	}

	if len(options) > 0 {
		for _, opt := range options {
			switch v := opt.(type) {
			case bool: // isError flag
				response.IsError = v
			case string: // ErrorMessage, DeveloperMessage, or UserMessage
				// Here, you can apply more logic to decide which string this is, if needed
				if response.IsError {
					response.ErrorMessage = v
				} else if response.DeveloperMessage == "" {
					response.DeveloperMessage = v
				} else {
					response.UserMessage = v
				}
			}
		}
	}

	return response
}

func WriteResponse(w http.ResponseWriter, res interface{}) error {
	// Check if res is of type Response
	r, ok := res.(Response)
	if !ok {
		// Handle the case where res is not a Response type
		http.Error(w, "invalid response type", http.StatusInternalServerError)
		return fmt.Errorf("invalid response type, expected Response, got %T", res)
	} else {
		// Now it's safe to access fields from r
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(r.StatusCode)

		return json.NewEncoder(w).Encode(r)
	}

}
