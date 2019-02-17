package admin

import (
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"

	"github.com/kataras/iris"
)

// GetnodeList 获得任务节点列表
func getNodeList(c iris.Context) {
	var (
		ctx      = wrapCtx(c)
		err      error
		nodeList []models.Node
		reqBody  pageReqParams
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

	if groupID == 0 {
		err = models.DB().Offset(reqBody.Page - 1).Limit(reqBody.Pagesize).Find(&nodeList).Error
	} else {
		err = models.DB().Where("group_id=?", groupID).Offset(reqBody.Page - 1).Limit(reqBody.Pagesize).Find(&nodeList).Error
	}

	if err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", nodeList)
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
	ctx.pubEvent(event_DelNodeDesc, reqBody.Addr, "")
	ctx.respSucc("", nil)
}
