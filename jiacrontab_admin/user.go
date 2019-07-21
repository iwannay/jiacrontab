package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/util"
	"jiacrontab/pkg/version"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type CustomerClaims struct {
	jwt.StandardClaims
	Version  int64
	UserID   uint
	Mail     string
	Username string
	GroupID  uint
	Root     bool
}

// Login 用户登录
func Login(ctx *myctx) {
	var (
		err            error
		reqBody        LoginReqParams
		user           models.User
		customerClaims CustomerClaims
		cfg            = ctx.adm.getOpts()
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}
	if !user.Verify(reqBody.Username, reqBody.Passwd) {
		ctx.respAuthFailed(errors.New("帐号或密码不正确"))
		return
	}

	customerClaims.ExpiresAt = cfg.Jwt.Expires + time.Now().Unix()
	customerClaims.Username = reqBody.Username
	customerClaims.UserID = user.ID
	customerClaims.Mail = user.Mail
	customerClaims.GroupID = user.GroupID
	customerClaims.Root = user.Root
	customerClaims.Version = user.Version

	if reqBody.Remember {
		customerClaims.ExpiresAt = time.Now().Add(24 * 30 * time.Hour).Unix()
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, customerClaims).SignedString([]byte(cfg.Jwt.SigningKey))

	if err != nil {
		ctx.respAuthFailed(errors.New("无法生成访问凭证"))
		return
	}

	ctx.respSucc("", map[string]interface{}{
		"token":    token,
		"groupID":  user.GroupID,
		"root":     user.Root,
		"mail":     user.Mail,
		"username": user.Username,
		"userID":   user.ID,
	})
}

func GetActivityList(ctx *myctx) {
	var (
		err     error
		reqBody ReadMoreReqParams
		events  []models.Event
		isSuper bool
		model   = models.DB().Debug()
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if ctx.isSuper() {
		isSuper = true
	}

	if reqBody.LastID == 0 {
		if !isSuper {
			model = model.Where("group_id=?", ctx.claims.GroupID)
		}
		err = model.Order(fmt.Sprintf("created_at %s", reqBody.Orderby)).
			Limit(reqBody.Pagesize).
			Find(&events).Error
	} else {
		if !isSuper {
			model = model.Where("group_id=? and id<?", ctx.claims.GroupID, reqBody.LastID)
		} else {
			model = model.Where("id<?", reqBody.LastID)
		}
		err = model.Order(fmt.Sprintf("created_at %s", reqBody.Orderby)).
			Limit(reqBody.Pagesize).
			Find(&events).Error
	}

	if err != nil && err != sql.ErrNoRows {
		ctx.respDBError(err)
		return
	}

	ctx.respSucc("", map[string]interface{}{
		"list":     events,
		"pagesize": reqBody.Pagesize,
	})
}

func GetJobHistory(ctx *myctx) {
	var (
		err      error
		reqBody  ReadMoreReqParams
		historys []models.JobHistory
		addrs    []string
		isSuper  bool
		model    = models.DB()
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if ctx.isSuper() {
		isSuper = true
	}

	if addrs, err = ctx.getGroupAddr(); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), err)
		return
	}

	if reqBody.LastID == 0 {
		if !isSuper {
			model = model.Where("addr in (?)", addrs)
		}
		err = model.Order(fmt.Sprintf("created_at %s", reqBody.Orderby)).
			Limit(reqBody.Pagesize).
			Find(&historys).Error
	} else {
		if !isSuper {
			model = model.Where("addr in (?) and id<?", addrs, reqBody.LastID)
		} else {
			model = model.Where("id<?", reqBody.LastID)
		}
		err = model.Where("addr in (?) and id<?", addrs, reqBody.LastID).
			Order(fmt.Sprintf("created_at %s", reqBody.Orderby)).
			Limit(reqBody.Pagesize).
			Find(&historys).Error
	}

	if err != nil && err != sql.ErrNoRows {
		ctx.respDBError(err)
		return
	}

	ctx.respSucc("", map[string]interface{}{
		"list":     historys,
		"pagesize": reqBody.Pagesize,
	})
}

