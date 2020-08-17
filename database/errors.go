package database

import "errors"

var NotFound error = errors.New("Not Found")
var InvalidRequest error = errors.New("Invalid Request")
var InternalServerError error = errors.New("Internal Server Error")
var RoomPresent error = errors.New("Room Present")
var NotAllowed error = errors.New("Not Allowed")
