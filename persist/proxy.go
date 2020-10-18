package persist

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"cntechpower.com/api-server/log"

	"cntechpower.com/api-server/model"
)

const cacheFromAllAllCustomProxyWebsitesKey = "proxy_web_site_keys"

const cacheForAllCustomProxyWebsites = "proxy_web_site_keys"

func GetAllCustomProxyWebsitesInOneCache() ([]*model.ProxyWebSite, error) {
	h := log.NewHeader("GetAllCustomProxyWebsitesInOneCache")
	res := make([]*model.ProxyWebSite, 0)
	bytes, err := cache.Get(context.Background(), cacheForAllCustomProxyWebsites).Bytes()
	if err == nil {
		//log.Infof(h, "from cache")
		return res, json.Unmarshal(bytes, &res)
	}
	//log.Infof(h, "into db")
	if err := db.Find(&res).Error; err != nil {
		return nil, err
	}
	bytes, err = json.Marshal(res)
	if err == nil {
		if err := cache.Set(context.Background(), cacheFromAllAllCustomProxyWebsitesKey, bytes, time.Minute).Err(); err != nil {
			log.Errorf(h, "set cache to redis error: %v", err)
		}
	} else {
		log.Errorf(h, "marshal list error: %v", err)
	}

	return res, nil
}

func GetAllCustomProxyWebsites() ([]*model.ProxyWebSite, error) {
	h := log.NewHeader("GetAllCustomProxyWebsites")
	res := make([]*model.ProxyWebSite, 0)
	keys, err := cache.Get(context.Background(), cacheFromAllAllCustomProxyWebsitesKey).Result()
	keysSlice := make([]string, 0)
	if err == nil {
		keysSlice = strings.Split(keys, ",")
		for _, key := range keysSlice {
			proxyWebSite := &model.ProxyWebSite{}
			if err := cache.Get(context.Background(), key).Scan(proxyWebSite); err != nil {
				break
			}
			res = append(res, proxyWebSite)
		}
	} else {
		log.Errorf(h, "query redis fail: %v", err)
	}
	if err == nil && len(res) == len(keysSlice) {
		//log.Infof(h, "cache hit")
		return res, nil
	}
	//log.Infof(h, "cache miss")
	if err := db.Find(&res).Error; err != nil {
		return nil, err
	}
	cacheKeys := make([]string, 0)
	for _, backToDBModeler := range res {
		if err := cache.Set(context.Background(), backToDBModeler.GetCacheKey(), backToDBModeler, time.Minute).Err(); err != nil {
			log.Errorf(h, "set cache to redis error %v", err)
		}
		cacheKeys = append(cacheKeys, backToDBModeler.GetCacheKey())
	}
	if err := cache.Set(context.Background(), cacheFromAllAllCustomProxyWebsitesKey, strings.Join(cacheKeys, ","), time.Minute).Err(); err != nil {
		log.Errorf(h, "set all cache keys to redis error %v", err)
	}
	return res, nil
}
