package params

type V2rayConfig struct {
	ConfigContent string `form:"config_content" binding:"required"`
}

type V2raySubscriptionIdParam struct {
	SubscriptionId int64 `form:"subscription_id" binding:"required"`
}

type UpdateV2raySubscriptionParam struct {
	SubscriptionId   int64  `form:"subscription_id" binding:"required"`
	SubscriptionAddr string `form:"subscription_addr" binding:"required"`
	SubscriptionName string `form:"subscription_name" binding:"required"`
}

type AddV2raySubscriptionParam struct {
	SubscriptionAddr string `form:"subscription_addr" binding:"required"`
	SubscriptionName string `form:"subscription_name" binding:"required"`
}

type V2raySwitchNodeParam struct {
	NodeId int64 `form:"node_id" binding:"required"`
}
