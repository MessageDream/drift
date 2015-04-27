package models

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Category struct {
	Id_         bson.ObjectId `bson:"_id"`
	Name        string
	Title       string
	Content     string
	CreatedTime time.Time
	UpdatedTime time.Time
	Views       int
	NodeTime    time.Time
	Nodes       []Node
}

func (category *Category) CreatCategory(db *mgo.Database) error {
	//category.Id_ = bson.NewObjectId()
	c := db.C(ColCategory)
	err := c.Insert(category)
	SetAppCategories(db)
	return err
}

func (category *Category) UpdateCategory(db *mgo.Database) error {
	c := db.C(ColCategory)
	err := c.UpdateId(category.Id_, category)
	SetAppCategories(db)
	return err
}
