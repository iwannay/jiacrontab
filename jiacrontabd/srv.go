package jiacrontabd

import (
	"errors"
	"fmt"
	"jiacrontab/models"
	"jiacrontab/pkg/crontab"
	"jiacrontab/pkg/finder"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/util"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/iwannay/log"
)

type Srv struct {
	jd *Jiacrontabd
}

func newSrv(jd *Jiacrontabd) *Srv {
	return &Srv{
		jd: jd,
	}
}

func (s *Srv) Ping(args proto.EmptyArgs, reply *proto.EmptyReply) error {
	return nil
}

func (s *Srv) SystemInfo(args proto.EmptyArgs, reply *map[string]interface{}) error {
	*reply = util.SystemInfo(s.jd.startTime)
	return nil
}

type CrontabJob struct {
	jd *Jiacrontabd
}

func newCrontabJobSrv(jd *Jiacrontabd) *CrontabJob {
	return &CrontabJob{
		jd: jd,
	}
}

func (j *CrontabJob) List(args proto.QueryJobArgs, reply *proto.QueryCrontabJobRet) error {
	err := models.DB().Model(&models.CrontabJob{}).Where("name like ?", "%"+args.SearchTxt+"%").Count(&reply.Total).Error
	if err != nil {
		return err
	}

	reply.Page = args.Page
	reply.Pagesize = args.Pagesize

	return models.DB().Where("name like ?", "%"+args.SearchTxt+"%").Order(gorm.Expr("created_user_id=? desc, id desc", args.UserID)).Offset(args.Page - 1).Limit(args.Pagesize).Find(&reply.List).Error
}

func (j *CrontabJob) Audit(args proto.AuditJobArgs, reply *[]models.CrontabJob) error {
	defer models.DB().Model(&models.CrontabJob{}).Find(reply, "id in(?)", args.JobIDs)
	return models.DB().Model(&models.CrontabJob{}).Where("id in(?) and status=?", args.JobIDs, models.StatusJobUnaudited).Update("status", models.StatusJobOk).Error
}

func (j *CrontabJob) Edit(args proto.EditCrontabJobArgs, reply *models.CrontabJob) error {

	var (
		model = models.DB()
	)

	if args.Job.MaxConcurrent == 0 {
		args.Job.MaxConcurrent = 1
	}

	if args.Job.ID == 0 {
		return models.DB().Create(&args.Job).Error
	}

	if args.Job.ID == 0 {
		model = models.DB().Save(&args.Job)
	} else {
		if args.Root {
			model = model.Where("id=?", args.Job.ID).Debug()
		} else {
			model = model.Where("id=? and created_user_id", args.Job.ID, args.Job.CreatedUserID).Debug()
		}
		model = model.Omit(
			"updated_at", "created_at", "deleted_at",
			"created_user_id", "created_username",
			"last_cost_time", "last_exec_time",
			"next_exec_time", "last_exit_status", "process_num",
		).Save(&args.Job)
	}
	*reply = args.Job
	return model.Error
}
func (j *CrontabJob) Get(args proto.GetJobArgs, reply *models.CrontabJob) error {
	return models.DB().Find(reply, "id=?", args.JobID).Error
}

func (j *CrontabJob) Start(args proto.ActionJobsArgs, jobs *[]models.CrontabJob) error {

	model := models.DB()

	if len(args.JobIDs) == 0 {
		return errors.New("empty ids")
	}

	if args.Root {
		model = model.Where("id in (?) and status in (?)",
			args.JobIDs, []models.JobStatus{models.StatusJobOk, models.StatusJobStop})
	} else {
		model = model.Where("created_user_id = ? and id in (?) and status in (?)",
			args.UserID, args.JobIDs, []models.JobStatus{models.StatusJobOk, models.StatusJobStop})
	}

	ret := model.Find(jobs)
	if ret.Error != nil {
		return ret.Error
	}

	for _, v := range *jobs {
		j.jd.addJob(&crontab.Job{
			ID:      v.ID,
			Second:  v.TimeArgs.Second,
			Minute:  v.TimeArgs.Minute,
			Hour:    v.TimeArgs.Hour,
			Day:     v.TimeArgs.Day,
			Month:   v.TimeArgs.Month,
			Weekday: v.TimeArgs.Weekday,
		})
	}

	return nil
}

func (j *CrontabJob) Stop(args proto.ActionJobsArgs, job *[]models.CrontabJob) error {
	model := models.DB()
	if args.Root {
		model = model.Where("id in (?) and status in (?)", args.JobIDs, []models.JobStatus{models.StatusJobTiming, models.StatusJobRunning})
	} else {
		model = model.Where("created_user_id = ? and id in (?) and status in (?)",
			args.UserID, args.JobIDs, []models.JobStatus{models.StatusJobTiming, models.StatusJobRunning})
	}
	return model.Find(job).Update("status", models.StatusJobStop).Error
}

