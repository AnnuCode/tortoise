package main

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
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
	tasks, err := client.ZRange(ctx, "tasks", 0, -1).Result()
	if err != nil {
		fmt.Println("Error listing tasks:", err)
		return nil, err
	}
	var filteredTasks []Task
	for _, task := range tasks {
		var item Task
		buffer := bytes.NewBuffer([]byte(task))
		decoder := gob.NewDecoder(buffer)
		if err := decoder.Decode(&item); err != nil {
			fmt.Printf("Error decoding member: %v", err)
			fmt.Printf("culprit is: %v", item)
		}
		if item.Taskstatus == status {
			filteredTasks = append(filteredTasks, item)
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
