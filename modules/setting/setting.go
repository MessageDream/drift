// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the license file.

package setting

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Unknwon/com"
	"github.com/macaron-contrib/session"
	"gopkg.in/ini.v1"

	"github.com/MessageDream/drift/modules/log"
)

type Scheme string

const (
	HTTP  Scheme = "http"
	HTTPS Scheme = "https"
)

var (
	// App settings.
	AppVer  string
	AppName string
	AppLogo string
	AppUrl  string

	// Server settings.
	Protocol           Scheme
	Domain             string
	HttpAddr, HttpPort string
	SshPort            int
	OfflineMode        bool
	DisableRouterLog   bool
	CertFile, KeyFile  string
	StaticRootPath     string
	EnableGzip         bool
	AttachmentMaxSize  int64

	// Security settings.
	InstallLock          bool
	SecretKey            string
	LogInRememberDays    int
	CookieUserName       string
	CookieRememberName   string
	ReverseProxyAuthUser string

	// Picture settings.
	PictureService  string
	DisableGravatar bool

	// Log settings.
	LogRootPath string
	LogModes    []string
	LogConfigs  []string

	// Time settings.
	TimeFormat string

	// Cache settings.
	CacheAdapter  string
	CacheInternal int
	CacheConn     string

	EnableRedis    bool
	EnableMemcache bool

	// Session settings.
	SessionConfig session.Options

	// Global setting objects.
	Cfg          *ini.File
	ConfRootPath string
	CustomPath   string // Custom directory path.
	ProdMode     bool
	RunUser      string

	// I18n settings.
	Langs, Names []string
)

func init() {
	log.NewLogger(0, "console", `{"level": 0}`)
}

func ExecPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	p, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	return p, nil
}

// WorkDir returns absolute path of work directory.
func WorkDir() (string, error) {
	execPath, err := ExecPath()
	return path.Dir(strings.Replace(execPath, "\\", "/", -1)), err
}

// NewConfigContext initializes configuration context.
// NOTE: do not print any log except error.
func NewConfigContext() {
	workDir, err := WorkDir()
	if err != nil {
		log.Fatal(4, "Fail to get work directory: %v", err)
	}
	ConfRootPath = path.Join(workDir, "conf")

	Cfg, err = ini.Load(path.Join(workDir, "conf/app.ini"))
	if err != nil {
		log.Fatal(4, "Fail to parse 'conf/app.ini': %v", err)
	}

	CustomPath = os.Getenv("GOGS_CUSTOM")
	if len(CustomPath) == 0 {
		CustomPath = path.Join(workDir, "custom")
	}

	cfgPath := path.Join(CustomPath, "conf/app.ini")
	if com.IsFile(cfgPath) {
		if err = Cfg.Append(cfgPath); err != nil {
			log.Fatal(4, "Fail to load custom 'conf/app.ini': %v", err)
		}
	} else {
		log.Warn("No custom 'conf/app.ini' found, please go to '/install'")
	}

	AppName = Cfg.Section("").Key("APP_NAME").MustString("drift: blog")
	AppLogo = Cfg.Section("").Key("APP_LOGO").MustString("img/favicon.png")
	sec := Cfg.Section("server")
	AppUrl = sec.Key("ROOT_URL").MustString("http://localhost:3000/")
	if AppUrl[len(AppUrl)-1] != '/' {
		AppUrl += "/"
	}

	Protocol = HTTP
	if sec.Key("PROTOCOL").String() == "https" {
		Protocol = HTTPS
		CertFile = sec.Key("CERT_FILE").String()
		KeyFile = sec.Key("KEY_FILE").String()
	}
	Domain = sec.Key("DOMAIN").MustString("localhost")
	HttpAddr = sec.Key("HTTP_ADDR").MustString("0.0.0.0")
	HttpPort = sec.Key("HTTP_PORT").MustString("3000")
	SshPort = sec.Key("SSH_PORT").MustInt(22)
	OfflineMode = sec.Key("OFFLINE_MODE").MustBool()
	DisableRouterLog = sec.Key("DISABLE_ROUTER_LOG").MustBool()
	StaticRootPath = sec.Key("STATIC_ROOT_PATH").MustString(workDir)
	EnableGzip = sec.Key("ENABLE_GZIP").MustBool()
	LogRootPath = Cfg.Section("log").Key("ROOT_PATH").MustString(path.Join(workDir, "log"))

	secSection := Cfg.Section("security")
	InstallLock = secSection.Key("INSTALL_LOCK").MustBool()
	SecretKey = secSection.Key("SECRET_KEY").String()
	LogInRememberDays = secSection.Key("LOGIN_REMEMBER_DAYS").MustInt(7)
	CookieUserName = secSection.Key("COOKIE_USERNAME").String()
	CookieRememberName = secSection.Key("COOKIE_REMEMBER_NAME").String()
	ReverseProxyAuthUser = secSection.Key("REVERSE_PROXY_AUTHENTICATION_USER").MustString("X-WEBAUTH-USER")

	TimeFormat = map[string]string{
		"ANSIC":       time.ANSIC,
		"UnixDate":    time.UnixDate,
		"RubyDate":    time.RubyDate,
		"RFC822":      time.RFC822,
		"RFC822Z":     time.RFC822Z,
		"RFC850":      time.RFC850,
		"RFC1123":     time.RFC1123,
		"RFC1123Z":    time.RFC1123Z,
		"RFC3339":     time.RFC3339,
		"RFC3339Nano": time.RFC3339Nano,
		"Kitchen":     time.Kitchen,
		"Stamp":       time.Stamp,
		"StampMilli":  time.StampMilli,
		"StampMicro":  time.StampMicro,
		"StampNano":   time.StampNano,
	}[Cfg.Section("time").Key("FORMAT").MustString("RFC1123")]

	RunUser = Cfg.Section("").Key("RUN_USER").String()
	curUser := os.Getenv("USER")
	if len(curUser) == 0 {
		curUser = os.Getenv("USERNAME")
	}
	// Does not check run user when the install lock is off.
	if InstallLock && RunUser != curUser {
		log.Fatal(4, "Expect user(%s) but current user is: %s", RunUser, curUser)
	}

	PictureService = sec.Key("SERVICE").In("server", []string{"server"})
	DisableGravatar = Cfg.Section("picture").Key("DISABLE_GRAVATAR").MustBool()

	Langs = Cfg.Section("i18n").Key("LANGS").Strings(",")
	Names = Cfg.Section("i18n").Key("NAMES").Strings(",")
}

