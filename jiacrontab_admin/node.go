package admin

import (
	"jiacrontab/models"
	"jiacrontab/pkg/log"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"
	"jiacrontab/pkg/util"

	"github.com/kataras/iris"
)

// GetnodeList 获得任务节点列表
func getNodeList(c iris.Context) {
	var (
		ctx      = wrapCtx(c)
		err      error
		nodeList []models.Node
		reqBody  getNodeListReqParams
		groupID  int
	)
	if groupID, err = ctx.getGroupIDFromToken(); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if reqBody.GroupID != 0 && groupID == 0 {
		groupID = reqBody.GroupID
	}

	sInfo := util.SystemInfo(cfg.ServerStartTime)
	if groupID == 0 {
		log.Info("here:", groupID, "page:", reqBody.Page, "pagesize:", reqBody.Pagesize)
		err = models.DB().Offset(reqBody.Page - 1).Limit(reqBody.Pagesize).Find(&nodeList).Error
	} else {
		err = models.DB().Where("group_id=?", groupID).Offset(reqBody.Page - 1).Limit(reqBody.Pagesize).Find(&nodeList).Error
	}

	if err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", map[string]interface{}{
		"nodeList":   nodeList,
		"systemInfo": sInfo,
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
