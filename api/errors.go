package api

import "log"

type HttpError struct {
	Message string `json:"message"`
	Status  int    `json:"-"`
	Cause   error  `json:"-"`
}

func (e HttpError) Error() string {
	if e.Cause == nil {
		return e.Message
	}

	return e.Message + " : " + e.Cause.Error()
}

type ClientError interface {
	Error() string
}

func NewHttpError(err error, status int, message string) HttpError {
	if err != nil {
		log.Println(err)
	}

	return HttpError{
		Message: message,
		Status:  status,
		Cause:   err,
	}
}
