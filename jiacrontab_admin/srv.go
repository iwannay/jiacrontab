package admin

import (
	"io"
	"io/ioutil"
	"jiacrontab/models"
	"jiacrontab/pkg/log"
	"jiacrontab/pkg/mailer"
	"jiacrontab/pkg/proto"
	"net/http"
	"strings"
)

type Srv struct{}

func (s *Srv) Register(args models.Node, reply *bool) error {
	*reply = true
	cfg.SetUsed()
	ret := models.DB().Model(&models.Node{}).Where("addr=? and group_id=?", args.Addr, args.GroupID).Updates(map[string]interface{}{
		"name":             args.Name,
		"daemon_task_num":  args.DaemonTaskNum,
		"crontab_task_num": args.CrontabTaskNum,
		"addr":             args.Addr,
	})
	if ret.RowsAffected == 0 {
		ret = models.DB().Create(&args)
	}

	return ret.Error
}

func (s *Srv) Depend(args proto.DepJobs, reply *bool) error {
	log.Infof("Callee Srv.Depend jobID:%d", args[0].JobID)
	*reply = true
	for _, v := range args {
		if err := rpcCall(v.Dest, "CrontabJob.ExecDepend", v, &reply); err != nil {
			*reply = false
			return err
		}
	}

	return nil
}

func (s *Srv) DependDone(args proto.DepJob, reply *bool) error {
	log.Infof("Callee Srv.DependDone jobID:%d", args.JobID)
	*reply = true
	if err := rpcCall(args.Dest, "CrontabJob.ResolvedDepend", args, &reply); err != nil {
		*reply = false
		return err
	}

	return nil
}

func (s *Srv) SendMail(args proto.SendMail, reply *bool) error {
	var err error
	if cfg.Mailer.Enabled {
		err = mailer.SendMail(args.MailTo, args.Subject, args.Content)
	}
	*reply = true
	return err
}

func (s *Srv) ApiPost(args proto.ApiPost, reply *bool) error {
	req, err := http.NewRequest("POST", args.Url, strings.NewReader(args.Data))
	if err != nil {
		log.Errorf("create req fail: %s", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Errorf("post url %s fail: %s", args.Url, err)
		return err
	}

	defer response.Body.Close()
	io.Copy(ioutil.Discard, response.Body)
	*reply = true
	return nil
}

func (s *Srv) Ping(args *proto.EmptyArgs, reply *proto.EmptyReply) error {
	return nil
}
