package server_api

import (
	"context"
	"encoding/json"

	v1 "gov2panel/api/server_api/v1"
	"gov2panel/internal/model/model"
	"gov2panel/internal/service"

	"github.com/gogf/gf/v2/net/ghttp"
)

func (c *ControllerV1) Config(ctx context.Context, req *v1.ConfigReq) (res *v1.ConfigRes, err error) {
	res = &v1.ConfigRes{}
	server, routeList, err := service.ProxyService().GetServiceAndRouteListById(req.NodeId)
	if err != nil {
		return
	}

	routeArr := make([]*model.Route, 0)

	for i := 0; i < len(routeList); i++ {

		var strSlice []string
		err = json.Unmarshal([]byte(routeList[i].Match), &strSlice)

		routeArr = append(routeArr, &model.Route{
			Id:          routeList[i].Id,
			Action:      routeList[i].Action,
			Match:       strSlice,
			ActionValue: routeList[i].ActionValue,
		})
	}

	json.Unmarshal([]byte(server.ServiceJson), &res)
	ress := map[string]interface{}(*res)
	ress["routes"] = routeArr
	ress["flow_rate"] = server.Rate //流量倍率

	// ress["plan"] = planList
	ghttp.RequestFromCtx(ctx).Response.WriteJsonExit(ress)
	return
}

func (c *ControllerV1) User(ctx context.Context, req *v1.UserReq) (res *v1.UserRes, err error) {
	res = &v1.UserRes{}
	_, planIds, err := service.ProxyService().GetServicePlanIdsById(req.NodeId)
	if err != nil {
		return
	}

	userArr, err := service.User().GetUserListByGroupIds(planIds)
	if err != nil {
		return
	}

	planArr, err := service.Plan().GetPlanResetTrafficMethod1List()
	if err != nil {
		return
	}

	for _, user := range userArr {

		var speedLimit int
		for _, plan := range planArr {
			if user.GroupId == plan.Id {
				speedLimit = plan.SpeedLimit
			}

		}
		u := map[string]interface{}{
			"id":          user.Id,
			"uuid":        user.Uuid,
			"speed_limit": speedLimit,
		}
		if speedLimit <= 0 {
			u["speed_limit"] = nil
		}
		res.Users = append(res.Users, u)
	}
	ghttp.RequestFromCtx(ctx).Response.WriteJsonExit(res)

	return
}

func (c *ControllerV1) Push(ctx context.Context, req *v1.PushReq) (res *v1.PushRes, err error) {
	res = &v1.PushRes{}

	decoder := json.NewDecoder(ghttp.RequestFromCtx(ctx).Request.Body)
	decoder.Decode(&req.Data)

	err = service.User().UpUserUAndDBy(req.Data)
	if err != nil {
		return
	}
	err = service.ProxyService().CacheServiceFlow(ghttp.RequestFromCtx(ctx).Get("node_id").Int(), req.Data)

	return
}
