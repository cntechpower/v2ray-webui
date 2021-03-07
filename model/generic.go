package model

type GenericStatus struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty" `
	Data    interface{} `json:"data,omitempty"`
}

func NewGenericStatus(code int, message string) *GenericStatus {
	return &GenericStatus{
		Code:    code,
		Message: message,
	}
}

func NewGenericData(code int, data interface{}) *GenericStatus {
	return &GenericStatus{
		Code: code,
		Data: data,
	}
}
