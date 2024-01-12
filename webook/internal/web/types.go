package web

import "github.com/gin-gonic/gin"

type handler interface {
	RegisterRouter(server *gin.Engine)
}
