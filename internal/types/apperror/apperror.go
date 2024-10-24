package apperror

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	ErrUnexpected = NewError("UNEXPECTED_ERROR", "unexpected error", http.StatusInternalServerError)
	ErrNotFound   = NewError("NOT_FOUND", "item not found", http.StatusNotFound)

	// Auth Errors
	ErrTokenExpired       = NewError("TOKEN_EXPIRED", "token expired", http.StatusUnauthorized)
	ErrForbidden          = NewError("FORBIDDEN", "insufficient permissions to perform this action", http.StatusForbidden)
	ErrInvalidCredentials = NewError("INVALID_CREDENTIALS", "the email or password you have provided is incorrect")

	// App Errors
	ErrEmailUnverified               = NewError("EMAIL_UNVERIFIED", "your account email must be verified", http.StatusBadRequest)
	ErrInvalidEmailVerificationToken = NewError("INVALID_EMAIL_VERIFICATION_TOKEN", "your email verification link is no longer valid", http.StatusBadRequest)
	ErrEmailAlreadyVerified          = NewError("EMAIL_ALREADY_VERIFIED", "your email has already been verified", http.StatusBadRequest)
	ErrAccountNotFound               = NewError("ACCOUNT_NOT_FOUND", "unable to find your account, please make sure you have registered", http.StatusNotFound)
)

func NewError(code string, msg string, httpCode ...int) *Error {
	err := &Error{
		Code:    code,
		Message: msg,
	}

	if len(httpCode) > 0 {
		err.HttpCode = httpCode[0]
	}

	return err
}

func NewInputValidationError(err error) *Error {
	appErr := &Error{
		HttpCode: http.StatusBadRequest,
		Code:     "INVALID_INPUT",
		Message:  "invalid input",
	}

	var vdErrs validation.Errors

	if errors.As(err, &vdErrs) {
		appErr.InputErrors = vdErrs
	}

	return appErr
}

type Error struct {
	HttpCode        int    `json:"-"`
	Code            string `json:"code"`
	Message         string `json:"message"`
	InternalMessage string `json:"internalMessage,omitempty"`

	// when error is of invalid input the following will have the error fields (USED ONLY FOR REST API)
	InputErrors validation.Errors `json:"inputErrors,omitempty"`
}

func (e Error) ErrorCode() string {
	return e.Code
}

// Implementing the error interface
func (e *Error) Error() string {
	return e.Message
}

func (e *Error) SetInternalMessage(msg ...any) *Error {
	e.InternalMessage = fmt.Sprintln(msg...)
	return e
}

func (e *Error) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HttpCode)
	return nil
}

// Input Error Codes
const (
	INPUT_REQUIRED     = "REQUIRED"
	INPUT_NOT_REQUIRED = "NOT_REQUIRED"
	INPUT_INVALID      = "INVALID"
	INPUT_DUPLICATE    = "DUPLICATE"
	INPUT_TOO_LOW      = "TOO_LOW"
	INPUT_TOO_HIGH     = "TOO_HIGH"
)
