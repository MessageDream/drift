// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/codegangsta/cli"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/macaron"
	"github.com/go-macaron/session"
	"github.com/go-macaron/toolbox"

	"github.com/MessageDream/drift/models"
	"github.com/MessageDream/drift/modules/auth"
	"github.com/MessageDream/drift/modules/avatar"
	"github.com/MessageDream/drift/modules/base"
	"github.com/MessageDream/drift/modules/log"
	"github.com/MessageDream/drift/modules/middleware"
	"github.com/MessageDream/drift/modules/setting"
	"github.com/MessageDream/drift/routers"
	"github.com/MessageDream/drift/routers/admin"
	"github.com/MessageDream/drift/routers/dev"
)

var CmdWeb = cli.Command{
	Name:  "web",
	Usage: "Start Gogs web server",
	Description: `Gogs web server is the only thing you need to run,
and it takes care of all the other things for you`,
	Action: runWeb,
	Flags:  []cli.Flag{},
}

// checkVersion checks if binary matches the version of templates files.
func checkVersion() {
	data, err := ioutil.ReadFile(path.Join(setting.StaticRootPath, "templates/.VERSION"))
	if err != nil {
		log.Fatal(4, "Fail to read 'templates/.VERSION': %v", err)
	}
	if string(data) != setting.AppVer {
		log.Fatal(4, "Binary and template file version does not match, did you forget to recompile?")
	}
}

// newMacaron initializes Macaron instance.
func newMacaron() *macaron.Macaron {
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Static("public",
		macaron.StaticOptions{
			SkipLogging: !setting.DisableRouterLog,
		},
	))
	// if setting.EnableGzip {
	// 	m.Use(macaron.Gzip())
	// }
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Directory:  path.Join(setting.StaticRootPath, "templates"),
		Funcs:      []template.FuncMap{base.TemplateFuncs},
		IndentJSON: macaron.Env != macaron.PROD,
	}))
	m.Use(cache.Cacher(cache.Options{
		Adapter:  setting.CacheAdapter,
		Interval: setting.CacheInternal,
		Conn:     setting.CacheConn,
	}))
	m.Use(session.Sessioner(session.Options{
		Provider: setting.SessionProvider,
		Config:   *setting.SessionConfig,
	}))
	m.Use(toolbox.Toolboxer(m, toolbox.Options{
		HealthCheckFuncs: []*toolbox.HealthCheckFuncDesc{
			&toolbox.HealthCheckFuncDesc{
				Desc: "Database connection",
				Func: models.Ping,
			},
		},
	}))
	m.Use(middleware.Contexter())
	return m
}

