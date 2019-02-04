package admin

import (
	"jiacrontab/model"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"
	"jiacrontab/pkg/util"

	"github.com/kataras/iris"
)

// GetnodeList 获得任务节点列表
func getNodeList(c iris.Context) {
	var nodeList []models.Node
	sInfo := util.SystemInfo(cfg.ServerStartTime)
	ctx := wrapCtx(c)
	model.DB().Model(&models.Node{}).Find(&nodeList)

	ctx.respSucc("", map[string]interface{}{
		"nodeList":       nodeList,
		"systemInfoList": sInfo,
	})
}

func deleteNode(c iris.Context) {
	var (
		err     error
		ctx     = wrapCtx(c)
		reqBody deleteNodeReqParams
		node    models.Node
	)
	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if err = node.Delete(reqBody.NodeID); err == nil {
		rpc.DelNode(reqBody.Addr)
		ctx.respError(proto.Code_Error, "删除失败", nil)
		return
	}

	ctx.respSucc("", nil)
}
