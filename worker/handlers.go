package worker

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guths/cube/task"
)

func (a *Api) StartTaskHandler(c *gin.Context) {
	te := task.TaskEvent{}

	if err := c.ShouldBindJSON(&te); err != nil {
		msg := fmt.Sprintf("Error unmarshalling body: %v\n", err)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": msg,
		})

		return
	}

	a.Worker.AddTask(te.Task)

	log.Printf("Added task %v\n", te.Task.ID)

	c.JSON(http.StatusCreated, te.Task)
}

func (a *Api) GetTasksHandler(c *gin.Context) {
	c.JSON(http.StatusAccepted, a.Worker.GetTasks())
}

func (a *Api) StopTaskHandler(c *gin.Context) {
	taskID := c.Param("taskID")

	if taskID == "" {
		log.Printf("No taskID passed in request.\n")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No taskID passed in request",
		})
	}

	tID, _ := uuid.Parse(taskID)

	_, ok := a.Worker.Db[tID]

	if !ok {
		log.Printf("No task with ID %v found", tID)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No task with provided ID",
		})
	}

	taskToStop := a.Worker.Db[tID]

	taskCopy := *taskToStop

	taskCopy.State = task.Completed
	a.Worker.AddTask(taskCopy)

	log.Printf("Added task %v to stop container %v\n", taskToStop.ID, taskToStop.ContainerID)

	c.JSON(http.StatusNoContent, gin.H{})
}
