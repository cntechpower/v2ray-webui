package model

type GenericStatus struct {
	Code    int
	Message string
}

func NewGenericStatus(code int, message string) *GenericStatus {
	return &GenericStatus{
		Code:    code,
		Message: message,
	}
}
