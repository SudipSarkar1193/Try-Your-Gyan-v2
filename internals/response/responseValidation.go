package response

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ValidateResponse(w http.ResponseWriter, errs interface{}) error {
	for _, err := range errs.(validator.ValidationErrors) {

		var errMsgs []string

		switch err.ActualTag() {
		case "email":
			errMsgs = append(errMsgs, "invalid email")
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("%s is required\n", err.Field()))

		default:
			errMsgs = append(errMsgs, "Invalid Input !")
		}

		

		http.Error(w, strings.Join(errMsgs, ","), http.StatusInternalServerError)

	}

	return nil
}
