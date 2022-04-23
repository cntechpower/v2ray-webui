module github.com/cntechpower/v2ray-webui

go 1.14

require (
	github.com/cntechpower/utils v0.0.0-20210114162751-4711c3f01d0b
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/contrib v0.0.0-20201101042839-6a891bf89f19
	github.com/gin-gonic/gin v1.6.3
	github.com/go-ping/ping v0.0.0-20201022122018-3977ed72668a
	github.com/go-playground/assert/v2 v2.0.1
	github.com/go-playground/validator/v10 v10.4.0
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/robfig/cron/v3 v3.0.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/ugorji/go v1.1.12 // indirect
	github.com/v2fly/v2ray-core/v4 v4.44.0
	go.uber.org/atomic v1.5.0
	google.golang.org/grpc v1.41.0
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/sqlite v1.1.3
	gorm.io/gorm v1.20.2
	v2ray.com/core v4.19.1+incompatible
)

replace v2ray.com/core => github.com/v2fly/v2ray-core/v4 v4.44.0
