package main

type ApiError struct {
	code    int
	message string
}

var LOGIN_NOT_FOUND = ApiError{0, "Unable to find specified login"}
var CAN_NOT_READ_PASSWORD = ApiError{1, "Unable to read database encoded password"}
var BAD_PASSWORD = ApiError{2, "Bad password"}
var CAN_NOT_CONNECT_DATABASE = ApiError{3, "Can't connect to database"}
var FAILED_TO_CONNECT_SESSION = ApiError{4, "Failed to save session"}
var LOGIN_ALREADY_REGISTERED = ApiError{5, "Login already registered"}
var PASSWORD_HASHING_FAILED = ApiError{6, "Password hashing failed"}
var USER_REGISTRATION_FAILED = ApiError{7, "User registration failed"}
var NOT_AUTHENTICATED = ApiError{8, "Not authenticated"}

func (err *ApiError) toArray() map[string]interface{} {
	return map[string]interface{}{"code": err.code, "message": err.message}
}
