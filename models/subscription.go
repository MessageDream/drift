package models
import (
	"gopkg.in/mgo.v2/bson"
)

type Subscription struct {
	Id_    bson.ObjectId `bson:"_id"`
	Email  string
	Uid    string
	Status bool
}

func (subscription *Subscription) Set() error {
	c := DB.C("subscription")
	var sub Subscription
	err := c.Find(bson.M{"email": subscription.Email}).One(&sub)
if err != nil {
subscription.Id_ = bson.NewObjectId()
err = c.Insert(subscription)
} else {
sub.Email = subscription.Email
sub.Status = subscription.Status
sub.Uid = subscription.Uid
err = c.UpdateId(sub.Id_, sub)
}
return err
}

func (subscription *Subscription) UpdateState() error {
	c := DB.C("subscription")
	var sub Subscription
	err := c.Find(bson.M{"uid": subscription.Uid}).One(&sub)
if err != nil {
return err
}
sub.Status = subscription.Status
err = c.UpdateId(sub.Id_, sub)
return err
}
