package models
import (
	"time"
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

func (node *Node) GetAllArticles(offset int, limit int, sort string) (*[]Article, int, error) {
	c := DB.C("article")

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

func (node *Node) GetArticleCount() (int, error) {
	c := DB.C("article")
	return c.Find(bson.M{"nname": node.Name}).Count()
}

// func (node *Node) GetCategory() (*Category, error) {
// 	c := DB.C("category")
// 	var category Category
// 	err := c.Find(bson.M{"_id": node.CId_}).One(&category)
// 	return &category, err
// }

// func (node *Node) CreatNode() error {
// 	//node.Id_ = bson.NewObjectId()
// 	c := DB.C("node")
// 	err := c.Insert(node)
// 	return err
// }

// func (node *Node) UpdateNode() error {
// 	c := DB.C("node")
// 	err := c.UpdateId(node.Id_, node)
// 	return err
// }