func (j *CrontabJob) Delete(args proto.ActionJobsArgs, job *[]models.CrontabJob) error {

	model := models.DB()
	if args.Root {
		model = model.Where("id in (?)", args.JobIDs)
	} else {
		model = model.Where("created_user_id = ? and id in (?)",
			args.UserID, args.JobIDs)
	}
	return model.Find(job).Delete(&models.CrontabJob{}).Error
}

func (j *CrontabJob) Kill(args proto.ActionJobsArgs, job *[]models.CrontabJob) error {
	model := models.DB()
	if args.Root {
		model = model.Where("id in (?)", args.JobIDs)
	} else {
		model = model.Where("created_user_id = ? and id in (?)",
			args.UserID, args.JobIDs)
	}

	err := model.Take(job).Error
	if err != nil {
		return err
	}

	for _, jobID := range args.JobIDs {
		j.jd.killTask(jobID)
	}
	return nil
}

func (j *CrontabJob) Exec(args proto.GetJobArgs, reply *proto.ExecCrontabJobReply) error {

	model := models.DB()
	if args.Root == true {
		model = model.Where("id=?", args.JobID)
	} else {
		model = model.Where("created_user_id = ? and id=?", args.UserID, args.JobID)
	}

	ret := model.Debug().Take(&reply.Job)

	if ret.Error == nil {
		jobInstance := newJobEntry(&crontab.Job{
			ID:    reply.Job.ID,
			Value: reply.Job,
		}, j.jd)

		j.jd.addTmpJob(jobInstance)
		defer j.jd.removeTmpJob(jobInstance)

		jobInstance.once = true
		jobInstance.exec()

		reply.Content = jobInstance.waitDone()

	} else {
		reply.Content = []byte(ret.Error.Error())
		return ret.Error
	}
	return nil

}

func (j *CrontabJob) Log(args proto.SearchLog, reply *proto.SearchLogResult) error {

	fd := finder.NewFinder(func(info os.FileInfo) bool {
		basename := filepath.Base(info.Name())
		arr := strings.Split(basename, ".")
		if len(arr) != 2 {
			return false
		}

		if arr[1] == "log" && arr[0] == fmt.Sprint(args.JobID) {
			return true
		}
		return false
	})

	if args.Date == "" {
		args.Date = time.Now().Format("2006/01/02")
	}
	if args.IsTail {
		fd.SetTail(true)
	}

	rootpath := filepath.Join(j.jd.getOpts().LogPath, "crontab_task", args.Date)
	err := fd.Search(rootpath, args.Pattern, &reply.Content, args.Offset, args.Pagesize)
	reply.Offset = fd.Offset()
	reply.FileSize = fd.FileSize()
	return err

}

// SetDependDone 依赖执行完毕时设置相关状态
func (j *CrontabJob) SetDependDone(args proto.DepJob, reply *bool) error {
	*reply = j.jd.SetDependDone(&depEntry{
		jobID:       args.JobID,
		processID:   args.ProcessID,
		jobUniqueID: args.JobUniqueID,
		dest:        args.Dest,
		from:        args.From,
		done:        true,
		logContent:  args.LogContent,
		err:         args.Err,
	})
	return nil
}

// ExecDepend 执行依赖
func (j *CrontabJob) ExecDepend(args proto.DepJob, reply *bool) error {
	j.jd.dep.add(&depEntry{
		jobUniqueID: args.JobUniqueID,
		processID:   args.ProcessID,
		jobID:       args.JobID,
		id:          args.ID,
		dest:        args.Dest,
		from:        args.From,
		name:        args.Name,
		commands:    args.Commands,
	})
	*reply = true
	log.Infof("job %s %v add to execution queue ", args.Name, args.Commands)
	return nil
}

func (j *CrontabJob) Ping(args *proto.EmptyArgs, reply *proto.EmptyReply) error {
	return nil
}

type DaemonJob struct {
	jd *Jiacrontabd
}

func newDaemonJobSrv(jd *Jiacrontabd) *DaemonJob {
	return &DaemonJob{
		jd: jd,
	}
}

func (j *DaemonJob) List(args proto.QueryJobArgs, reply *proto.QueryDaemonJobRet) error {
	err := models.DB().Model(&models.DaemonJob{}).Where("name like ?", "%"+args.SearchTxt+"%").Count(&reply.Total).Error
	if err != nil {
		return err
	}

	reply.Page = args.Page
	reply.Pagesize = args.Pagesize

	return models.DB().Where("name like ?", "%"+args.SearchTxt+"%").Order(gorm.Expr("created_user_id=? desc, id desc", args.UserID)).Offset(args.Page - 1).Limit(args.Pagesize).Find(&reply.List).Error
}

