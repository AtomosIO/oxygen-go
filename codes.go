package oxygen

type OxygenResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

const (
	SUCCESS            = iota + 1000000
	ERROR_PARSING_JSON = iota + 1
	ERROR_INVALID_METHOD
	ERROR_INVALID_VERSION
	ERROR_MISSING_ARGUMENTS
	ERROR_INVALID_PROJECTNAME
	ERROR_PROJECT_ALREADY_EXISTS
	ERROR_INVALID_TOKEN
	ERROR_INVALID_RANGE
	ERROR_INVALID_CONTENT_LENGTH
	ERROR_INVALID_ID_PARAMETER
	ERROR_REQUIRE_TOKEN
	ERROR_MAX_PROJECTS_REACHED
	ERROR_INVALID_CREDENTIALS
	ERROR_NEED_EMAIL_PASSWORD_OR_TOKEN
	ERROR_INVALID_EMAIL
	ERROR_INVALID_PASSWORD
	ERROR_INVALID_EXPIRES
	ERROR_INVALID_ARGUMENT_TYPE
	ERROR_TOKEN_EMPTY
	ERROR_PROJECT_DOES_NOT_EXIST
	ERROR_INTERNAL_ERROR
	ERROR_PATH_NOT_FOUND
	ERROR_PATH_NOT_FOUND_SOURCE
	ERROR_PATH_NOT_FOUND_DESTINATION
	ERROR_INVALID_PATH
	ERROR_INVALID_SOURCE_PATH
	ERROR_INVALID_DESTINATIO_PATH
	ERROR_DIRECTORY_ALREADY_EXISTS
	ERROR_INVALID_USERNAME
	ERROR_USERNAME_ALREADY_EXISTS
	ERROR_EMAIL_ALREADY_IN_USE
	ERROR_NO_WRITE_PERMISSION
	ERROR_NOT_A_DIRECTORY
	ERROR_DIRECTORY_NOT_EMPTY
	ERROR_PATH_ALREADY_EXISTS
	ERROR_USER_CANNOT_TAKE_PROJECT
	ERROR_PROJECT_ALREADY_SHARED_WITH_USER
	ERROR_INVALID_PERMISSIONS
	ERROR_INSUFFICIENT_PERMISSIONS
	ERROR_INSUFFICIENT_PERMISSIONS_SOURCE
	ERROR_INSUFFICIENT_PERMISSIONS_DESTINATION
	ERROR_405
	ERROR_OUT_OF_RANGE
	NO_RESPONSE_REQUIRED
)