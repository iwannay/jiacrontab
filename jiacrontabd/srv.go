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
	model := models.DB().Model(&models.CrontabJob{})
	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("name like ?", "%"+args.SearchTxt+"%")
	} else if args.Root {
		model = model.Where("name like ? and group_id=?", "%"+args.SearchTxt+"%", args.GroupID)
	} else {
		model = model.Where("name like ? and created_user_id=? and group_id=?", "%"+args.SearchTxt+"%", args.UserID, args.GroupID)
	}

	err := model.Count(&reply.Total).Error
	if err != nil {
		return err
	}

	reply.Page = args.Page
	reply.Pagesize = args.Pagesize

	return model.Order(gorm.Expr("created_user_id=? desc, id desc", args.UserID)).Offset((args.Page - 1) * args.Pagesize).Limit(args.Pagesize).Find(&reply.List).Error
}

func (j *CrontabJob) Audit(args proto.AuditJobArgs, reply *[]models.CrontabJob) error {
	model := models.DB().Model(&models.CrontabJob{})

	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("id in (?)", args.JobIDs)
	} else {
		model = model.Where("id in (?) and group_id=?", args.JobIDs, args.GroupID)
	}

	defer model.Find(reply)
	return model.Where("status=?", models.StatusJobUnaudited).Update("status", models.StatusJobOk).Error
}

func (j *CrontabJob) Edit(args proto.EditCrontabJobArgs, reply *models.CrontabJob) error {

	var (
		model = models.DB()
	)

	if args.Job.MaxConcurrent == 0 {
		args.Job.MaxConcurrent = 1
	}

	if args.Job.ID == 0 {
		model = models.DB().Save(&args.Job)
	} else {
		if args.GroupID == models.SuperGroup.ID {
			model = model.Where("id=?", args.Job.ID)
		} else if args.Root {
			model = model.Where("id=? and group_id=?", args.Job.ID, args.Job.GroupID)
		} else {
			model = model.Where("id=? and created_user_id=? and group_id=?", args.Job.ID, args.Job.CreatedUserID, args.Job.GroupID)
		}
		args.Job.NextExecTime = time.Time{}
		model = model.Omit(
			"updated_at", "created_at", "deleted_at",
			"created_user_id", "created_username",
			"last_cost_time", "last_exec_time",
			"last_exit_status", "process_num",
		).Save(&args.Job)
	}
	*reply = args.Job
	return model.Error
}
func (j *CrontabJob) Get(args proto.GetJobArgs, reply *models.CrontabJob) error {
	model := models.DB()
	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("id=?", args.JobID)
	} else if args.Root {
		model = model.Where("id=? and group_id=?", args.JobID, args.GroupID)
	} else {
		model = model.Where("id=? and created_user_id=? and group_id=?", args.JobID, args.UserID, args.GroupID)
	}
	return model.Find(reply).Error
}

func (j *CrontabJob) Start(args proto.ActionJobsArgs, jobs *[]models.CrontabJob) error {

	model := models.DB()

	if len(args.JobIDs) == 0 {
		return errors.New("empty ids")
	}

	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("id in (?) and status in (?)",
			args.JobIDs, []models.JobStatus{models.StatusJobOk, models.StatusJobStop})
	} else if args.Root {
		model = model.Where("id in (?) and status in (?) and group_id=?",
			args.JobIDs, []models.JobStatus{models.StatusJobOk, models.StatusJobStop}, args.GroupID)
	} else {
		model = model.Where("created_user_id = ? and id in (?) and status in (?) and group_id=?",
			args.UserID, args.JobIDs, []models.JobStatus{models.StatusJobOk, models.StatusJobStop}, args.GroupID)
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
	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("id in (?) and status in (?)", args.JobIDs, []models.JobStatus{models.StatusJobTiming, models.StatusJobRunning})
	} else if args.Root {
		model = model.Where(" id in (?) and status in (?) and group_id=?",
			args.JobIDs, []models.JobStatus{models.StatusJobTiming, models.StatusJobRunning}, args.GroupID)
	} else {
		model = model.Where("created_user_id = ? and id in (?) and status in (?)  and group_id=?",
			args.UserID, args.JobIDs, []models.JobStatus{models.StatusJobTiming, models.StatusJobRunning}, args.GroupID)
	}

	for _, jobID := range args.JobIDs {
		j.jd.killTask(jobID)
	}

	return model.Find(job).Update(map[string]interface{}{
		"status":         models.StatusJobStop,
		"next_exec_time": time.Time{},
	}).Error
}

