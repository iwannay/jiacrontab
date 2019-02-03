package admin

import (
	"io"
	"io/ioutil"
	"jiacrontab/models"
	"jiacrontab/pkg/log"
	"jiacrontab/pkg/mailer"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"
	"jiacrontab/server/conf"
	"net/http"
	"strings"
	"time"
)

type Srv struct{}

func (s *Srv) Register(args models.Node, reply *proto.MailArgs) error {

	*reply = proto.MailArgs{
		Host: conf.MailService.Host,
		User: conf.MailService.User,
		Pass: conf.MailService.Passwd,
	}

	ret := models.DB().Model(&models.Node{}).Where("addr=?", args.Addr).Update(map[string]interface{}{})

	if ret.RowsAffected == 0 {
		args.Name = time.Now().Format("20060102150405")
		ret = models.DB().Create(&args)
	}

	return ret.Error
}

func (s *Srv) Depends(args proto.DependsTasks, reply *bool) error {
	log.Infof("Callee Logic.Depend taskId %d", args[0].JobEntryID)
	*reply = true
	for _, v := range args {
		if err := rpc.Call(v.Dest, "CrontabTask.ExecDepend", v, &reply); err != nil {
			*reply = false
			return err
		}
	}

	return nil
}

func (s *Srv) DependDone(args proto.DependsTask, reply *bool) error {
	log.Infof("Callee Logic.DependDone task %s", args.Name)
	*reply = true
	if err := rpc.Call(args.Dest, "CrontabTask.ResolvedDepends", args, &reply); err != nil {
		*reply = false
		return err
	}

	return nil
}

func (s *Srv) SendMail(args proto.SendMail, reply *bool) error {
	var err error
	if conf.MailService.Enabled {
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
