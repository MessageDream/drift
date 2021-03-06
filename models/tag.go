package models

import (
	"time"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type TagWrapper struct {
	Id_          bson.ObjectId `bson:"_id"`
	Name         string
	Title        string
	Count        int
	CreatedTime  time.Time
	ModifiedTime time.Time
	ArticleIds   []bson.ObjectId
}

func (tag *TagWrapper) SetTag(db *mgo.Database) error {
	c := db.C(ColTags)
	var err error
	flag := false
	for _, v := range Tags {
		if tag.Name == v.Name {
			v.ArticleIds = append(v.ArticleIds, tag.ArticleIds...)
			removeDuplicate(&v.ArticleIds)
			v.Count = len(v.ArticleIds)
			v.ModifiedTime = bson.Now()
			err = c.UpdateId(v.Id_, v)
			flag = true
			break
		}
	}

	if !flag {
		tag.Id_ = bson.NewObjectId()
		tag.CreatedTime = bson.Now()
		tag.ModifiedTime = bson.Now()
		Tags = append(Tags, *tag)
		err = c.Insert(tag)
	}

	SetAppTags(db)
	return err
}
