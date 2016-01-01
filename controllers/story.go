package controllers

import (
	"fmt"
	"github.com/aacanakin/hn"
	"github.com/aacanakin/hnr/resources"
	"github.com/aacanakin/hnr/util"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"strconv"
)

var storyTypes = [...]string{"top", "new", "ask", "show", "jobs"}

type StoryController struct {
	Controller
	Story *resources.StoryResource
	HN    *hn.Client
}

func (c *StoryController) isValidType(storyType string) bool {

	for _, st := range storyTypes {
		if st == storyType {
			return true
		}
	}

	return false
}

func (c *StoryController) GetComments(ctx *gin.Context) {

	storyId, _ := strconv.Atoi(ctx.Param("storyId"))
	story := c.Story.FindById(storyId)
	comments, err := c.Story.FindByIds(story.Kids)
	if err != nil {
		ctx.JSON(404, gin.H{
			"error": "Not Found",
		})
		ctx.Abort()
		return
	}

	ctx.JSON(200, comments)
}

func (c *StoryController) Get(ctx *gin.Context) {

	storyType := ctx.Param("type")
	offset, _ := strconv.Atoi(ctx.Query("offset"))
	count, _ := strconv.Atoi(ctx.Query("count"))

	if !c.isValidType(storyType) {

		ctx.JSON(400, gin.H{
			"error": "Invalid story type",
		})
		ctx.Abort()
		return
	}

	ids, err := c.Story.FindCacheStoryIds(storyType)

	if err != nil {

		ctx.JSON(404, gin.H{
			"error": "Not Found",
		})

		ctx.Abort()
		return
	}

	if len(ids) < offset || len(ids) < offset+count {
		ctx.JSON(404, gin.H{
			"error": "Not Found",
		})
		ctx.Abort()
		return
	}

	stories, err := c.Story.FindByIds(ids[offset : offset+count])

	if err != nil {
		ctx.JSON(404, gin.H{
			"error": "Not Found",
		})
		ctx.Abort()
		return
	}

	ctx.JSON(200, stories)
}

func (c *StoryController) Cache(ctx *gin.Context) {

	var (
		ids []int
		err error
	)

	storyType := ctx.Param("type")

	if !c.isValidType(storyType) {

		ctx.JSON(400, gin.H{
			"error": "Invalid story type",
		})
		ctx.Abort()
		return
	}

	// check if lock exists
	filename := fmt.Sprintf("./storage/locks/%s.lock", storyType)
	if util.Exists(filename) {
		ctx.JSON(400, gin.H{
			"error": "Already caching",
		})
		ctx.Abort()
		return
	}

	// put lock
	_, err = os.Create(filename)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(500, nil)
		return
	}

	syncLogFile, err := os.OpenFile("./storage/logs/sync.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	logger := log.New(io.Writer(syncLogFile), "", log.Ldate|log.Ltime|log.Lshortfile)

	fmt.Printf("Starting saving %s stories..\n", storyType)

	if storyType == "top" {
		ids, err = c.HN.TopStories()
	} else if storyType == "new" {
		ids, err = c.HN.NewStories()
	} else if storyType == "ask" {
		ids, err = c.HN.AskStories()
	} else if storyType == "show" {
		ids, err = c.HN.ShowStories()
	} else if storyType == "jobs" {
		ids, err = c.HN.JobStories()
	}

	if err != nil {
		logger.Println(err)
		util.CleanLock(storyType)
		ctx.JSON(500, gin.H{
			"error": "Could not retrieve stories",
		})
		ctx.Abort()
		return
	}

	// slice & limit ids
	limit := viper.GetInt("story.limit")
	if limit > len(ids) {
		limit = len(ids)
	}
	ids = ids[0:limit]

	ctx.JSON(202, gin.H{
		"status": "started",
	})

	go func() {

		for k, id := range ids {

			item, err := c.HN.Item(id)

			if err != nil {
				logger.Println(err)
				util.CleanLock(storyType)
				break
			}

			logger.Printf("Saving story type %s with index: %d id: %d title: %s\n", storyType, k, id, item.Title)

			err = c.Story.SaveStory(id, item)
			if err != nil {
				logger.Println(err)
			}

			// cache comments
			/*
				for j, kidId := range item.Kids {
					comment, err := c.HN.Item(kidId)
					if err != nil {
						logger.Println(err)
					}

					logger.Printf("Saving comment with index: %d id: %d item title: %s\n", j, kidId, item.Title)

					err = c.Story.SaveStory(kidId, comment)
					if err != nil {
						logger.Println(err)
					}
				}
			*/

		}

		// write ids to memcache & clean locks
		c.Story.CacheStoryIds(storyType, ids)
		util.CleanLock(storyType)

	}()
}
