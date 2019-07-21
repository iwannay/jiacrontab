package admin

import (
	"fmt"
	"io"
	"io/ioutil"
	"jiacrontab/models"
	"jiacrontab/pkg/mailer"
	"jiacrontab/pkg/proto"
	"net/http"
	"strings"
	"time"

	"github.com/iwannay/log"
)

type Srv struct {
	adm *Admin
}

func NewSrv(adm *Admin) *Srv {
	return &Srv{
		adm: adm,
	}
}

func (s *Srv) Register(args map[uint]models.Node, reply *bool) error {
	*reply = true

	for _, node := range args {
		ret := models.DB().Unscoped().Model(&models.Node{}).Where("addr=? and group_id=?", node.Addr, node.GroupID).Updates(map[string]interface{}{
			"daemon_task_num":       node.DaemonTaskNum,
			"crontab_task_num":      node.CrontabTaskNum,
			"name":                  node.Name,
			"crontab_job_audit_num": node.CrontabJobAuditNum,
			"daemon_job_audit_Num":  node.DaemonJobAuditNum,
			"crontab_job_fail_num":  node.CrontabJobFailNum,
			"deleted_at":            nil,
			"disabled":              false,
		})
		if ret.RowsAffected == 0 && node.GroupID == models.SuperGroup.ID {
			ret = models.DB().Create(&node)
		}

		if ret.Error != nil {
			return ret.Error
		}
	}

	return nil
}

func (s *Srv) ExecDepend(args proto.DepJobs, reply *bool) error {
	log.Infof("Callee Srv.ExecDepend jobID:%d", args[0].JobID)
	*reply = true
	for _, v := range args {
		if err := rpcCall(v.Dest, "CrontabJob.ExecDepend", v, &reply); err != nil {
			*reply = false
			return err
		}
	}

	return nil
}

func (s *Srv) SetDependDone(args proto.DepJob, reply *bool) error {
	log.Infof("Callee Srv.SetDependDone jobID:%d", args.JobID)
	*reply = true
	if err := rpcCall(args.Dest, "CrontabJob.SetDependDone", args, &reply); err != nil {
		*reply = false
		return err
	}

	return nil
}

func (s *Srv) SendMail(args proto.SendMail, reply *bool) error {
	var (
		err error
		cfg = s.adm.getOpts()
	)
	if cfg.Mailer.Enabled {
		err = mailer.SendMail(args.MailTo, args.Subject, args.Content)
	}
	*reply = true
	return err
}

func (s *Srv) PushJobLog(args models.JobHistory, reply *bool) error {
	models.PushJobHistory(&args)
	*reply = true
	return nil
}

func (s *Srv) ApiPost(args proto.ApiPost, reply *bool) error {
	var (
		err  error
		errs []error
	)

	for _, url := range args.Urls {

		client := http.Client{
			Timeout: time.Minute,
		}

		response, err := client.Post(url, "application/json", strings.NewReader(args.Data))

		if err != nil {
			errs = append(errs, err)
			log.Errorf("post url %s fail: %s", url, err)
			continue
		}
		defer response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)
	}

	for _, v := range errs {
		if err != nil {
			err = fmt.Errorf("%s\n%s", err, v)
		} else {
			err = v
		}
	}

	*reply = true
	return err
}

func (s *Srv) Ping(args *proto.EmptyArgs, reply *proto.EmptyReply) error {
	return nil
}
