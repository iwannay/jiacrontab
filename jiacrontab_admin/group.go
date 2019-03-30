package admin

import (
	"database/sql"
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

func GetUserList(c iris.Context) {
	var (
		ctx      = wrapCtx(c)
		reqBody  GetUsersParams
		userList []models.User
		err      error
		total    int
	)
	if err = reqBody.verify(ctx); err != nil {
		ctx.respBasicError(err)
		return
	}
	if err = ctx.parseClaimsFromToken(); err != nil {
		ctx.respJWTError(err)
		return
	}

	if reqBody.QueryGroupID != ctx.claims.GroupID && ctx.claims.GroupID != 0 {
		ctx.respNotAllowed()
		return
	}

	err = models.DB().Model(&models.User{}).Where("group_id=?", reqBody.QueryGroupID).Count(&total).Error
	if err != nil && err != sql.ErrNoRows {
		ctx.respBasicError(err)
		return
	}

	err = models.DB().Where("group_id=?", reqBody.QueryGroupID).Offset(reqBody.Page - 1).Limit(reqBody.Pagesize).Find(&userList).Error
	if err != nil && err != sql.ErrNoRows {
		ctx.respBasicError(err)
		return
	}

	ctx.respSucc("", map[string]interface{}{
		"list":     userList,
		"total":    total,
		"page":     reqBody.Page,
		"pagesize": reqBody.Pagesize,
	})
}

// EditGroup 编辑分组
// 当groupID=0时创建新的分组
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
	ctx.pubEvent(event_EditGroup, "", reqBody)
	ctx.respSucc("", nil)
}

// GroupUser 超级管理员设置普通用户分组
// 超级管理员
func GroupUser(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		reqBody SetGroupReqParams
		err     error
		user    models.User
		groupID uint
	)

	if err = reqBody.verify(c); err != nil {
		ctx.respBasicError(err)
		return
	}

	if groupID, err = ctx.getGroupIDFromToken(); err != nil {
		ctx.respJWTError(err)
		return
	}

	if groupID != 0 {
		ctx.respNotAllowed()
		return
	}

	if reqBody.UserID != 0 {
		user.ID = reqBody.UserID
		user.GroupID = reqBody.TargetGroupID
		user.Root = reqBody.Root
		if err = user.SetGroup(); err != nil {
			ctx.respBasicError(err)
			return
		}
	}

	ctx.pubEvent(event_GroupUser, "", reqBody)
	ctx.respSucc("", nil)

}
