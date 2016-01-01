package resources

import (
	"encoding/json"
	"github.com/aacanakin/hn"
	"github.com/bradfitz/gomemcache/memcache"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type StoryResource struct {
	MC *memcache.Client
	MG *mgo.Session
}

func (r *StoryResource) FindById(id int) hn.Item {

	var story hn.Item
	r.MG.DB("hnr").C("stories").FindId(id).One(&story)
	return story
}

func (r *StoryResource) FindByIds(ids []int) ([]hn.Item, error) {

	var stories []hn.Item
	c := r.MG.DB("hnr").C("stories")
	query := bson.M{"_id": bson.M{"$in": ids}}
	c.Find(query).All(&stories)

	orderedStories := make([]hn.Item, len(stories))
	for k, id := range ids {
		for _, story := range stories {
			if story.ID == id {
				orderedStories[k] = story
				break
			}
		}
	}
	return orderedStories, nil
}

func (r *StoryResource) CacheStoryIds(storyType string, ids []int) error {

	jids, err := json.Marshal(ids)
	if err != nil {
		return err
	}

	r.MC.Set(&memcache.Item{
		Key:   storyType,
		Value: jids,
	})

	return nil
}

func (r *StoryResource) FindCacheStoryIds(storyType string) ([]int, error) {

	item, err := r.MC.Get(storyType)
	if err != nil {
		return []int{}, err
	}

	var ids []int
	err = json.Unmarshal(item.Value, &ids)
	return ids, nil
}

func (r *StoryResource) SaveStory(id int, item *hn.Item) error {

	c := r.MG.DB("hnr").C("stories")
	_, err := c.UpsertId(id, item)
	return err
}