func AuditJob(ctx *myctx) {
	var (
		err     error
		reqBody AuditJobReqParams
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respBasicError(err)
		return
	}

	if !ctx.verifyNodePermission(reqBody.Addr) {
		ctx.respNotAllowed()
		return
	}

	if ctx.claims.GroupID != models.SuperGroup.ID && !ctx.claims.Root {
		ctx.respNotAllowed()
		return
	}

	if reqBody.JobType == "crontab" {
		var reply []models.CrontabJob
		if err = rpcCall(reqBody.Addr, "CrontabJob.Audit", proto.AuditJobArgs{
			Root:    ctx.claims.Root,
			GroupID: ctx.claims.GroupID,
			JobIDs:  reqBody.JobIDs,
		}, &reply); err != nil {
			ctx.respRPCError(err)
			return
		}
		var targetNames []string
		for _, v := range reply {
			targetNames = append(targetNames, v.Name)
		}
		ctx.pubEvent(strings.Join(targetNames, ","), event_AuditCrontabJob, models.EventSourceName(reqBody.Addr), reqBody)
	} else {
		var reply []models.DaemonJob
		if err = rpcCall(reqBody.Addr, "DaemonJob.Audit", proto.AuditJobArgs{
			Root:    ctx.claims.Root,
			GroupID: ctx.claims.GroupID,
			JobIDs:  reqBody.JobIDs,
		}, &reply); err != nil {
			ctx.respRPCError(err)
			return
		}
		var targetNames []string
		for _, v := range reply {
			targetNames = append(targetNames, v.Name)
		}
		ctx.pubEvent(strings.Join(targetNames, ","), event_AuditDaemonJob, models.EventSourceName(reqBody.Addr), reqBody)
	}

	ctx.respSucc("", nil)
}

