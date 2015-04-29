// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	"strings"

	"github.com/Unknwon/com"

	"github.com/MessageDream/drift/models"
	"github.com/MessageDream/drift/modules/auth"
	"github.com/MessageDream/drift/modules/base"
	"github.com/MessageDream/drift/modules/log"
	"github.com/MessageDream/drift/modules/middleware"
)

const (
	SETTINGS_PROFILE  base.TplName = "user/settings/profile"
	SETTINGS_PASSWORD base.TplName = "user/settings/password"
	SETTINGS_SSH_KEYS base.TplName = "user/settings/sshkeys"
	SETTINGS_SOCIAL   base.TplName = "user/settings/social"
	SETTINGS_DELETE   base.TplName = "user/settings/delete"
	NOTIFICATION      base.TplName = "user/notification"
	SECURITY          base.TplName = "user/security"
)

func Settings(ctx *middleware.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsUserSettings"] = true
	ctx.Data["PageIsSettingsProfile"] = true
	ctx.HTML(200, SETTINGS_PROFILE)
}

func SettingsPost(ctx *middleware.Context, form auth.UpdateProfileForm) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsUserSettings"] = true
	ctx.Data["PageIsSettingsProfile"] = true

	if ctx.HasError() {
		ctx.HTML(200, SETTINGS_PROFILE)
		return
	}

	// Check if user name has been changed.
	if ctx.User.Name != form.UserName {
		isExist, err := models.IsUserExist(form.UserName)
		if err != nil {
			ctx.Handle(500, "IsUserExist", err)
			return
		} else if isExist {
			ctx.RenderWithErr(ctx.Tr("form.username_been_taken"), SETTINGS_PROFILE, &form)
			return
		} else if err = models.ChangeUserName(ctx.User, form.UserName); err != nil {
			if err == models.ErrUserNameIllegal {
				ctx.Flash.Error(ctx.Tr("form.illegal_username"))
				ctx.Redirect("/user/settings")
				return
			} else {
				ctx.Handle(500, "ChangeUserName", err)
			}
			return
		}
		log.Trace("User name changed: %s -> %s", ctx.User.Name, form.UserName)
		ctx.User.Name = form.UserName
	}

	ctx.User.FullName = form.FullName
	ctx.User.Email = form.Email
	ctx.User.Website = form.Website
	ctx.User.Location = form.Location
	ctx.User.Avatar = base.EncodeMd5(form.Avatar)
	ctx.User.AvatarEmail = form.Avatar
	if err := models.UpdateUser(ctx.User); err != nil {
		ctx.Handle(500, "UpdateUser", err)
		return
	}
	log.Trace("User setting updated: %s", ctx.User.Name)
	ctx.Flash.Success(ctx.Tr("settings.update_profile_success"))
	ctx.Redirect("/user/settings")
}

func SettingsPassword(ctx *middleware.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsUserSettings"] = true
	ctx.Data["PageIsSettingsPassword"] = true
	ctx.HTML(200, SETTINGS_PASSWORD)
}

func SettingsPasswordPost(ctx *middleware.Context, form auth.ChangePasswordForm) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsUserSettings"] = true
	ctx.Data["PageIsSettingsPassword"] = true

	if ctx.HasError() {
		ctx.HTML(200, SETTINGS_PASSWORD)
		return
	}

	tmpUser := &models.User{
		Password: form.OldPassword,
		Salt:   ctx.User.Salt,
	}
	tmpUser.EncodePasswd()
	if ctx.User.Passwd != tmpUser.Passwd {
		ctx.Flash.Error(ctx.Tr("settings.password_incorrect"))
	} else if form.Password != form.Retype {
		ctx.Flash.Error(ctx.Tr("form.password_not_match"))
	} else {
		ctx.User.Passwd = form.Password
		ctx.User.Salt = models.GetUserSalt()
		ctx.User.EncodePasswd()
		if err := models.UpdateUser(ctx.User); err != nil {
			ctx.Handle(500, "UpdateUser", err)
			return
		}
		log.Trace("User password updated: %s", ctx.User.Name)
		ctx.Flash.Success(ctx.Tr("settings.change_password_success"))
	}

	ctx.Redirect("/user/settings/password")
}

func SettingsDelete(ctx *middleware.Context) {
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsUserSettings"] = true
	ctx.Data["PageIsSettingsDelete"] = true

	if ctx.Req.Method == "POST" {
		// tmpUser := models.User{
		// 	Passwd: ctx.Query("password"),
		// 	Salt:   ctx.User.Salt,
		// }
		// tmpUser.EncodePasswd()
		// if tmpUser.Passwd != ctx.User.Passwd {
		// 	ctx.Flash.Error("Password is not correct. Make sure you are owner of this account.")
		// } else {
		if err := models.DeleteUser(ctx.User); err != nil {
			switch err {
			case models.ErrUserOwnRepos:
				ctx.Flash.Error(ctx.Tr("form.still_own_repo"))
				ctx.Redirect("/user/settings/delete")
			default:
				ctx.Handle(500, "DeleteUser", err)
			}
		} else {
			log.Trace("Account deleted: %s", ctx.User.Name)
			ctx.Redirect("/")
		}
		return
	}

	ctx.HTML(200, SETTINGS_DELETE)
}
