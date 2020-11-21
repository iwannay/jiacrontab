package admin

import (
	"fmt"
	"jiacrontab/models"
	"time"

	"gorm.io/gorm"
)

func LogInfo(ctx *myctx) {
	if !ctx.isSuper() {
		ctx.respNotAllowed()
		return
	}
	var jobTotal int64
	err := models.DB().Model(&models.JobHistory{}).Count(&jobTotal).Error
	if err != nil {
		ctx.respDBError(err)
	}
	var eventTotal int64
	err = models.DB().Model(&models.Event{}).Count(&eventTotal).Error
	if err != nil {
		ctx.respDBError(err)
	}
	ctx.respSucc("", map[string]interface{}{
		"event_total": eventTotal,
		"job_total":   jobTotal,
	})
}

func CleanLog(ctx *myctx) {
	var (
		err     error
		reqBody CleanLogParams
		isSuper = ctx.isSuper()
	)
	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}
	if !isSuper {
		ctx.respNotAllowed()
		return
	}
	offset := time.Now()
	if reqBody.Unit == "day" {
		offset = offset.AddDate(0, 0, -reqBody.Offset)
	}
	if reqBody.Unit == "month" {
		offset = offset.AddDate(0, -reqBody.Offset, 0)
	}
	var tx *gorm.DB
	if reqBody.IsEvent {
		tx = models.DB().Where("created_at<?", offset).Delete(&models.Event{})

	} else {
		tx = models.DB().Where("created_at<?", offset).Delete(&models.JobHistory{})
	}
	err = tx.Error
	if err != nil {
		ctx.respDBError(err)
		return

	}
	if reqBody.IsEvent {
		ctx.pubEvent(fmt.Sprintf("%d %s", reqBody.Offset, reqBody.Unit), event_CleanUserEvent, "", reqBody)
	} else {
		ctx.pubEvent(fmt.Sprintf("%d %s", reqBody.Offset, reqBody.Unit), event_CleanJobHistory, "", reqBody)
	}

	ctx.respSucc("清理成功", map[string]interface{}{
		"total": tx.RowsAffected,
	})
	return
}
