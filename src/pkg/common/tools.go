package common

import (
	"apigw/src/pkg/answer"
	"apigw/src/pkg/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

func CreateTimestamp() int64 {
	t := time.Now().UnixNano() / 1e6
	return t
}

const (
	defaultLimit  = 10  // default page size
	maxLimit      = 100 // max allowed page size
	defaultOffset = 0   // default offset value
)

// GetPagingQuery extracts and validates pagination parameters from HTTP request.
// It handles both 'limit' (page size) and 'offset' (starting position) parameters
// with proper validation and default values.
//
// Parameters:
//   - c: The request context containing HTTP request information
//
// Returns:
//   - limit: Number of items to return (validated between 1 and maxLimit)
//   - offset: Starting position (must be >= 0)
//   - err: Non-nil if validation fails (caller should return immediately if err != nil)
//
// Error Handling:
//   - Returns error and sends HTTP 400 response if parameters are invalid
//   - Error message is suitable for both client consumption and logging
func GetPagingQuery(c *app.RequestContext) (limit, offset int, err error) {

	// Validate and parse limit parameter
	limit, err = utils.Str2Int(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil {
		return handleInvalidParam(c, "limit must be a valid integer")
	}
	if limit < 1 || limit > maxLimit {
		return handleInvalidParam(c, fmt.Sprintf("limit must be between 1 and %d", maxLimit))
	}

	// Validate and parse offset parameter
	offset, err = utils.Str2Int(c.DefaultQuery("offset", strconv.Itoa(defaultOffset)))
	if err != nil {
		return handleInvalidParam(c, "offset must be a valid integer")
	}
	if offset < 0 {
		return handleInvalidParam(c, "offset must be greater than or equal to 0")
	}

	return limit, offset, nil
}

// handleInvalidParam sends a standardized error response for invalid parameters
// and returns the error tuple (0, 0, error) to simplify caller error handling.
//
// Parameters:
//   - c: The request context for sending the error response
//   - message: Human-readable error message describing the validation failure
//
// Returns:
//   - Always returns (0, 0, error) to maintain consistent function signature
func handleInvalidParam(c *app.RequestContext, message string) (int, int, error) {
	c.JSON(http.StatusBadRequest, answer.ResBody(answer.EcodeInvalidRequestErr, message, ""))
	return 0, 0, errors.New(message)
}
