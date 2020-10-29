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

const cacheForAllCustomProxyWebsites = "proxy_web_site_values"

func GetAllCustomProxyWebsitesInOneCache() ([]*model.PacWebSite, error) {
	h := log.NewHeader("GetAllCustomProxyWebsitesInOneCache")
	res := make([]*model.PacWebSite, 0)
	if cache != nil {
		bytes, err := cache.Get(context.Background(), cacheForAllCustomProxyWebsites).Bytes()
		if err == nil {
			return res, json.Unmarshal(bytes, &res)
		}
	}

	if err := db.Find(&res).Error; err != nil {
		return nil, err
	}
	bytes, err := json.Marshal(res)
	if err == nil && cache != nil {
		if err := cache.Set(context.Background(), cacheForAllCustomProxyWebsites, bytes, time.Second*3).Err(); err != nil {
			log.Errorf(h, "set cache to redis error: %v", err)
		}
	} else {
		log.Errorf(h, "marshal list error: %v", err)
	}

	return res, nil
}

func GetAllCustomProxyWebsites() ([]*model.PacWebSite, error) {
	h := log.NewHeader("GetAllCustomProxyWebsites")
	res := make([]*model.PacWebSite, 0)
	if cache != nil {
		keys, err := cache.Get(context.Background(), cacheFromAllAllCustomProxyWebsitesKey).Result()
		keysSlice := make([]string, 0)
		if err == nil {
			keysSlice = strings.Split(keys, ",")
			for _, key := range keysSlice {
				proxyWebSite := &model.PacWebSite{}
				if err := cache.Get(context.Background(), key).Scan(proxyWebSite); err != nil {
					break
				}
				res = append(res, proxyWebSite)
			}
		} else {
			log.Errorf(h, "query redis fail: %v", err)
		}
		if err == nil && len(res) == len(keysSlice) {
			return res, nil
		}
	}
	//log.Infof(h, "cache miss")
	if err := db.Find(&res).Error; err != nil {
		return nil, err
	}
	cacheKeys := make([]string, 0)
	if cache != nil {
		for _, backToDBModeler := range res {
			if err := cache.Set(context.Background(), backToDBModeler.GetCacheKey(), backToDBModeler, time.Second*3).Err(); err != nil {
				log.Errorf(h, "set cache to redis error %v", err)
			}
			cacheKeys = append(cacheKeys, backToDBModeler.GetCacheKey())
		}
		if err := cache.Set(context.Background(), cacheFromAllAllCustomProxyWebsitesKey, strings.Join(cacheKeys, ","), time.Second*3).Err(); err != nil {
			log.Errorf(h, "set all cache keys to redis error %v", err)
		}
	}

	return res, nil
}
