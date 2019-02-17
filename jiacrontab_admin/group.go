package admin

import (
	"jiacrontab/models"
	"jiacrontab/pkg/proto"

	"github.com/kataras/iris"
)

func getGroupList(c iris.Context) {
	var (
		ctx       = wrapCtx(c)
		err       error
		groupList []models.Group
		reqBody   pageReqParams
		groupID   int
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
		err = models.DB().Offset(reqBody.Page - 1).Limit(reqBody.Pagesize).Find(&groupList).Error
	} else {
		err = models.DB().Where("group_id=?", groupID).Offset(reqBody.Page - 1).Limit(reqBody.Pagesize).Find(&groupList).Error
	}

	if err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", groupList)

}

func editGroup(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		reqBody editGroupReqParams
		err     error
		group   models.Group
	)

	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}
	group.ID = reqBody.GroupID
	group.Name = reqBody.Name
	group.NodeAddr = reqBody.NodeAddr
	if err = group.Save(); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}
	ctx.pubEvent(event_EditGroup, reqBody.NodeAddr, reqBody)
	ctx.respSucc("", nil)
}

func setGroup(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		reqBody setGroupReqParams
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
