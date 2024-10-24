package graphql

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"

	gql "github.com/99designs/gqlgen/graphql"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgx/v5"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/thanishsid/dingilink-server/internal/types/apperror"
)

func ErrPresenter(ctx context.Context, err error) *gqlerror.Error {
	gqlErr := gql.DefaultErrorPresenter(ctx, err)

	if gqlErr.Extensions == nil {
		gqlErr.Extensions = make(map[string]interface{})
	}

	// Input errors from internal api
	var vdErrs validation.Errors

	// App Error
	var appErr *apperror.Error

	switch true {
	case errors.As(err, &vdErrs):
		gqlErr.Message = "invalid input"
		gqlErr.Extensions["code"] = "INVALID_INPUT"
		gqlErr.Extensions["inputErrors"] = vdErrs
	case errors.As(err, &appErr):
		gqlErr.Message = appErr.Message
		gqlErr.Extensions["code"] = appErr.Code
	case errors.Is(err, pgx.ErrNoRows):
		gqlErr.Message = apperror.ErrNotFound.Message
		gqlErr.Extensions["code"] = apperror.ErrNotFound.Code
		gqlErr.Extensions["internalMessage"] = err.Error()
	default:
		errStr := err.Error()

		gqlErr.Message = apperror.ErrUnexpected.Message
		gqlErr.Extensions["code"] = apperror.ErrUnexpected.Code
		gqlErr.Extensions["internalMessage"] = errStr
	}

	return gqlErr
}

// Panic recover func
func recoverFunc(ctx context.Context, panicPayload any) error {

	var gqlErr *gqlerror.Error

	var appErr *apperror.Error

	err, isErr := panicPayload.(error)
	if isErr && errors.As(err, &appErr) {
		gqlErr = gql.DefaultErrorPresenter(ctx, err)

		if gqlErr.Extensions == nil {
			gqlErr.Extensions = make(map[string]interface{})
		}

		gqlErr.Message = appErr.Message
		gqlErr.Extensions["code"] = appErr.Code

		return gqlErr
	}

	log.Printf("graphql panic recovered: %v", panicPayload)

	fmt.Printf("stacktrace from panic: \n %s \n" + string(debug.Stack()))

	errUnexpected := *apperror.ErrUnexpected

	gqlErr = gql.DefaultErrorPresenter(ctx, &errUnexpected)

	if gqlErr.Extensions == nil {
		gqlErr.Extensions = make(map[string]interface{})
	}

	gqlErr.Message = fmt.Sprintln(panicPayload)
	gqlErr.Extensions["code"] = errUnexpected.Code

	return gqlErr
}