func (j *DaemonJob) Edit(args proto.EditDaemonJobArgs, job *models.DaemonJob) error {

	if args.Job.ID == 0 {
		ret := models.DB().Create(&args.Job)
		return ret.Error
	}

	model := models.DB()
	if args.Root {
		model = model.Where("id=?", args.Job.ID)
	} else {
		model = model.Where("id=? and created_user_id=?", args.Job.ID, args.Job.CreatedUserID).Omit(
			"updated_at", "created_at", "deleted_at",
			"createdUserID", "createdUsername")
	}

	ret := model.Save(&args)
	*job = args.Job
	return ret.Error
}

func (j *DaemonJob) Start(args proto.ActionJobsArgs, jobs *[]models.DaemonJob) error {

	model := models.DB()
	if args.Root {
		model = model.Where("id in(?) and status in (?)",
			args.JobIDs, []models.JobStatus{models.StatusJobOk, models.StatusJobStop})
	} else {
		model = model.Where("created_user_id = ? and id in(?) and status in (?)",
			args.UserID, args.JobIDs, []models.JobStatus{models.StatusJobOk, models.StatusJobStop})
	}

	ret := model.Find(&jobs)
	if ret.Error != nil {
		return ret.Error
	}

	for _, v := range *jobs {
		job := v
		j.jd.daemon.add(&daemonJob{
			job: &job,
		})
	}

	return nil
}

func (j *DaemonJob) Stop(args proto.ActionJobsArgs, jobs *[]models.DaemonJob) error {

	model := models.DB()
	if args.Root {
		model = model.Where("id in(?) and status in (?)",
			args.JobIDs, []models.JobStatus{models.StatusJobRunning, models.StatusJobTiming})
	} else {
		model = model.Where("created_user_id = ? and id in(?) and status in (?)",
			args.UserID, args.JobIDs, []models.JobStatus{models.StatusJobRunning, models.StatusJobTiming})
	}

	if err := model.Find(jobs).Error; err != nil {
		return err
	}
	args.JobIDs = nil
	for _, job := range *jobs {
		args.JobIDs = append(args.JobIDs, job.ID)
		j.jd.daemon.PopJob(job.ID)
	}

	ret := models.DB().Model(&models.DaemonJob{}).Where("id in(?)", args.JobIDs).Update("status", models.StatusJobStop)

	if ret.Error != nil {
		return ret.Error
	}

	return nil
}

func (j *DaemonJob) Delete(args proto.ActionJobsArgs, jobs *[]models.DaemonJob) error {

	model := models.DB()
	if args.Root {
		model = model.Where("id in(?) and status in (?)",
			args.JobIDs, []models.JobStatus{models.StatusJobRunning, models.StatusJobTiming})
	} else {
		model = model.Where("created_user_id = ? and id in(?) and status in (?)",
			args.UserID, args.JobIDs, []models.JobStatus{models.StatusJobRunning, models.StatusJobTiming})
	}

	if err := model.Find(jobs).Error; err != nil {
		return err
	}
	args.JobIDs = nil
	for _, job := range *jobs {
		args.JobIDs = append(args.JobIDs, job.ID)
		j.jd.daemon.PopJob(job.ID)
	}
	ret := models.DB().Delete(&models.DaemonJob{}, "id in (?)", args.JobIDs)

	if ret.Error != nil {
		return ret.Error
	}

	return nil
}

func (j *DaemonJob) Get(args proto.GetJobArgs, job *models.DaemonJob) error {
	return models.DB().Take(job, "id=?", args.JobID).Error
}

func (j *DaemonJob) Log(args proto.SearchLog, reply *proto.SearchLogResult) error {

	fd := finder.NewFinder(func(info os.FileInfo) bool {
		basename := filepath.Base(info.Name())
		arr := strings.Split(basename, ".")
		if len(arr) != 2 {
			return false
		}
		if arr[1] == "log" && arr[0] == fmt.Sprint(args.JobID) {
			return true
		}
		return false
	})

	if args.Date == "" {
		args.Date = time.Now().Format("2006/01/02")
	}

	if args.IsTail {
		fd.SetTail(true)
	}

	rootpath := filepath.Join(j.jd.getOpts().LogPath, "daemon_job", args.Date)
	err := fd.Search(rootpath, args.Pattern, &reply.Content, args.Offset, args.Pagesize)
	reply.Offset = fd.Offset()
	reply.FileSize = fd.FileSize()
	return err

}

func (j *DaemonJob) Audit(args proto.AuditJobArgs, jobs *[]models.DaemonJob) error {
	defer models.DB().Find(jobs, "id in (?)", args.JobIDs)
	return models.DB().Model(&models.DaemonJob{}).Where("id in (?) and status=?", args.JobIDs, models.StatusJobUnaudited).Update("status", models.StatusJobOk).Error
}
