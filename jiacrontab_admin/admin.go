package admin

import (
	"jiacrontab/models"
	"jiacrontab/pkg/mailer"
	"jiacrontab/pkg/rpc"

	"github.com/kataras/iris"
)

type Admin struct {
}

func New() *Admin {
	return &Admin{}
}

func (a *Admin) init() {
	if err := models.InitModel(cfg.Database.DriverName, cfg.Database.DSN); err != nil {
		panic(err)
	}
	// mail
	if cfg.Mailer.Enabled {
		mailer.InitMailer(&mailer.Mailer{
			QueueLength:    cfg.Mailer.QueueLength,
			SubjectPrefix:  cfg.Mailer.SubjectPrefix,
			From:           cfg.Mailer.From,
			Host:           cfg.Mailer.Host,
			User:           cfg.Mailer.User,
			Passwd:         cfg.Mailer.Passwd,
			FromEmail:      cfg.Mailer.FromEmail,
			DisableHelo:    cfg.Mailer.DisableHelo,
			HeloHostname:   cfg.Mailer.HeloHostname,
			SkipVerify:     cfg.Mailer.SkipVerify,
			UseCertificate: cfg.Mailer.UseCertificate,
			CertFile:       cfg.Mailer.CertFile,
			KeyFile:        cfg.Mailer.KeyFile,
			UsePlainText:   cfg.Mailer.UsePlainText,
			HookMode:       false,
		})
	}
}

func (a *Admin) Main() {

	var initModel bool
	if cfg.Database.DriverName != "" && cfg.Database.DSN != "" {
		initModel = true
	}

	if initModel {
		a.init()
		defer models.DB().Close()
		go rpc.ListenAndServe(cfg.App.RPCListenAddr, NewSrv(a))
	}

	app := newApp(initModel)
	app.Run(iris.Addr(cfg.App.HTTPListenAddr))
}
