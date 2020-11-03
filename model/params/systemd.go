package params

type SystemdServiceName struct {
	SystemdServiceName string `form:"service_name" binding:"required"`
}
