package admin

import (
	"jiacrontab/model"
	"jiacrontab/models"
	"jiacrontab/pkg/util"

	"github.com/kataras/iris"
)

// GetClientList 获得任务节点列表
func getClientList(c iris.Context) {
	var clientList []models.Client
	sInfo := util.SystemInfo(cfg.ServerStartTime)
	ctx := wrapCtx(c)
	model.DB().Model(&models.Client{}).Find(&clientList)

	ctx.respSucc("", map[string]interface{}{
		"clientList":     clientList,
		"systemInfoList": sInfo,
	})
}
