package admin

import (
	"jiacrontab/models"
	"jiacrontab/pkg/mailer"
	"jiacrontab/pkg/rpc"

	"github.com/kataras/iris/middleware/logger"

	"github.com/kataras/iris"
)

type Admin struct {
}

func New() *Admin {
	return &Admin{}
}

func (a *Admin) init() {

	models.CreateDB(cfg.Database.DriverName, cfg.Database.DSN)

	models.DB().CreateTable(&models.Node{})
	models.DB().AutoMigrate(&models.Node{})

	models.DB().CreateTable(&models.Group{})
	models.DB().AutoMigrate(&models.Group{})

	models.DB().CreateTable(&models.User{})
	models.DB().AutoMigrate(&models.User{})

	models.DB().CreateTable(&models.Event{})
	models.DB().AutoMigrate(&models.Event{})

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

	a.init()

	app := iris.New()
	// html := iris.HTML(cfg.App.TplDir, cfg.App.TplExt)
	// html.Reload(true)
	// app.RegisterView(html)
	app.Logger().SetLevel("debug")
	app.Use(logger.New())

	route(app)
	go rpc.ListenAndServe(cfg.App.RPCListenAddr, &Srv{})

	app.Run(iris.Addr(cfg.App.HTTPListenAddr))
}
