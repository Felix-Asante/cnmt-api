package httpx

import "errors"

var (
	BadRequestError                     = errors.New("The request could not be understood or was missing required parameters.")
	InvalidJSONError                    = errors.New("The request body is not valid JSON.")
	UnauthorizedError                   = errors.New("Authentication is required to access this resource.")
	ForbiddenError                      = errors.New("You do not have permission to perform this action.")
	NotFoundError                       = errors.New("The requested resource was not found.")
	ConflictError                       = errors.New("The request conflicts with the current state of the resource.")
	UnprocessableEntityError            = errors.New("The request was well-formed but could not be processed.")
	TooManyRequestsError                = errors.New("Too many requests. Please try again later.")
	InternalServerError                 = errors.New("Something went wrong. Please try again later.")
	BadGatewayError                     = errors.New("A dependent service returned an invalid response. Please try again later.")
	ServiceUnavailableError             = errors.New("The service is temporarily unavailable. Please try again later.")
	GatewayTimeoutError                 = errors.New("A dependent service took too long to respond. Please try again later.")
	NetworkAuthenticationRequiredError  = errors.New("Network authentication is required before continuing.")
)