func (j *CrontabJob) Delete(args proto.ActionJobsArgs, job *[]models.CrontabJob) error {
	model := models.DB()
	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("id in (?)", args.JobIDs)
	} else if args.Root {
		model = model.Where("id in (?) and group_id=?", args.JobIDs, args.GroupID)
	} else {
		model = model.Where("created_user_id = ? and id in (?) and group_id=?",
			args.UserID, args.JobIDs, args.GroupID)
	}
	return model.Find(job).Delete(&models.CrontabJob{}).Error
}

func (j *CrontabJob) Kill(args proto.ActionJobsArgs, job *[]models.CrontabJob) error {
	model := models.DB()
	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("id in (?)", args.JobIDs)
	} else if args.Root {
		model = model.Where("id in (?) and group_id=?", args.JobIDs, args.GroupID)
	} else {
		model = model.Where("created_user_id = ? and id in (?) and group_id=?",
			args.UserID, args.JobIDs, args.GroupID)
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
	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("id=?", args.JobID)
	} else if args.Root {
		model = model.Where("id=? and group_id=?", args.JobID, args.GroupID)
	} else {
		model = model.Where("created_user_id = ? and id=? and group_id=?", args.UserID, args.JobID, args.GroupID)
	}

	ret := model.Take(&reply.Job)

	if ret.Error == nil {
		jobInstance := newJobEntry(&crontab.Job{
			ID:    reply.Job.ID,
			Value: reply.Job,
		}, j.jd)
		jobInstance.setOnce(true)
		j.jd.addTmpJob(jobInstance)
		defer j.jd.removeTmpJob(jobInstance)
		jobInstance.once = true
		jobInstance.exec()
		reply.Content = jobInstance.GetLog()
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
		id:          args.ID,
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

	model := models.DB().Model(&models.DaemonJob{})

	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("name like ?", "%"+args.SearchTxt+"%")
	} else if args.Root {
		model = model.Where("name like ? and group_id=?", "%"+args.SearchTxt+"%", args.GroupID)
	} else {
		model = model.Where("name like ? and created_user_id=? and group_id=?", "%"+args.SearchTxt+"%", args.UserID, args.GroupID)
	}

	err := model.Count(&reply.Total).Error
	if err != nil {
		return err
	}

	reply.Page = args.Page
	reply.Pagesize = args.Pagesize

	return model.Order(gorm.Expr("created_user_id=? desc, id desc", args.UserID)).Offset((args.Page - 1) * args.Pagesize).Limit(args.Pagesize).Find(&reply.List).Error
}

func (j *DaemonJob) Edit(args proto.EditDaemonJobArgs, job *models.DaemonJob) error {

	model := models.DB()
	if args.Job.ID == 0 {
		model = models.DB().Create(&args.Job)
	} else {
		if args.GroupID == models.SuperGroup.ID {
			model = model.Where("id=?", args.Job.ID)
		} else if args.Root {
			model = model.Where("id=? and group_id=?", args.Job.ID, args.GroupID)
		} else {
			model = model.Where("id=? and created_user_id=? and group_id=?", args.Job.ID, args.Job.CreatedUserID, args.GroupID)
		}
		model = model.Omit(
			"updated_at", "created_at", "deleted_at",
			"created_user_id", "created_username", "start_at").Save(&args.Job)
	}

	*job = args.Job
	return model.Error
}

