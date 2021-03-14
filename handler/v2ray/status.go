package v2ray

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	stats "v2ray.com/core/app/stats/command"

	"github.com/cntechpower/utils/log"
	"github.com/cntechpower/v2ray-webui/model"
)

const (
	grpcPort = 10085
)

func (h *Handler) refreshStatusLoop() {
	ticker := time.NewTicker(30 * time.Second)
	header := log.NewHeader("refreshStatusLoop")
	for range ticker.C {
		h.refreshCurrentNodePing(header)
		h.v2rayStatusRefreshTime = time.Now()
		header.Infof("refreshed status")
	}

}

func (h *Handler) Status(inboundTags, outboundTags []string) (res *model.V2rayStatus, err error) {
	l := log.NewHeader("Status")
	res = &model.V2rayStatus{
		InboundTraffic:  make([]model.Traffic, 0),
		OutboundTraffic: make([]model.Traffic, 0),
	}
	h.v2rayConfigMu.Lock()
	if h.v2rayCurrentNode != nil {
		res.CurrentNode = h.v2rayCurrentNode.DeepCopy()
	}
	h.v2rayConfigMu.Unlock()

	res.Core = &model.V2rayCoreStatus{}
	h.v2rayServerMu.Lock()
	res.Core.StartTime = h.v2rayServerStartTime.Format(time.RFC3339)
	h.v2rayServerMu.Unlock()
	res.RefreshTime = h.v2rayStatusRefreshTime.Format(time.RFC3339)
	if len(inboundTags) != 0 || len(outboundTags) != 0 {
		var cc *grpc.ClientConn
		cc, err = grpc.Dial(fmt.Sprintf("127.0.0.1:%v", grpcPort), grpc.WithInsecure())
		if err != nil {
			l.Errorf("grpc.Dial error: %v", err)
			return
		}
		defer cc.Close()
		client := stats.NewStatsServiceClient(cc)
		for _, tag := range inboundTags {
			statRes, errGrpc := client.GetStats(context.Background(), &stats.GetStatsRequest{
				Name: fmt.Sprintf("inbound>>>%v>>>traffic>>>uplink", tag),
			})
			if errGrpc != nil {
				err = errGrpc
				l.Errorf("client.GetStats error: %v", err)
				return
			}
			res.InboundTraffic = append(res.InboundTraffic, model.Traffic{
				Tag:     statRes.Stat.Name,
				Traffic: statRes.Stat.Value,
			})
		}
		for _, tag := range outboundTags {
			statRes, errGrpc := client.GetStats(context.Background(), &stats.GetStatsRequest{
				Name: fmt.Sprintf("outbound>>>%v>>>traffic>>>downlink", tag),
			})
			if errGrpc != nil {
				err = errGrpc
				l.Errorf("client.GetStats error: %v", err)
				return
			}
			res.OutboundTraffic = append(res.OutboundTraffic, model.Traffic{
				Tag:     statRes.Stat.Name,
				Traffic: statRes.Stat.Value,
			})
		}

	}

	return
}
