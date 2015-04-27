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
	ColArticle      = "article"
	ColCategory     = "category"
	ColNode         = "node"
	ColTags         = "tags"
	ColSubscription = "subscription"
	ColUser         = "user"

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
		Type, Host, Port, Name, User, Pwd, Path, LogMode string
	}

	Categories []Category   //常驻内存
	Tags       []TagWrapper //常驻内存
)

func LoadModelsConfig() {
	DbCfg.Type = setting.Cfg.MustValue("database", "DB_TYPE")
	DbCfg.Host = setting.Cfg.MustValue("database", "HOST")
	DbCfg.Name = setting.Cfg.MustValue("database", "NAME")
	DbCfg.User = setting.Cfg.MustValue("database", "USER")
	if len(DbCfg.Pwd) == 0 {
		DbCfg.Pwd = setting.Cfg.MustValue("database", "PASSWD")
	}
	DbCfg.LogMode = setting.Cfg.MustValue("database", "LOG_MODE")
	DbCfg.Path = setting.Cfg.MustValue("database", "PATH", "data/gogs.db")
}

func GetSession() (*mgo.Session, error) {
	cnnstr := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
		DbCfg.User, DbCfg.Pwd, DbCfg.Host, DbCfg.Port, DbCfg.Name)
	return mgo.Dial(cnnstr)
}

func getDb(se *mgo.Session) *mgo.Database {
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

func SetAppCategories(db *mgo.Database) {
	Categories, _ = GetAllCategory(db)
}

func SetAppTags(db *mgo.Database) {
	tags, _ := GetAllTags(db)
	Tags = *tags
}

func GetArticles(db *mgo.Database, condition *bson.M, offset int, limit int, sort string) (*[]Article, int, error) {
	c := db.C(ColArticle)
	var article []Article
	query := c.Find(condition).Skip(offset).Limit(limit)
	if sort != "" {
		query = query.Sort(sort)
	}
	err := query.All(&article)
	total, _ := c.Find(condition).Count()
	return &article, total, err
}

func GetArticlesByTag(db *mgo.Database, tagname string, offset int, limit int, sort string) (*[]Article, int, error) {
	var tag TagWrapper
	for _, v := range Tags {
		if tagname == v.Name {
			tag = v
			break
		}
	}
	return GetArticles(db, &bson.M{"_id": bson.M{"$in": tag.ArticleIds}}, offset, limit, sort)
}

func GetArticlesByNode(db *mgo.Database, condition *bson.M, offset int, limit int, sort string) (*[]Article, int, error) {
	return GetArticles(db, condition, offset, limit, sort)
}

func GetArticleCount(db *mgo.Database) int {
	c := db.C(ColArticle)
	total, _ := c.Count()
	return total
}

func GetArticle(db *mgo.Database, condition *bson.M) (*Article, error) {
	c := db.C(ColArticle)
	var article Article
	err := c.Find(condition).One(&article)
	return &article, err
}

func DeleteArticles(db *mgo.Database, condition *bson.M) (*mgo.ChangeInfo, error) {
	c := db.C(ColArticle)
	return c.RemoveAll(condition)
}

func DeleteArticle(db *mgo.Database, condition *bson.M) error {
	c := db.C(ColArticle)
	return c.Remove(condition)
}

func GetTags(db *mgo.Database, condition *bson.M, offset int, limit int, sort string) (*[]TagWrapper, int, error) {
	c := db.C(ColTags)
	var tags []TagWrapper
	query := c.Find(condition).Skip(offset).Limit(limit)
	if sort != "" {
		query = query.Sort(sort)
	}
	err := query.All(&tags)
	total, _ := c.Find(condition).Count()

	return &tags, total, err
}

func GetAllTags(db *mgo.Database) (*[]TagWrapper, error) {
	c := db.C(ColTags)
	var tags []TagWrapper
	err := c.Find(&bson.M{}).All(&tags)
	return &tags, err
}

func GetTagCount() int {
	return len(Tags)
}

func GetAllCategory(db *mgo.Database) ([]Category, error) {
	c := db.C(ColCategory)
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

func DeleteCategory(db *mgo.Database, condition *bson.M) error {
	c := db.C(ColCategory)
	err := c.Remove(condition)
	SetAppCategories(db)
	return err
}

func GetSubscribes(db *mgo.Database, condition *bson.M, offset int, limit int, sort string) (*[]Subscription, int, error) {
	c := db.C(ColSubscription)
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