func (j *DaemonJob) Start(args proto.ActionJobsArgs, jobs *[]models.DaemonJob) error {

	model := models.DB()
	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("id in(?) and status in (?)",
			args.JobIDs, []models.JobStatus{models.StatusJobOk, models.StatusJobStop})
	} else if args.Root {
		model = model.Where(" and id in(?) and status in (?) and group_id=?",
			args.JobIDs, []models.JobStatus{models.StatusJobOk, models.StatusJobStop}, args.GroupID)
	} else {
		model = model.Where("created_user_id = ? and id in(?) and status in (?) and group_id=?",
			args.UserID, args.JobIDs, []models.JobStatus{models.StatusJobOk, models.StatusJobStop}, args.GroupID)
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
	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("id in(?) and status in (?)",
			args.JobIDs, []models.JobStatus{models.StatusJobRunning, models.StatusJobTiming})
	} else if args.Root {
		model = model.Where("id in(?) and status in (?) and group_id=?",
			args.JobIDs, []models.JobStatus{models.StatusJobRunning, models.StatusJobTiming}, args.GroupID)
	} else {
		model = model.Where("created_user_id = ? and id in(?) and status in (?) and group_id=?",
			args.UserID, args.JobIDs, []models.JobStatus{models.StatusJobRunning, models.StatusJobTiming}, args.GroupID)
	}

	if err := model.Find(jobs).Error; err != nil {
		return err
	}
	args.JobIDs = nil
	for _, job := range *jobs {
		args.JobIDs = append(args.JobIDs, job.ID)
		j.jd.daemon.PopJob(job.ID)
	}

	return model.Model(&models.DaemonJob{}).Update("status", models.StatusJobStop).Error
}

func (j *DaemonJob) Delete(args proto.ActionJobsArgs, jobs *[]models.DaemonJob) error {

	model := models.DB()
	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("id in(?) and status in (?)",
			args.JobIDs, []models.JobStatus{models.StatusJobRunning, models.StatusJobTiming})
	} else if args.Root {
		model = model.Where("id in(?) and status in (?) and group_id=?",
			args.JobIDs, []models.JobStatus{models.StatusJobRunning, models.StatusJobTiming}, args.GroupID)
	} else {
		model = model.Where("created_user_id = ? and id in(?) and status in (?) and group_id=?",
			args.UserID, args.JobIDs, []models.JobStatus{models.StatusJobRunning, models.StatusJobTiming}, args.GroupID)
	}

	if err := model.Find(jobs).Error; err != nil {
		return err
	}
	args.JobIDs = nil
	for _, job := range *jobs {
		args.JobIDs = append(args.JobIDs, job.ID)
		j.jd.daemon.PopJob(job.ID)
	}
	return model.Delete(&models.DaemonJob{}).Error
}

func (j *DaemonJob) Get(args proto.GetJobArgs, job *models.DaemonJob) error {
	model := models.DB()
	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("id=?", args.JobID)
	} else if args.Root {
		model = model.Where("id=? and group_id=?", args.JobID, args.GroupID)
	} else {
		model = model.Where("id=? and group_id=? and created_user_id=?", args.JobID, args.GroupID, args.UserID)
	}
	return model.Take(job).Error
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
	model := models.DB().Model(&models.DaemonJob{})
	if args.GroupID == models.SuperGroup.ID {
		model = model.Where("id in (?)", args.JobIDs)
	} else if args.Root {
		model = model.Where("id in (?) and group_id=?", args.JobIDs, args.GroupID)
	} else {
		model = model.Where("id in (?) and group_id=? and created_user_id=?", args.JobIDs, args.GroupID, args.UserID)
	}
	return model.Where("status=?", models.StatusJobUnaudited).Find(jobs).Update("status", models.StatusJobOk).Error
}
