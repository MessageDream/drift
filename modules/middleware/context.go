// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-macaron/cache"
	"github.com/go-macaron/i18n"
	"github.com/go-macaron/macaron"
	"github.com/go-macaron/session"
	mgo "gopkg.in/mgo.v2"

	"github.com/MessageDream/drift/models"
	"github.com/MessageDream/drift/modules/auth"
	"github.com/MessageDream/drift/modules/base"
	"github.com/MessageDream/drift/modules/log"
	"github.com/MessageDream/drift/modules/setting"
)

// Context represents context of a request.
type Context struct {
	*macaron.Context
	i18n.Locale
	Cache   cache.Cache
	Flash   *session.Flash
	Session session.Store

	User     *models.User
	IsSigned bool

	DbSession *mgo.Session
	Db        *mgo.Database
}

// Query querys form parameter.
func (ctx *Context) Query(name string) string {
	ctx.Req.ParseForm()
	return ctx.Req.Form.Get(name)
}

// HasError returns true if error occurs in form validation.
func (ctx *Context) HasApiError() bool {
	hasErr, ok := ctx.Data["HasError"]
	if !ok {
		return false
	}
	return hasErr.(bool)
}

func (ctx *Context) GetErrMsg() string {
	return ctx.Data["ErrorMsg"].(string)
}

// HasError returns true if error occurs in form validation.
func (ctx *Context) HasError() bool {
	hasErr, ok := ctx.Data["HasError"]
	if !ok {
		return false
	}
	ctx.Flash.ErrorMsg = ctx.Data["ErrorMsg"].(string)
	ctx.Data["Flash"] = ctx.Flash
	return hasErr.(bool)
}

// HTML calls Context.HTML and converts template name to string.
func (ctx *Context) HTML(status int, name base.TplName) {
	ctx.Context.HTML(status, string(name))
}

// RenderWithErr used for page has form validation but need to prompt error to users.
func (ctx *Context) RenderWithErr(msg string, tpl base.TplName, form interface{}) {
	if form != nil {
		auth.AssignForm(form, ctx.Data)
	}
	ctx.Flash.ErrorMsg = msg
	ctx.Data["Flash"] = ctx.Flash
	ctx.HTML(200, tpl)
}

// Handle handles and logs error by given status.
func (ctx *Context) Handle(status int, title string, err error) {
	if err != nil {
		log.Error(4, "%s: %v", title, err)
		if macaron.Env != macaron.PROD {
			ctx.Data["ErrorMsg"] = err
		}
	}

	switch status {
	case 404:
		ctx.Data["Title"] = "Page Not Found"
	case 500:
		ctx.Data["Title"] = "Internal Server Error"
	}
	ctx.HTML(status, base.TplName(fmt.Sprintf("status/%d", status)))
}

func (ctx *Context) ServeFile(file string, names ...string) {
	var name string
	if len(names) > 0 {
		name = names[0]
	} else {
		name = path.Base(file)
	}
	ctx.Resp.Header().Set("Content-Description", "File Transfer")
	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Content-Disposition", "attachment; filename="+name)
	ctx.Resp.Header().Set("Content-Transfer-Encoding", "binary")
	ctx.Resp.Header().Set("Expires", "0")
	ctx.Resp.Header().Set("Cache-Control", "must-revalidate")
	ctx.Resp.Header().Set("Pragma", "public")
	http.ServeFile(ctx.Resp, ctx.Req, file)
}

func (ctx *Context) ServeContent(name string, r io.ReadSeeker, params ...interface{}) {
	modtime := time.Now()
	for _, p := range params {
		switch v := p.(type) {
		case time.Time:
			modtime = v
		}
	}
	ctx.Resp.Header().Set("Content-Description", "File Transfer")
	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Content-Disposition", "attachment; filename="+name)
	ctx.Resp.Header().Set("Content-Transfer-Encoding", "binary")
	ctx.Resp.Header().Set("Expires", "0")
	ctx.Resp.Header().Set("Cache-Control", "must-revalidate")
	ctx.Resp.Header().Set("Pragma", "public")
	http.ServeContent(ctx.Resp, ctx.Req, name, modtime, r)
}

// Contexter initializes a classic context for a request.
func Contexter() macaron.Handler {
	return func(c *macaron.Context, l i18n.Locale, cache cache.Cache, sess session.Store, f *session.Flash, dbSession *mgo.Session, db *mgo.Database) {

		ctx := &Context{
			Context:   c,
			Locale:    l,
			Cache:     cache,
			Flash:     f,
			Session:   sess,
			DbSession: dbSession,
			Db:        db,
		}
		//		dbSession, err := models.GetSession()
		//		if err != nil {
		//			log.Error("Get DbSession error:%v", err)
		//			ctx.Handle(500, "Get DbSession", err)
		//			return
		//		}
		//		ctx.DbSession = dbSession
		//		ctx.Db = models.GetDb(dbSession)
		// Compute current URL for real-time change language.
		link := ctx.Req.RequestURI
		i := strings.Index(link, "?")
		if i > -1 {
			link = link[:i]
		}
		ctx.Data["Link"] = link

		ctx.Data["PageStartTime"] = time.Now()

		// Get user from session if logined.
		ctx.User = auth.SignedInUser(ctx.Req.Header, ctx.Session)
		if ctx.User != nil {
			ctx.IsSigned = true
			ctx.Data["IsSigned"] = ctx.IsSigned
			ctx.Data["SignedUser"] = ctx.User
			ctx.Data["IsAdmin"] = ctx.User.IsAdmin
		}

		// If request sends files, parse them here otherwise the Query() can't be parsed and the CsrfToken will be invalid.
		if ctx.Req.Method == "POST" && strings.Contains(ctx.Req.Header.Get("Content-Type"), "multipart/form-data") {
			if err := ctx.Req.ParseMultipartForm(setting.AttachmentMaxSize << 20); err != nil && !strings.Contains(err.Error(), "EOF") { // 32MB max size
				ctx.Handle(500, "ParseMultipartForm", err)
				return
			}
		}
		c.Map(ctx)
	}
}
