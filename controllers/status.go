package controllers

import (
	"github.com/gin-gonic/gin"
)

type StatusController struct{}

func (c *StatusController) Ping(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"message": "pong"})
}