var Service struct {
	RegisterEmailConfirm   bool
	DisableRegistration    bool
	RequireSignInView      bool
	EnableCacheAvatar      bool
	EnableNotifyMail       bool
	EnableReverseProxyAuth bool
	LdapAuth               bool
	ActiveCodeLives        int
	ResetPwdCodeLives      int
}

func newService() {
	sec := Cfg.Section("service")
	Service.ActiveCodeLives = sec.Key("ACTIVE_CODE_LIVE_MINUTES").MustInt(180)
	Service.ResetPwdCodeLives = Cfg.MustInt("service", "RESET_PASSWD_CODE_LIVE_MINUTES", 180)
	Service.DisableRegistration = Cfg.MustBool("service", "DISABLE_REGISTRATION")
	Service.RequireSignInView = Cfg.MustBool("service", "REQUIRE_SIGNIN_VIEW")
	Service.EnableCacheAvatar = Cfg.MustBool("service", "ENABLE_CACHE_AVATAR")
	Service.EnableReverseProxyAuth = Cfg.MustBool("service", "ENABLE_REVERSE_PROXY_AUTHENTICATION")
}

var logLevels = map[string]string{
	"Trace":    "0",
	"Debug":    "1",
	"Info":     "2",
	"Warn":     "3",
	"Error":    "4",
	"Critical": "5",
}

func newLogService() {
	log.Info("%s %s", AppName, AppVer)

	// Get and check log mode.
	LogModes = strings.Split(Cfg.MustValue("log", "MODE", "console"), ",")
	LogConfigs = make([]string, len(LogModes))
	for i, mode := range LogModes {
		mode = strings.TrimSpace(mode)
		modeSec := "log." + mode
		if _, err := Cfg.GetSection(modeSec); err != nil {
			log.Fatal(4, "Unknown log mode: %s", mode)
		}

		// Log level.
		levelName := Cfg.MustValueRange("log."+mode, "LEVEL", "Trace",
			[]string{"Trace", "Debug", "Info", "Warn", "Error", "Critical"})
		level, ok := logLevels[levelName]
		if !ok {
			log.Fatal(4, "Unknown log level: %s", levelName)
		}

		// Generate log configuration.
		switch mode {
		case "console":
			LogConfigs[i] = fmt.Sprintf(`{"level":%s}`, level)
		case "file":
			logPath := Cfg.MustValue(modeSec, "FILE_NAME", path.Join(LogRootPath, "gogs.log"))
			os.MkdirAll(path.Dir(logPath), os.ModePerm)
			LogConfigs[i] = fmt.Sprintf(
				`{"level":%s,"filename":"%s","rotate":%v,"maxlines":%d,"maxsize":%d,"daily":%v,"maxdays":%d}`, level,
				logPath,
				Cfg.MustBool(modeSec, "LOG_ROTATE", true),
				Cfg.MustInt(modeSec, "MAX_LINES", 1000000),
				1<<uint(Cfg.MustInt(modeSec, "MAX_SIZE_SHIFT", 28)),
				Cfg.MustBool(modeSec, "DAILY_ROTATE", true),
				Cfg.MustInt(modeSec, "MAX_DAYS", 7))
		case "conn":
			LogConfigs[i] = fmt.Sprintf(`{"level":%s,"reconnectOnMsg":%v,"reconnect":%v,"net":"%s","addr":"%s"}`, level,
				Cfg.MustBool(modeSec, "RECONNECT_ON_MSG"),
				Cfg.MustBool(modeSec, "RECONNECT"),
				Cfg.MustValueRange(modeSec, "PROTOCOL", "tcp", []string{"tcp", "unix", "udp"}),
				Cfg.MustValue(modeSec, "ADDR", ":7020"))
		case "smtp":
			LogConfigs[i] = fmt.Sprintf(`{"level":%s,"username":"%s","password":"%s","host":"%s","sendTos":"%s","subject":"%s"}`, level,
				Cfg.MustValue(modeSec, "USER", "example@example.com"),
				Cfg.MustValue(modeSec, "PASSWD", "******"),
				Cfg.MustValue(modeSec, "HOST", "127.0.0.1:25"),
				Cfg.MustValue(modeSec, "RECEIVERS", "[]"),
				Cfg.MustValue(modeSec, "SUBJECT", "Diagnostic message from serve"))
		case "database":
			LogConfigs[i] = fmt.Sprintf(`{"level":%s,"driver":"%s","conn":"%s"}`, level,
				Cfg.MustValue(modeSec, "DRIVER"),
				Cfg.MustValue(modeSec, "CONN"))
		}

		log.NewLogger(Cfg.MustInt64("log", "BUFFER_LEN", 10000), mode, LogConfigs[i])
		log.Info("Log Mode: %s(%s)", strings.Title(mode), levelName)
	}
}

