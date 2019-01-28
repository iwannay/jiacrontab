package handle

import (
	"jiacrontab/model"
	"jiacrontab/pkg/util"
	"jiacrontab/server/conf"
	"time"

	"github.com/kataras/iris"
)

// Index 服务器列表页面
func Index(ctx iris.Context) {
	sInfo := util.SystemInfo(conf.AppService.ServerStartTime)

	var clientList []model.Client
	model.DB().Model(&model.Client{}).Find(&clientList)

	for k, v := range clientList {
		if time.Now().Sub(v.UpdatedAt) > 10*time.Minute {
			clientList[k].State = 0
		}
	}

	ctx.ViewData("clientList", clientList)
	ctx.ViewData("systemInfoList", sInfo)
	ctx.View("index.html")
}

// GetClientList 获得任务节点列表
func GetClientList(ctx iris.Context) {
	sInfo := util.SystemInfo(conf.AppService.ServerStartTime)

	var clientList []model.Client
	model.DB().Model(&model.Client{}).Find(&clientList)

	for k, v := range clientList {
		if time.Now().Sub(v.UpdatedAt) > 10*time.Minute {
			clientList[k].State = 0
		}
	}

	ctx.JSON(successResp("", map[string]interface{}{
		"clientList":     clientList,
		"systemInfoList": sInfo,
	}))
}
