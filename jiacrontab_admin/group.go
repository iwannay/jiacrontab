package admin

import (
	"jiacrontab/models"
	"jiacrontab/pkg/proto"

	"github.com/kataras/iris"
)

// GetGroupList 获得所有的分组列表
// 调用者需要为group_id=0
func GetGroupList(c iris.Context) {
	var (
		ctx       = wrapCtx(c)
		err       error
		groupList []models.Group
		reqBody   GetGroupListReqParams
		groupID   uint
	)
	if groupID, err = ctx.getGroupIDFromToken(); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if groupID != 0 {
		ctx.respError(proto.Code_NotAllowed, proto.Msg_NotAllowed, nil)
	}

	err = models.DB().Offset(reqBody.Page - 1).Limit(reqBody.Pagesize).Group("create_at").Find(&groupList).Error

	if err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", groupList)

}

// EditGroup 编辑分组，目前仅支持分组名
func EditGroup(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		reqBody EditGroupReqParams
		err     error
		group   models.Group
	)

	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}
	group.ID = reqBody.GroupID
	group.Name = reqBody.Name

	if err = group.Save(); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}
	ctx.pubEvent(event_EditGroup, "", reqBody)
	ctx.respSucc("", nil)
}

// SetGroup 设置分组
// 当分组不为0
func SetGroup(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		reqBody SetGroupReqParams
		err     error
		user    models.User
		node    models.Node
	)

	if err = reqBody.verify(c); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if reqBody.UserID != 0 {
		user.ID = reqBody.UserID
		user.GroupID = reqBody.TargetGroupID
		if err = user.SetGroup(); err != nil {
			ctx.respError(proto.Code_Error, err.Error(), nil)
			return
		}
	}

	if reqBody.NodeAddr != "" {
		node.Addr = reqBody.NodeAddr
		node.GroupID = reqBody.TargetGroupID
		if err = node.SetGroup(); err != nil {
			ctx.respError(proto.Code_Error, err.Error(), nil)
			return
		}
	}

	ctx.pubEvent(event_SetSrouceGroup, "", reqBody)
	ctx.respSucc("", nil)

}
