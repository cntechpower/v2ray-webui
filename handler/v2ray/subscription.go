package v2ray

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cntechpower/utils/log"
	"github.com/cntechpower/v2ray-webui/model"
	"github.com/cntechpower/v2ray-webui/persist"
)

func (h *Handler) AddSubscription(subscriptionName, subscriptionAddr string) error {
	return persist.Save(model.NewSubscription(subscriptionName, subscriptionAddr))
}

func (h *Handler) DelSubscription(subscriptionId int64) error {
	return persist.Delete(&model.Subscription{Id: subscriptionId})

}
func (h *Handler) GetAllSubscriptions() ([]*model.Subscription, error) {
	res := make([]*model.Subscription, 0)
	return res, persist.DB.Find(&res).Error
}

func (h *Handler) EditSubscription(subscriptionId int64, subscriptionName, subscriptionAddr string) error {
	subscriptionConfig := model.NewSubscription(subscriptionName, subscriptionAddr)
	subscriptionConfig.Id = subscriptionId
	return persist.Save(subscriptionConfig)

}

func (h *Handler) RefreshV2raySubscription(subscriptionId int64) error {
	subscriptionConfig := &model.Subscription{Id: subscriptionId}
	if err := persist.Get(subscriptionConfig); err != nil {
		return err
	}
	subscriptionName := subscriptionConfig.Name
	subscriptionUrl := subscriptionConfig.Addr
	if h.v2raySubscriptionRefreshing.Load() {
		return fmt.Errorf("refreshing is already doing")
	}
	h.v2raySubscriptionRefreshing.Store(true)
	defer h.v2raySubscriptionRefreshing.Store(false)
	header := log.NewHeader("RefreshV2raySubscription")
	log.Infof(header, "starting fetch subscription %v: %v", subscriptionName, subscriptionUrl)
	resp, err := http.Get(subscriptionUrl)
	if err != nil {
		return err
	}
	log.Infof(header, "fetch %v response code: %v, status: %v, content length: %v", subscriptionUrl, resp.StatusCode, resp.Status, resp.ContentLength)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http request fail")
	}
	res := make([]*model.V2rayNode, 0)
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	res, err = h.decodeSubscription(subscriptionId, subscriptionName, bs)
	if err != nil {
		return err
	}
	if err := persist.DB.Exec("delete from v2ray_nodes where subscription_id =?", subscriptionId).Error; err != nil {
		log.Errorf(header, "truncate table v2ray_nodes fail: %v", err)
		return err
	}
	for _, server := range res {
		if err := persist.DB.Create(&server).Error; err != nil {
			log.Errorf(header, "save v2ray server to db error: %v", err)
		}
	}
	return nil
}

func (h *Handler) decodeSubscriptionLine(header *log.Header, subscriptionId int64, subscriptionName, line string) (node *model.V2rayNode, err error) {
	if strings.HasPrefix(line, "vmess://") {
		s := strings.TrimPrefix(line, "vmess://")
		var bs []byte
		bs, err = tryDecode(s)
		if err != nil {
			header.Errorf("some line decode fail: %v", err)
			return
		}
		if len(bs) == 0 {
			err = fmt.Errorf("some line decode fail: bs nil")
			header.Errorf("%v", err)
			return
		}
		node = model.NewV2rayNode(subscriptionId, subscriptionName, model.SubscriptionTypeVmess)
		if err = json.Unmarshal(bs, &node); err != nil {
			log.Errorf(header, "unmarshal %v error: %v", string(bs), err)
			return
		}
	} else if strings.HasPrefix(line, "trojan://") {
		var password, host, sni, name, portS string
		var port int
		s := strings.TrimPrefix(line, "trojan://")
		ss1 := strings.Split(s, "@")
		password, elseWithoutPass := ss1[0], ss1[1]
		ss2 := strings.Split(elseWithoutPass, "?")
		hostPort, elseWithoutHostPort := ss2[0], ss2[1]
		hostPortSlice := strings.Split(hostPort, ":")
		host, portS = hostPortSlice[0], hostPortSlice[1]
		port, err = strconv.Atoi(portS)
		if err != nil {
			return
		}
		ss3 := strings.Split(elseWithoutHostPort, "#")
		sni = strings.TrimPrefix(ss3[0], "sni=")
		name, err = url.QueryUnescape(ss3[1])
		if err != nil {
			return
		}

		node = model.NewV2rayNode(subscriptionId, subscriptionName, model.SubscriptionTypeTrojan)
		node.Host = host
		node.Password = password
		node.Port = model.FlexString(strconv.Itoa(port))
		node.Name = name
		node.Sni = sni
	} else {
		err = fmt.Errorf("unknown subscription type")
	}
	return
}

func (h *Handler) decodeSubscription(subscriptionId int64, subscriptionName string, data []byte) (res []*model.V2rayNode, err error) {
	header := log.NewHeader("decodeSubscription")
	decodeBs, err := tryDecode(string(data))
	if err != nil {
		log.Errorf(header, "decode response body error: %v", err)
		return
	}

	for _, line := range split(string(decodeBs)) {
		if line == "" {
			continue
		}
		//s := strings.TrimRight(strings.TrimPrefix(line, "vmess://"), "=")
		server, err := h.decodeSubscriptionLine(header, subscriptionId, subscriptionName, line)
		if err != nil {
			header.Errorf("decodeSubscriptionLine error: %v, line: %v", err, line)
			continue
		}
		res = append(res, server)
	}
	return
}

func tryDecode(str string) (de []byte, err error) {
	de, err = base64.StdEncoding.DecodeString(str)
	if err == nil {
		return
	}

	de, err = base64.RawStdEncoding.DecodeString(str)
	if err == nil {
		return
	}
	de, err = base64.URLEncoding.DecodeString(str)
	if err == nil {
		return
	}
	de, err = base64.RawURLEncoding.DecodeString(str)
	if err == nil {
		return
	}

	return nil, fmt.Errorf("no proper base64 decode method for: " + str)
}

func encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

var sep = map[rune]bool{
	' ':  true,
	'\n': true,
	',':  true,
	';':  true,
	'\t': true,
	'\f': true,
	'\v': true,
	'\r': true,
}

func split(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return sep[r]
	})
}
