package worker

import "github.com/gin-gonic/gin"

type Api struct {
	Address string
	Port    int
	Worker  *Worker
	Router  *gin.Engine
}

func (a *Api) initRouter() {
	a.Router = gin.New()

	a.Router.POST("/tasks", a.StartTaskHandler)
	a.Router.GET("/tasks", a.GetTasksHandler)
	a.Router.DELETE("/tasks/:taskID", a.StopTaskHandler)
}

func (a *Api) Start() {
	a.initRouter()
	a.Router.Run()
}
