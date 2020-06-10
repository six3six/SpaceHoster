package main

import "fmt"

type Response struct {
	Code  int
	Error ApiError
	Data  interface{}
}

type ApiError struct {
	Code    int
	Context string
	Message string
}

var LOGIN_NOT_FOUND = ApiError{0, "Login", "Unable to find specified login"}
var CAN_NOT_READ_PASSWORD = ApiError{1, "Login", "Unable to read database encoded password"}
var BAD_PASSWORD = ApiError{2, "Login", "Bad password"}
var CAN_NOT_CONNECT_DATABASE = ApiError{3, "Global", "Can't connect to database"}
var FAILED_TO_CONNECT_SESSION = ApiError{4, "Login", "Failed to save session"}
var LOGIN_ALREADY_REGISTERED = ApiError{5, "Register", "Login already registered"}
var PASSWORD_HASHING_FAILED = ApiError{6, "Register", "Password hashing failed"}
var USER_REGISTRATION_FAILED = ApiError{7, "Register", "User registration failed"}
var NOT_AUTHENTICATED = ApiError{8, "Global", "Not authenticated"}

func (err *ApiError) Error() string {
	return fmt.Sprintf("Context : %s Message : %s", err.Context, err.Message)
}
