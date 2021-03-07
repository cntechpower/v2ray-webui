package handler

type statusHandler struct {
	*baseHandler
}

func newStatusHandler() *statusHandler {
	return &statusHandler{
		&baseHandler{},
	}
}
