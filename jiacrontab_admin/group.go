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
	)

	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if !ctx.isSuper() {
		ctx.respNotAllowed()
		return
	}

	err = models.DB().Offset(reqBody.Page - 1).Limit(reqBody.Pagesize).Find(&groupList).Error

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
