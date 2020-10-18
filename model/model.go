package model

import "encoding"

type Modeler interface {
	GetCacheKey() string
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

func GetAllModels() []interface{} {
	return []interface{}{
		&ProxyWebSite{},
	}
}
