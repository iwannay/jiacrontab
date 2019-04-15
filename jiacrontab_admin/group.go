package admin

import (
	"jiacrontab/models"
	"jiacrontab/pkg/proto"

	"github.com/kataras/iris"
)

// GetGroupList 获得所有的分组列表
func GetGroupList(c iris.Context) {
	var (
		ctx       = wrapCtx(c)
		err       error
		groupList []models.Group
		reqBody   GetGroupListReqParams
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
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

// EditGroup 编辑分组
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
	group.Name = reqBody.GroupName

	if err = group.Save(); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}
	ctx.pubEvent(group.Name, event_EditGroup, "", reqBody)
	ctx.respSucc("", nil)
}