func newCacheService() {
	CacheAdapter = Cfg.MustValueRange("cache", "ADAPTER", "memory", []string{"memory", "redis", "memcache"})
	if EnableRedis {
		log.Info("Redis Enabled")
	}
	if EnableMemcache {
		log.Info("Memcache Enabled")
	}

	switch CacheAdapter {
	case "memory":
		CacheInternal = Cfg.MustInt("cache", "INTERVAL", 60)
	case "redis", "memcache":
		CacheConn = strings.Trim(Cfg.MustValue("cache", "HOST"), "\" ")
	default:
		log.Fatal(4, "Unknown cache adapter: %s", CacheAdapter)
	}

	log.Info("Cache Service Enabled")
}

func newSessionService() {
	SessionConfig.Provider = Cfg.Section("session").Key("PROVIDER").In("memory",
		[]string{"memory", "file", "redis", "mysql"})
	SessionConfig.ProviderConfig = strings.Trim(Cfg.Section("session").Key("PROVIDER_CONFIG").String(), "\" ")
	SessionConfig.CookieName = Cfg.Section("session").Key("COOKIE_NAME").MustString("i_like_gogits")
	SessionConfig.CookiePath = AppSubUrl
	SessionConfig.Secure = Cfg.Section("session").Key("COOKIE_SECURE").MustBool()
	SessionConfig.Gclifetime = Cfg.Section("session").Key("GC_INTERVAL_TIME").MustInt64(86400)
	SessionConfig.Maxlifetime = Cfg.Section("session").Key("SESSION_LIFE_TIME").MustInt64(86400)

	log.Info("Session Service Enabled")
}

// Mailer represents mail service.
type Mailer struct {
	Name         string
	Host         string
	From         string
	User, Passwd string
}

type OauthInfo struct {
	ClientId, ClientSecret string
	Scopes                 string
	AuthUrl, TokenUrl      string
}

// Oauther represents oauth service.
type Oauther struct {
	GitHub, Google, Tencent,
	Twitter, Weibo bool
	OauthInfos map[string]*OauthInfo
}

var (
	MailService  *Mailer
	OauthService *Oauther
)

func newMailService() {
	// Check mailer setting.
	if !Cfg.MustBool("mailer", "ENABLED") {
		return
	}

	MailService = &Mailer{
		Name:   Cfg.MustValue("mailer", "NAME", AppName),
		Host:   Cfg.MustValue("mailer", "HOST"),
		User:   Cfg.MustValue("mailer", "USER"),
		Passwd: Cfg.MustValue("mailer", "PASSWD"),
	}
	MailService.From = Cfg.MustValue("mailer", "FROM", MailService.User)
	log.Info("Mail Service Enabled")
}

func newRegisterMailService() {
	if !Cfg.MustBool("service", "REGISTER_EMAIL_CONFIRM") {
		return
	} else if MailService == nil {
		log.Warn("Register Mail Service: Mail Service is not enabled")
		return
	}
	Service.RegisterEmailConfirm = true
	log.Info("Register Mail Service Enabled")
}

func newNotifyMailService() {
	if !Cfg.MustBool("service", "ENABLE_NOTIFY_MAIL") {
		return
	} else if MailService == nil {
		log.Warn("Notify Mail Service: Mail Service is not enabled")
		return
	}
	Service.EnableNotifyMail = true
	log.Info("Notify Mail Service Enabled")
}

func NewServices() {
	newService()
	newLogService()
	newCacheService()
	newSessionService()
	newMailService()
	newRegisterMailService()
	newNotifyMailService()
	// ssh.Listen("2022")
}