// Signup 注册新用户
func Signup(ctx *myctx) {
	var (
		err     error
		user    models.User
		reqBody UserReqParams
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if !ctx.isSuper() {
		ctx.respNotAllowed()
		return
	}

	if reqBody.GroupName != "" {
		group := models.Group{
			Name: reqBody.GroupName,
		}
		if err = models.DB().Save(&group).Error; err != nil {
			ctx.respDBError(err)
			return
		}
		reqBody.GroupID = group.ID
	}

	user.Username = reqBody.Username
	user.Passwd = reqBody.Passwd
	user.GroupID = reqBody.GroupID
	user.Root = reqBody.Root
	user.Avatar = reqBody.Avatar
	user.Mail = reqBody.Mail
	reqBody.Passwd = ""
	if user.GroupID == models.SuperGroup.ID {
		user.Root = true
	}

	if err = user.Create(); err != nil {
		ctx.respDBError(err)
		return
	}

	ctx.pubEvent(user.Username, event_SignUpUser, "", reqBody)
	ctx.respSucc("", true)
}

func EditUser(ctx *myctx) {
	var (
		err     error
		user    models.User
		reqBody EditUserReqParams
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	user.ID = reqBody.UserID
	user.Username = reqBody.Username
	user.Passwd = reqBody.Passwd
	user.Avatar = reqBody.Avatar
	user.Mail = reqBody.Mail

	if err = user.Update(); err != nil {
		ctx.respDBError(err)
		return
	}

	ctx.pubEvent(user.Username, event_EditUser, "", reqBody)
	ctx.respSucc("", true)
}

func DeleteUser(ctx *myctx) {
	var (
		err     error
		user    models.User
		reqBody DeleteUserReqParams
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if !ctx.isSuper() {
		ctx.respNotAllowed()
		return
	}
	user.ID = reqBody.UserID
	if err = user.Delete(); err != nil {
		ctx.respDBError(err)
		return
	}
	ctx.pubEvent(user.Username, event_DeleteUser, "", reqBody)
	ctx.respSucc("", true)
}

// UserStat 统计信息
func UserStat(ctx *myctx) {
	var (
		err          error
		auditNumStat struct {
			CrontabJobAuditNum  uint
			DaemonJobAuditNum   uint
			CrontabJobFailNum   uint
			DaemonJobRunningNum uint
			NodeNum             uint
		}
		cfg = ctx.adm.getOpts()
	)

	err = models.DB().Raw(
		`select 
			sum(crontab_job_audit_num) as crontab_job_audit_num, 
			sum(daemon_job_audit_num) as daemon_job_audit_num,
			sum(crontab_job_fail_num) as crontab_job_fail_num,
			sum(daemon_job_running_num) as daemon_job_running_num,
			count(*) as node_num
		from nodes 
		where group_id=?`, ctx.claims.GroupID).Scan(&auditNumStat).Error
	if err != nil {
		ctx.respDBError(err)
		return
	}

	ctx.respSucc("", map[string]interface{}{
		"systemInfo": util.SystemInfo(cfg.ServerStartTime),
		"auditStat":  auditNumStat,
		"version":    version.String("jiacrontab"),
	})
}

// GroupUser 超级管理员设置普通用户分组
func GroupUser(ctx *myctx) {
	var (
		reqBody SetGroupReqParams
		err     error
		user    models.User
		group   models.Group
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respBasicError(err)
		return
	}

	if !ctx.isSuper() {
		ctx.respNotAllowed()
		return
	}

	if reqBody.TargetGroupName != "" {
		group.Name = reqBody.TargetGroupName
		if err = models.DB().Save(&group).Error; err != nil {
			ctx.respDBError(err)
			return
		}
		reqBody.TargetGroupID = group.ID
	}

	user.ID = reqBody.UserID
	user.GroupID = reqBody.TargetGroupID

	if reqBody.TargetGroupID == models.SuperGroup.ID {
		user.Root = true
	} else {
		user.Root = reqBody.Root
	}

	if err = user.SetGroup(&group); err != nil {
		ctx.respDBError(err)
		return
	}

	ctx.pubEvent(group.Name, event_GroupUser, models.EventSourceUsername(user.Username), reqBody)
	ctx.respSucc("", nil)
}

// GetUserList 获得用户列表
// 支持获得全部用户，所属分组用户，指定分组用户（超级管理员）
func GetUserList(ctx *myctx) {
	var (
		reqBody  GetUsersParams
		userList []models.User
		err      error
		total    int
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
	}

	if reqBody.IsAll && ctx.claims.GroupID != models.SuperGroup.ID {
		ctx.respNotAllowed()
		return
	}

	if !reqBody.IsAll && reqBody.QueryGroupID != ctx.claims.GroupID && ctx.claims.GroupID != models.SuperGroup.ID {
		ctx.respNotAllowed()
		return
	}

	if reqBody.QueryGroupID == 0 {
		reqBody.QueryGroupID = ctx.claims.GroupID
	}

	m := models.DB().Model(&models.User{})
	if reqBody.IsAll {
		err = m.Where("username like ?", "%"+reqBody.SearchTxt+"%").Count(&total).Error
	} else {
		err = m.Where("group_id=? and username like ?", reqBody.QueryGroupID, "%"+reqBody.SearchTxt+"%").Count(&total).Error
	}

	if err != nil && err != sql.ErrNoRows {
		ctx.respBasicError(err)
		return
	}

	if reqBody.IsAll {
		err = models.DB().Preload("Group").Where("username like ?", "%"+reqBody.SearchTxt+"%").Order("id desc").Offset((reqBody.Page - 1) * reqBody.Pagesize).Limit(reqBody.Pagesize).Find(&userList).Error
	} else {
		err = models.DB().Preload("Group").Where("group_id=? and username like ?", reqBody.QueryGroupID, "%"+reqBody.SearchTxt+"%").Offset((reqBody.Page - 1) * reqBody.Pagesize).Limit(reqBody.Pagesize).Find(&userList).Error
	}

	if err != nil && err != sql.ErrNoRows {
		ctx.respDBError(err)
		return
	}

	ctx.respSucc("", map[string]interface{}{
		"list":     userList,
		"total":    total,
		"page":     reqBody.Page,
		"pagesize": reqBody.Pagesize,
	})
}
