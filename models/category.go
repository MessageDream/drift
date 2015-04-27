package models

import (
	"gopkg.in/mgo.v2/bson"
	mgo "gopkg.in/mgo.v2"
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
	c := DB.C("category")
	err := c.Insert(category)
	SetAppCategories()
	return err
}

func (category *Category) UpdateCategory() error {
	c := DB.C("category")
	err := c.UpdateId(category.Id_, category)
	SetAppCategories()
	return err
}

// func (category *Category) GetAllNodes() *[]Node {
// 	c := DB.C("node")
// 	var nodes []Node
// 	c.Find(&bson.M{"_cid": category.Id_}).All(&nodes)
// 	return &nodes
// }

// func (category *Category) SetNodeId(nid bson.ObjectId) {
// 	if category.NodeIds != nil {
// 		category.NodeIds = append(category.NodeIds, nid)
// 		removeDuplicate(&category.NodeIds)
// 	} else {
// 		category.NodeIds = []bson.ObjectId{nid}
// 	}
// 	category.NodeCount = len(category.NodeIds)
// }