func runWeb(*cli.Context) {
	routers.GlobalInit()
	checkVersion()

	m := newMacaron()

	reqSignIn := middleware.Toggle(&middleware.ToggleOptions{SignInRequire: true})
	ignSignIn := middleware.Toggle(&middleware.ToggleOptions{SignInRequire: setting.Service.RequireSignInView})
	reqSignOut := middleware.Toggle(&middleware.ToggleOptions{SignOutRequire: true})

	bindIgnErr := binding.BindIgnErr

	// Routers.
	m.Get("/", ignSignIn, routers.Home)
	m.Get("/explore", routers.Explore)
	m.Get("/install", bindIgnErr(auth.InstallForm{}), routers.Install)
	m.Post("/install", bindIgnErr(auth.InstallForm{}), routers.InstallPost)
	m.Group("", func(r *macaron.Router) {
		r.Get("/pulls", user.Pulls)
		r.Get("/issues", user.Issues)
	}, reqSignIn)

	// User routers.
	m.Group("/user", func(r *macaron.Router) {
		r.Get("/login", user.SignIn)
		r.Post("/login", bindIgnErr(auth.SignInForm{}), user.SignInPost)
		r.Get("/login/:name", user.SocialSignIn)
		r.Get("/sign_up", user.SignUp)
		r.Post("/sign_up", bindIgnErr(auth.RegisterForm{}), user.SignUpPost)
		r.Get("/reset_password", user.ResetPasswd)
		r.Post("/reset_password", user.ResetPasswdPost)
	}, reqSignOut)
	m.Group("/user/settings", func(r *macaron.Router) {
		r.Get("", user.Settings)
		r.Post("", bindIgnErr(auth.UpdateProfileForm{}), user.SettingsPost)
		r.Get("/password", user.SettingsPassword)
		r.Post("/password", bindIgnErr(auth.ChangePasswordForm{}), user.SettingsPasswordPost)
		r.Get("/ssh", user.SettingsSSHKeys)
		r.Post("/ssh", bindIgnErr(auth.AddSSHKeyForm{}), user.SettingsSSHKeysPost)
		r.Get("/social", user.SettingsSocial)
		r.Route("/delete", "GET,POST", user.SettingsDelete)
	}, reqSignIn)
	m.Group("/user", func(r *macaron.Router) {
		// r.Get("/feeds", binding.Bind(auth.FeedsForm{}), user.Feeds)
		r.Any("/activate", user.Activate)
		r.Get("/email2user", user.Email2User)
		r.Get("/forget_password", user.ForgotPasswd)
		r.Post("/forget_password", user.ForgotPasswdPost)
		r.Get("/logout", user.SignOut)
	})

	m.Get("/user/:username", ignSignIn, user.Profile) // TODO: Legacy

	// Gravatar service.
	avt := avatar.CacheServer("public/img/avatar/", "public/img/avatar_default.jpg")
	os.MkdirAll("public/img/avatar/", os.ModePerm)
	m.Get("/avatar/:hash", avt.ServeHTTP)

	adminReq := middleware.Toggle(&middleware.ToggleOptions{SignInRequire: true, AdminRequire: true})

	m.Group("/admin", func(r *macaron.Router) {
		m.Get("", adminReq, admin.Dashboard)
		r.Get("/config", admin.Config)
		r.Get("/monitor", admin.Monitor)

		m.Group("/users", func(r *macaron.Router) {
			r.Get("", admin.Users)
			r.Get("/new", admin.NewUser)
			r.Post("/new", bindIgnErr(auth.RegisterForm{}), admin.NewUserPost)
			r.Get("/:userid", admin.EditUser)
			r.Post("/:userid", bindIgnErr(auth.AdminEditUserForm{}), admin.EditUserPost)
			r.Post("/:userid/delete", admin.DeleteUser)
		})

		m.Group("/auths", func(r *macaron.Router) {
			r.Get("", admin.Authentications)
			r.Get("/new", admin.NewAuthSource)
			r.Post("/new", bindIgnErr(auth.AuthenticationForm{}), admin.NewAuthSourcePost)
			r.Get("/:authid", admin.EditAuthSource)
			r.Post("/:authid", bindIgnErr(auth.AuthenticationForm{}), admin.EditAuthSourcePost)
			r.Post("/:authid/delete", admin.DeleteAuthSource)
		})
	}, adminReq)

	m.Get("/:username", ignSignIn, user.Profile)

	if macaron.Env == macaron.DEV {
		m.Get("/template/*", dev.TemplatePreview)
	}

	// Not found handler.
	m.NotFound(routers.NotFound)

	var err error
	listenAddr := fmt.Sprintf("%s:%s", setting.HttpAddr, setting.HttpPort)
	log.Info("Listen: %v://%s", setting.Protocol, listenAddr)
	switch setting.Protocol {
	case setting.HTTP:
		err = http.ListenAndServe(listenAddr, m)
	case setting.HTTPS:
		err = http.ListenAndServeTLS(listenAddr, setting.CertFile, setting.KeyFile, m)
	default:
		log.Fatal(4, "Invalid protocol: %s", setting.Protocol)
	}

	if err != nil {
		log.Fatal(4, "Fail to start server: %v", err)
	}
}
