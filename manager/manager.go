package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/guths/cube/task"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[uuid.UUID]*task.Task
	EventDB       map[uuid.UUID]*task.TaskEvent
	Workers       []string
	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
	LastWorker    int
}

func (m *Manager) SelectWorker() string {
	var newWorker int
	if m.LastWorker+1 < len(m.Workers) {
		newWorker = m.LastWorker + 1
		m.LastWorker++
	} else {
		newWorker = 0
		m.LastWorker = 0
	}

	return m.Workers[newWorker]
}

func (m *Manager) UpdateTasks() {
	for _, worker := range m.Workers {
		log.Printf("checking worker %v for task updates", worker)

		url := fmt.Sprintf("http://%s/tasks", worker)

		resp, err := http.Get(url)

		if err != nil {
			log.Printf("Error connecting to %v: %v\n", worker, err)
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error sending request: %v\n", err)
		}

		d := json.NewDecoder(resp.Body)

		var tasks []*task.Task

		err = d.Decode(&tasks)

		if err != nil {
			log.Printf("Error unmarshalling tasks: %s\n", err.Error())
		}

		for _, t := range tasks {
			log.Printf("Attempting to update task %v\n", t.ID)

			_, ok := m.TaskDb[t.ID]

			if !ok {
				log.Printf("Task with ID %s not found\n", t.ID)
				return
			}

			if m.TaskDb[t.ID].State != t.State {
				m.TaskDb[t.ID].State = t.State
			}

			m.TaskDb[t.ID].StartTime = t.StartTime
			m.TaskDb[t.ID].FinishTime = t.FinishTime
			m.TaskDb[t.ID].ContainerID = t.ContainerID
		}

	}
}

func (m *Manager) SendWork() {
	if m.Pending.Len() > 0 {
		w := m.SelectWorker()

		e := m.Pending.Dequeue()

		te := e.(task.TaskEvent)

		t := te.Task

		log.Printf("Pulled %v off pending queue\n", t)

		m.EventDB[te.ID] = &te
		m.WorkerTaskMap[w] = append(m.WorkerTaskMap[w], te.Task.ID)
		m.TaskWorkerMap[t.ID] = w

		t.State = task.Scheduled
		m.TaskDb[t.ID] = &t

		data, err := json.Marshal(te)

		if err != nil {
			log.Printf("unable to marshal task object %v \n", t)
		}

		url := fmt.Sprintf("http://%s/tasks", w)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))

		if err != nil {
			log.Printf("Error connection to %v: %v\n", w, err)
			m.Pending.Enqueue(te)
			return
		}

		d := json.NewDecoder(resp.Body)

		if resp.StatusCode != http.StatusCreated {
			log.Printf("Response error (%d)", resp.StatusCode)
			return
		}

		t = task.Task{}

		err = d.Decode(&t)

		if err != nil {
			fmt.Printf("Error decoding response: %s\n", err.Error())
			return
		}

		log.Printf("%#v\n", t)

	} else {
		log.Println("No work in the queue")
	}
}
