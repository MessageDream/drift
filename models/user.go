package models

import (
//	"crypto/sha256"
//	"encoding/hex"
//	"fmt"
//	"strings"
	"sync"
	"time"
	"errors"

//	"github.com/Unknwon/com"
//	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

//	"github.com/MessageDream/drift/modules/base"
// 	"github.com/MessageDream/drift/modules/log"
// 	"github.com/MessageDream/drift/modules/setting"
)

type UserType int

const (
	INDIVIDUAL UserType = iota // Historic reason to make it starts at 0.
	ORGANIZATION
)

var (
	ErrUserOwnRepos          = errors.New("User still have ownership of repositories")
	ErrUserHasOrgs           = errors.New("User still have membership of organization")
	ErrUserAlreadyExist      = errors.New("User already exist")
	ErrUserNotExist          = errors.New("User does not exist")
	ErrUserNotKeyOwner       = errors.New("User does not the owner of public key")
	ErrEmailAlreadyUsed      = errors.New("E-mail already used")
	ErrUserNameIllegal       = errors.New("User name contains illegal characters")
	ErrLoginSourceNotExist   = errors.New("Login source does not exist")
	ErrLoginSourceNotActived = errors.New("Login source is not actived")
	ErrUnsupportedLoginType  = errors.New("Login source is unknown")
)

var (
	lock          sync.Locker = &sync.Mutex{}
	userIndex                 = 0
	illegalEquals             = []string{"debug", "raw", "install", "api", "avatar", "user", "org", "help", "stars", "issues", "pulls", "commits", "repo", "template", "admin", "new"}
)

// 用户
type User struct {
	Id_          bson.ObjectId `bson:"_id"`
	UserName     string
	LowerName    string
	FullName     string
	Password     string
	Salt         string `bson:"salt"`
	Email        string
	Avatar       string
	LoginSource  int64
	Website      string
	Location     string
	Tagline      string
	Bio          string
	Twitter      string
	Weibo        string
	JoinedAt     time.Time
	Follow       []string
	Fans         []string
	IsAdmin      bool
	IsActive     bool
	ValidateCode string
	ResetCode    string
	Rands        string
	Index        int
}
