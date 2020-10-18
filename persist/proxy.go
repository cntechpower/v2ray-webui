package persist

import (
	"context"
	"strings"
	"time"

	"cntechpower.com/api-server/log"

	"cntechpower.com/api-server/model"
)

const cacheFromAllAllCustomProxyWebsitesKey = "proxy_web_site_keys"

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
