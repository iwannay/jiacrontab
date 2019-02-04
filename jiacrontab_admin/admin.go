package admin

import (
	"jiacrontab/models"
	"jiacrontab/pkg/mailer"
	"jiacrontab/pkg/rpc"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kataras/iris"
)

type Admin struct {
}

func New() *Admin {
	return &Admin{}
}

func (a *Admin) init() {
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
	models.CreateDB("sqlite3", "data/jiacrontab_admin.db")

	models.DB().CreateTable(&models.Node{})
	models.DB().AutoMigrate(&models.Node{})

	models.DB().CreateTable(&models.Group{})
	models.DB().AutoMigrate(&models.Group{})

	models.DB().CreateTable(&models.User{})
	models.DB().AutoMigrate(&models.User{})

	a.init()

	app := iris.New()
	// html := iris.HTML(cfg.App.TplDir, cfg.App.TplExt)
	// html.Reload(true)
	// app.RegisterView(html)

	route(app)
	go rpc.ListenAndServe(cfg.App.RpcListenAddr, &Srv{})
	app.Run(iris.Addr(cfg.App.HttpListenAddr))
}
