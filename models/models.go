package models

import (
	"fmt"
	"log"
	"os"
	"path"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/MessageDream/drift/modules/setting"
)

const (
	TypeTopic     = 'T'
	TypeArticle   = 'A'
	DefaultAvatar = "gopher_teal.jpg"

	ADS                = "ads"
	ARTICLE_CATEGORIES = "articlecategories"
	COMMENTS           = "comments"
	CONTENTS           = "contents"
	NODES              = "nodes"
	LINK_EXCHANGES     = "link_exchanges"
	STATUS             = "status"
	USERS              = "users"
	LOGIN_SOURCE       = "LoginSource"
	OAUTH_2            = "Oauth2"

	DB_LOG_MODE_DEBUG = "log_debug"
)

var (
	DbCfg struct {
		Type, Host, Port, Name, User, Pwd, Path, SslMode, LogMode string
	}
)

func LoadModelsConfig() {
	DbCfg.Type = setting.Cfg.MustValue("database", "DB_TYPE")
	DbCfg.Host = setting.Cfg.MustValue("database", "HOST")
	DbCfg.Name = setting.Cfg.MustValue("database", "NAME")
	DbCfg.User = setting.Cfg.MustValue("database", "USER")
	if len(DbCfg.Pwd) == 0 {
		DbCfg.Pwd = setting.Cfg.MustValue("database", "PASSWD")
	}
	DbCfg.SslMode = setting.Cfg.MustValue("database", "SSL_MODE")
	DbCfg.LogMode = setting.Cfg.MustValue("database", "LOG_MODE")
	DbCfg.Path = setting.Cfg.MustValue("database", "PATH", "data/gogs.db")
}

func GetSession() (*mgo.Session, error) {
	cnnstr := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
		DbCfg.User, DbCfg.Pwd, DbCfg.Host, DbCfg.Port, DbCfg.Name)
	return mgo.Dial(cnnstr)
}

func GetDb(se *mgo.Session) *mgo.Database {
	return se.DB(DbCfg.Name)
}

// 状态,MongoDB中只存储一个状态
type Status struct {
	Id_        bson.ObjectId `bson:"_id"`
	UserCount  int
	TopicCount int
	ReplyCount int
	UserIndex  int
}

func SetDB() (err error) {

	if DbCfg.LogMode == DB_LOG_MODE_DEBUG {
		logPath := path.Join(setting.LogRootPath, "mongo.log")
		os.MkdirAll(path.Dir(logPath), os.ModePerm)

		f, err := os.Create(logPath)
		if err != nil {
			return fmt.Errorf("models.init(fail to create xorm.log): %v", err)
		}
		mgo.SetLogger(log.New(f, "mongodb", log.Ltime))
		mgo.SetDebug(true)
	}

	return nil
}

func Ping() error {
	se, err := GetSession()
	if err != nil {
		return err
	}
	defer se.Close()
	return se.Ping()
}



var (
DB         *mgo.Database
Categories []Category   //常驻内存
Tags       []TagWrapper //常驻内存
)

func InitDb() {

	conn := common.Webconfig.Dbconn
	if conn == "" {
		beego.Error("数据库地址还没有配置,请到config内配置db字段.")
		os.Exit(1)
	}

	session, err := mgo.Dial(conn)
	if err != nil {
		beego.Error("MongoDB连接失败:", err.Error())
		os.Exit(1)
	}

	session.SetMode(mgo.Monotonic, true)

	DB = session.DB("messageblog")
	SetAppCategories()
	SetAppTags()
}

func SetAppCategories() {
	Categories, _ = GetAllCategory()
}

func SetAppTags() {
	tags, _ := GetAllTags()
	Tags = *tags
}

func GetArticles(condition *bson.M, offset int, limit int, sort string) (*[]Article, int, error) {
	c := DB.C("article")
	var article []Article
	query := c.Find(condition).Skip(offset).Limit(limit)
	if sort != "" {
		query = query.Sort(sort)
	}
	err := query.All(&article)
	total, _ := c.Find(condition).Count()
	return &article, total, err
}

func GetArticlesByTag(tagname string, offset int, limit int, sort string) (*[]Article, int, error) {
	var tag TagWrapper
	for _, v := range Tags {
		if tagname == v.Name {
			tag = v
			break
		}
	}
	return GetArticles(&bson.M{"_id": bson.M{"$in": tag.ArticleIds}}, offset, limit, sort)
}

func GetArticlesByNode(condition *bson.M, offset int, limit int, sort string) (*[]Article, int, error) {
	return GetArticles(condition, offset, limit, sort)
}

func GetArticleCount() int {
	c := DB.C("article")
	total, _ := c.Count()
	return total
}

func GetArticle(condition *bson.M) (*Article, error) {
	c := DB.C("article")
	var article Article
	err := c.Find(condition).One(&article)
	return &article, err
}

func DeleteArticles(condition *bson.M) (*mgo.ChangeInfo, error) {
	c := DB.C("article")
	return c.RemoveAll(condition)
}

func DeleteArticle(condition *bson.M) error {
	c := DB.C("article")
	return c.Remove(condition)
}

func GetTags(condition *bson.M, offset int, limit int, sort string) (*[]TagWrapper, int, error) {
	c := DB.C("tags")
	var tags []TagWrapper
	query := c.Find(condition).Skip(offset).Limit(limit)
	if sort != "" {
		query = query.Sort(sort)
	}
	err := query.All(&tags)
	total, _ := c.Find(condition).Count()

	return &tags, total, err
}

func GetAllTags() (*[]TagWrapper, error) {
	c := DB.C("tags")
	var tags []TagWrapper
	err := c.Find(&bson.M{}).All(&tags)
	return &tags, err
}

func GetTagCount() int {
	return len(Tags)
}

func GetAllCategory() ([]Category, error) {
	c := DB.C("category")
	var categories []Category
	err := c.Find(bson.M{}).Sort("createdtime").All(&categories)
	return categories, err
}

func GetCategoryById(id bson.ObjectId) Category {
	var category Category
	for _, v := range Categories {
		if v.Id_ == id {
			category = v
			break
		}
	}
	return category
}
func GetCategoryNodeName(nname string) Category {
	var category Category
	for _, v := range Categories {
		flag := false
		for _, va := range v.Nodes {
			if va.Name == nname {
				flag = true
				break
			}
		}
		if flag {
			category = v
		}
	}
	return category
}

func DeleteCategory(condition *bson.M) error {
	c := DB.C("category")
	err := c.Remove(condition)
	SetAppCategories()
	return err
}

func GetSubscribes(condition *bson.M, offset int, limit int, sort string) (*[]Subscription, int, error) {
	c := DB.C("subscription")
	var subs []Subscription
	query := c.Find(condition).Skip(offset).Limit(limit)
	if sort != "" {
		query = query.Sort(sort)
	}
	err := query.All(&subs)
	total, _ := c.Find(condition).Count()
	return &subs, total, err
}

func removeDuplicate(slis *[]bson.ObjectId) {
	found := make(map[bson.ObjectId]bool)
	j := 0
	for i, val := range *slis {
		if _, ok := found[val]; !ok {
			found[val] = true
			(*slis)[j] = (*slis)[i]
			j++
		}
	}
	*slis = (*slis)[:j]
}

func setTags(tags *[]string, aid bson.ObjectId) {
	for _, v := range *tags {
		tag := &TagWrapper{
			Name:       v,
			Title:      v,
			Count:      1,
			ArticleIds: []bson.ObjectId{aid},
		}
		tag.SetTag()
	}
}
