package admin

import (
	"jiacrontab/models"
	"jiacrontab/pkg/proto"

	"github.com/kataras/iris"
)

// GetnodeList 获得任务节点列表
// groupID为0的分组为超级管理员分组,该分组中保留所有的node节点信息
// 当新建分组时，copy超级管理员分组中的节点到新的分组
// 超级管理员获得所有的节点
// 普通用户获得所属分组节点
func GetNodeList(c iris.Context) {
	var (
		ctx      = wrapCtx(c)
		err      error
		nodeList []models.Node
		reqBody  GetNodeListReqParams
		groupID  uint
		count    int
	)
	if groupID, err = ctx.getGroupIDFromToken(); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if groupID == models.SuperGroup.ID {
		err = models.DB().Offset(reqBody.Page - 1).Limit(reqBody.Pagesize).Find(&nodeList).Error
		models.DB().Model(&models.Node{}).Count(&count)
	} else {
		err = models.DB().Where("group_id=?", groupID).Offset(reqBody.Page - 1).Limit(reqBody.Pagesize).Find(&nodeList).Error
		models.DB().Model(&models.Node{}).Where("group_id=?", groupID).Count(&count)
	}

	if err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", map[string]interface{}{
		"list":     nodeList,
		"total":    count,
		"page":     reqBody.Page,
		"pagesize": reqBody.Pagesize,
	})
}

// DeleteNode 删除分组内节点
// 仅超级管理员有权限
func DeleteNode(c iris.Context) {
	var (
		err     error
		ctx     = wrapCtx(c)
		reqBody DeleteNodeReqParams
		group   models.Group
		node    models.Node
	)

	if err = reqBody.verify(ctx); err != nil {
		ctx.respBasicError(err)
		return
	}

	if err = ctx.parseClaimsFromToken(); err != nil {
		ctx.respBasicError(err)
		return
	}

	// 普通用户不允许删除节点
	if ctx.claims.GroupID != 0 {
		ctx.respNotAllowed()
		return
	}

	if err := models.DB().Take(&group, "id=?", reqBody.GroupID).Error; err != nil {
		ctx.respDBError(err)
		return
	}

	if err = node.Delete(reqBody.GroupID, reqBody.Addr); err != nil {
		ctx.respDBError(err)
		return
	}

	ctx.pubEvent(node.Name, event_DelNodeDesc, reqBody.Addr, group.Name)
	ctx.respSucc("", nil)
}

// GroupNode 超级管理员为node分组
// 分组不存在时自动创建分组
func GroupNode(c iris.Context) {
	var (
		err     error
		ctx     = wrapCtx(c)
		reqBody GroupNodeReqParams
		node    models.Node
	)

	if !ctx.isSuper() {
		ctx.respNotAllowed()
		return
	}

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if err = node.GroupNode(reqBody.Addr, reqBody.TargetGroupID,
		reqBody.TargetNodeName, reqBody.TargetGroupName); err != nil {
		ctx.respError(proto.Code_Error, "分组失败", err)
		return
	}

	ctx.pubEvent(reqBody.TargetNodeName, event_GroupNode, reqBody.Addr, reqBody)
	ctx.respSucc("", nil)
}
