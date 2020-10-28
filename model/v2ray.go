package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type V2rayNode struct {
	Id               int64  `json:"primary_key"`
	SubscriptionId   int64  `json:"subscription_id"`
	SubscriptionName string `json:"subscription_name"`
	Host             string `json:"host"`
	Path             string `json:"path"`
	TLS              string `json:"tls"`
	Address          string `json:"add"`
	Port             int64  `json:"port"`
	Aid              int64  `json:"aid"`
	Net              string `json:"net"`
	Type             string `json:"type"`
	V                string `json:"v"`
	Name             string `json:"ps"`
	ServerId         string `json:"id"`
	PingRTT          int64  `json:"ping_rtt"`
}

func NewV2rayNode(subscriptionId int64, subscriptionName string) *V2rayNode {
	return &V2rayNode{
		SubscriptionId:   subscriptionId,
		SubscriptionName: subscriptionName}
}

func (s *V2rayNode) GetCacheKey() string {
	return fmt.Sprintf("v2ray_node_%v", s.Id)
}

func (s *V2rayNode) GetCacheDuration() time.Duration {
	return time.Minute
}

func (s *V2rayNode) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}
func (s *V2rayNode) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *V2rayNode) TableName() string {
	return "v2ray_nodes"
}

type Subscription struct {
	Id   int64  `json:"id"`
	Name string `json:"subscription_name"`
	Addr string `json:"subscription_addr"`
}

func NewSubscription(name, addr string) *Subscription {
	return &Subscription{
		Name: name,
		Addr: addr,
	}
}

func (s *Subscription) GetCacheKey() string {
	return fmt.Sprintf("v2ray_sub_%v", s.Id)
}

func (s *Subscription) GetCacheDuration() time.Duration {
	return time.Minute
}

func (s *Subscription) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}
func (s *Subscription) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *Subscription) TableName() string {
	return "v2ray_subscriptions"
}
