package middleware

import (
	"net/url"
	"strings"

	"github.com/Unknwon/macaron"
	"github.com/macaron-contrib/csrf"

	"github.com/MessageDream/drift/modules/setting"
)

type DbOptions struct {
	Type, Host, Port, Name, User, Pwd, Path, LogMode string
}

func CreatDbConnect(options *DbOptions) macaron.Handler {
	return func(ctx *Context) {
		// Cannot view any page before installation.
		if !setting.InstallLock {
			ctx.Redirect("/install")
			return
		}

		// Redirect to dashboard if user tries to visit any non-login page.
		if options.SignOutRequire && ctx.IsSigned && ctx.Req.RequestURI != "/" {
			ctx.Redirect("/")
			return
		}

		if !options.SignOutRequire && !options.DisableCsrf && ctx.Req.Method == "POST" {
			csrf.Validate(ctx.Context, ctx.csrf)
			if ctx.Written() {
				return
			}
		}

		if options.SignInRequire {
			if !ctx.IsSigned {
				// Ignore watch repository operation.
				if strings.HasSuffix(ctx.Req.RequestURI, "watch") {
					return
				}
				ctx.SetCookie("redirect_to", "/"+url.QueryEscape(ctx.Req.RequestURI))
				ctx.Redirect("/user/login")
				return
			} else if !ctx.User.IsActive && setting.Service.RegisterEmailConfirm {
				ctx.Data["Title"] = ctx.Tr("auth.active_your_account")
				ctx.HTML(200, "user/auth/activate")
				return
			}
		}

		if options.AdminRequire {
			if !ctx.User.IsAdmin {
				ctx.Error(403)
				return
			}
			ctx.Data["PageIsAdmin"] = true
		}
	}
}
