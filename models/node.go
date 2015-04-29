package models

import (
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

)

type Node struct {
	Name        string
	Title       string
	Content     string
	CreatedTime time.Time
	UpdatedTime time.Time
	Views       int64
	ArticleTime time.Time
}

func (node *Node) GetAllArticles(db *mgo.Database, offset int, limit int, sort string) (*[]Article, int, error) {
	c := db.C(ColArticle)

	var article []Article
	q := bson.M{"nname": node.Name}
	query := c.Find(q).Skip(offset).Limit(limit)
	if sort != "" {
		query = query.Sort(sort)
	}
	err := query.All(&article)
	total, _ := c.Find(q).Count()
	return &article, total, err
}

func (node *Node) GetArticleCount(db *mgo.Database) (int, error) {
	c := db.C(ColArticle)
	return c.Find(bson.M{"nname": node.Name}).Count()
}
