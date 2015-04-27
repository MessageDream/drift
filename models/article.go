package models

import (
"gopkg.in/mgo.v2/bson"
	"time"
	"strings"
	"html/template"
	"regexp"
)

type Article struct {
	Id_            bson.ObjectId `bson:"_id"`
	CName          string
	NName          string
	Name           string
	Author         string
	Title          string
	Text           template.HTML
	Tags           []string
	FeaturedPicURL string
	Summary        template.HTML
	Views          int
	Comments       []*Comment
	IsThumbnail    bool
	CreatedTime    time.Time
	ModifiedTime   time.Time
}

func (article *Article) SetSummary() {
	if article.IsThumbnail {
		article.Summary = article.Text
	} else {
		strs := strings.Split(string(article.Text), "<!--more-->")
		//beego.Error(strs[0])
		n := len(strs)
		if n > 0 {
			article.Summary = template.HTML(strs[0])
		}
	}
}

func (article *Article) GetFirstParagraph() *template.HTML {
	rx := regexp.MustCompile(`<p>(.*)</p>`)
	p := rx.FindStringSubmatch(string(article.Text))
	//beego.Error(p)
	n := len(p)
	if n > 1 {
		rep := template.HTML(p[1] + "...")
		return &rep
	}
	return nil
}

func (article *Article) GetCategory() *Category {
	// c := DB.C("category")
	// var category Category
	// c.Find(bson.M{"name": article.CName}).One(&category)
	// return &category
	var category Category
	for _, v := range Categories {
		if v.Name == article.CName {
			category = v
			break
		}
	}
	return &category
}

func (article *Article) GetNode() *Node {
	var node Node
	for _, v := range Categories {
		if v.Name == article.CName {
			for _, va := range v.Nodes {
				if va.Name == article.NName {
					node = va
					break
				}
			}
			break
		}
	}

	return &node
}

func (article *Article) GetTags() *[]TagWrapper {
	return article.GetSelfTags()
}

func (article *Article) CreatArticle() error {
	article.Id_ = bson.NewObjectId()
	c := DB.C("article")
	err := c.Insert(article)
	go setTags(&article.Tags, article.Id_)
	return err
}

func (article *Article) UpdateArticle() error {
	c := DB.C("article")
	err := c.UpdateId(article.Id_, article)
	go setTags(&article.Tags, article.Id_)

	return err
}

func (article *Article) GetCommentCount() int {
	return 1
}

func (article *Article) GetAroundArticle() (*Article, *Article, error) {
	c := DB.C("article")
	var preresult, nextresult Article
	err := c.Find(&bson.M{"createdtime": &bson.M{"$lt": article.CreatedTime}}).Sort("-createdtime").Limit(1).One(&preresult)

err = c.Find(&bson.M{"createdtime": &bson.M{"$gt": article.CreatedTime}}).Sort("createdtime").Limit(1).One(&nextresult)
return &preresult, &nextresult, err
}

func (article *Article) GetSameTagArticles(limit int) (articles []Article) {
	ids := make([]bson.ObjectId, 0)
	for _, v := range Tags {
		for _, tag := range article.Tags {
			if tag == v.Title || tag == v.Name {
				for _, va := range v.ArticleIds {
					if va != article.Id_ {
						ids = append(ids, va)
					}
				}
			}
		}
	}
	d := DB.C("article")
	d.Find(&bson.M{"_id": &bson.M{"$in": ids}}).Limit(limit).All(&articles)
return
}

func (article *Article) GetSelfTags() *[]TagWrapper {
	var tags []TagWrapper
	for _, v := range Tags {
		for _, va := range v.ArticleIds {
			if va != article.Id_ {
				tags = append(tags, v)
			}
		}
	}
	return &tags
}

func (article *Article) HasFeaturedPic() bool {
	if len(article.FeaturedPicURL) == 0 {
		return false
	}
	return true
}

func (article *Article) HasSummary() bool {
	if len(article.Summary) == 0 {
		return false
	}
	return true
}

func (article *Article) UpdateViews() {
	article.Views++
	article.UpdateArticle()
}
