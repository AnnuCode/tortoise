package main

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/redis/go-redis/v9"
)

// Provides the mock data to fill the kanban board
func (b *Board) initLists() {
	b.cols = []column{
		newColumn(todo),
		newColumn(inProgress),
		newColumn(done),
	}
	// Init todos
	todos, err := getTasksByStatus(todo)
	if err != nil {
		fmt.Printf("Error getting todo tasks: %v", err)
	}
	todoTaskItems := tasksToItems(todos)

	board.cols[todo].list.Title = "To Do"
	board.cols[todo].list.SetItems(todoTaskItems)

	// Init in progress
	inProgressItems, err := getTasksByStatus(inProgress)
	if err != nil {
		fmt.Printf("Error getting todo tasks: %v", err)
	}
	inProgressTaskItems := tasksToItems(inProgressItems)

	board.cols[inProgress].list.Title = "In Progress"
	board.cols[inProgress].list.SetItems(inProgressTaskItems)

	// Init done
	doneItems, err := getTasksByStatus(done)
	if err != nil {
		fmt.Printf("Error getting todo tasks: %v", err)
	}
	doneTaskItems := tasksToItems(doneItems)

	board.cols[done].list.Title = "Done"
	board.cols[done].list.SetItems(doneTaskItems)
}

func getTasksByStatus(status status) ([]Task, error) {
	var filteredTasks []Task

	// Fetch task IDs from the sorted set
	taskIDs, err := client.ZRange(ctx, "tasks:sorted", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("error fetching task IDs from sorted set: %v", err)
	}

	// Fetch task details from the hash
	commands, err := client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, taskID := range taskIDs {
			// hashKey := fmt.Sprintf("task:%s", taskID) this step isn't required
			pipe.HGet(ctx, "tasks", taskID)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error executing Redis pipeline: %v", err)
	}

	// Deserialize task data
	for _, cmd := range commands {
		data, err := cmd.(*redis.StringCmd).Bytes()
		if err != nil {

			if err == redis.Nil {
				// Handle the case where the task does not exist in the hash
				fmt.Println("Task data not found in Redis")
				continue
			}

			return nil, fmt.Errorf("error getting task data from Redis: %v", err)
		}

		var task Task
		buffer := bytes.NewBuffer(data)
		decoder := gob.NewDecoder(buffer)
		if err := decoder.Decode(&task); err != nil {
			return nil, fmt.Errorf("error decoding task data: %v", err)
		}

		// Filter tasks by the specified status
		if task.Taskstatus == status {
			filteredTasks = append(filteredTasks, task)
		}
	}
	return filteredTasks, nil
}

func tasksToItems(tasks []Task) []list.Item {
	var items []list.Item
	for _, t := range tasks {
		items = append(items, t)
	}
	return items
}
