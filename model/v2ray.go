package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type V2rayNode struct {
	Id               int64  `json:"primary_key"`
	SubscriptionId   int64  `json:"subscription_id"`
	SubscriptionName string `json:"subscription_name"`
	SubscriptionType string `json:"subscription_type"`

	//props for v2ray start
	Host     string     `json:"host"`
	Path     string     `json:"path"`
	TLS      string     `json:"tls"`
	Address  string     `json:"add"`
	Port     FlexString `json:"port"`
	Aid      FlexString `json:"aid"`
	Net      string     `json:"net"`
	Type     string     `json:"type"`
	V        string     `json:"v"`
	Name     string     `json:"ps"`
	ServerId string     `json:"id"`
	PingRTT  int64      `json:"ping_rtt"`
	//props for v2ray end

	//props for trojan start
	Password string `json:"password"`
	Sni      string `json:"sni"`
	//props for trojan end
}

// A FlexString is an string that can be unmarshalled from a JSON field
// that has either a number or a string value.
// E.g. if the json field contains an string "42", the
// FlexString value will be "42".
type FlexString string

func (fi *FlexString) UnmarshalJSON(b []byte) error {
	if b[0] == '"' { //start with ", it is already a string
		return json.Unmarshal(b, (*string)(fi))
	}
	var i int64
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	*fi = FlexString(strconv.FormatInt(i, 10))
	return nil
}

const (
	SubscriptionTypeVmess  = "vmess"
	SubscriptionTypeTrojan = "trojan"
)

func NewV2rayNode(subscriptionId int64, subscriptionName, subscriptionType string) *V2rayNode {
	return &V2rayNode{
		SubscriptionId:   subscriptionId,
		SubscriptionName: subscriptionName,
		SubscriptionType: subscriptionType,
	}
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

func (s *V2rayNode) DeepCopy() *V2rayNode {
	return &V2rayNode{
		Id:               s.Id,
		SubscriptionId:   s.SubscriptionId,
		SubscriptionName: s.SubscriptionName,
		Host:             s.Host,
		Path:             s.Path,
		TLS:              s.TLS,
		Address:          s.Address,
		Port:             s.Port,
		Aid:              s.Aid,
		Net:              s.Net,
		Type:             s.Type,
		V:                s.V,
		Name:             s.Name,
		ServerId:         s.ServerId,
		PingRTT:          s.PingRTT,
	}
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

type V2rayCoreStatus struct {
	StartTime string `json:"start_time"`
}

type Traffic struct {
	Tag     string `json:"tag"`
	Traffic int64  `json:"traffic"`
}
type V2rayStatus struct {
	CurrentNode     *V2rayNode       `json:"current_node"`
	Core            *V2rayCoreStatus `json:"v2ray_core"`
	RefreshTime     string           `json:"refresh_time"`
	InboundTraffic  []Traffic        `json:"inbound_traffic"`
	OutboundTraffic []Traffic        `json:"outbound_traffic"`
}
