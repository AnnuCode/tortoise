package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/redis/go-redis/v9"
)

type Form struct {
	help         help.Model
	title        textinput.Model
	description  textarea.Model
	deadline     textarea.Model
	col          column
	index        int
	originalTask Task // Store the original task for comparison
}

func newDefaultForm() *Form {
	return NewForm("task name", "task description", "task deadline")
}

func NewForm(title, description, deadline string) *Form {
	form := Form{
		help:        help.New(),
		title:       textinput.New(),
		description: textarea.New(),
		deadline:    textarea.New(),
	}
	form.title.Placeholder = title
	form.description.Placeholder = description
	form.deadline.Placeholder = deadline
	form.title.Focus()
	return &form
}

var nextTaskID int

func latestTaskID() (int, error) {
	// Get the current value of the task ID counter from Redis
	taskIDStr, err := client.Get(ctx, "taskIDCounter").Result()
	if err != nil {
		return 0, err
	}

	// Convert the task ID string to an integer
	lastTaskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		return 0, err
	}
	return lastTaskID, nil
}

func (f Form) CreateTask() Task {

	return Task{f.col.status, f.title.Value(), f.description.Value(), f.deadline.Value(), nextTaskID}
}

func (f Form) Init() tea.Cmd {
	return nil
}

func (f Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case column:
		f.col = msg
		f.col.list.Index()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return f, tea.Quit

		case key.Matches(msg, keys.Back):
			return board.Update(nil)
		case key.Matches(msg, keys.Enter):
			if f.title.Focused() {
				f.title.Blur()
				f.description.Focus()
				return f, textarea.Blink
			}
			if f.description.Focused() {
				f.description.Blur()
				f.deadline.Focus()
				return f, textarea.Blink
			}
			dl, err := time.Parse("2006-01-02", f.deadline.Value())
			if err != nil {
				fmt.Println("Error parsing deadline:", err)
			}

			if f.index != APPEND {
				editedTask := Task{
					Taskid:          f.originalTask.Taskid,
					Tasktitle:       f.title.Value(),
					Taskdescription: f.description.Value(),
					Taskdeadline:    f.deadline.Value(),
					Taskstatus:      f.originalTask.Taskstatus,
				}

				if editedTask != f.originalTask {
					RemoveTask(f.originalTask) // Remove the old task

					AddTask(editedTask) // Add the new task
					// f.col.list.SetItem(f.index, editedTask)
				}
			} else {
				id, err := latestTaskID()
				if err == nil {
					nextTaskID = id + 1
				} else {
					nextTaskID++
				}

				// Create the new task
				newTask := Task{
					Taskid:          nextTaskID,
					Tasktitle:       f.title.Value(),
					Taskdescription: f.description.Value(),
					Taskstatus:      f.col.status,
					Taskdeadline:    f.deadline.Value(),
				}

				var buffer bytes.Buffer
				encoder := gob.NewEncoder(&buffer)
				if err := encoder.Encode(newTask); err != nil {
					fmt.Printf("Error encoding custom type: %v", err)
				}

				// Using a pipeline to perform multiple operations
				_, err = client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
					hashKey := fmt.Sprintf("task:%d", newTask.Taskid)
					pipe.HSet(ctx, "tasks", hashKey, buffer.Bytes())
					pipe.ZAdd(ctx, "tasks:sorted", redis.Z{Score: float64(dl.Unix()), Member: hashKey})
					return nil
				})
				if err != nil {
					fmt.Println("Error executing Redis pipeline:", err)
					return f, nil
				}
				_, err = client.Incr(ctx, "taskIDCounter").Result()
				if err != nil {
					fmt.Println("error incrementing the counter")
				}
				fmt.Println("Task added successfully!")
			}
			return board.Update(f)
		}
	}
	if f.title.Focused() {
		f.title, cmd = f.title.Update(msg)
		return f, cmd
	} else if f.description.Focused() {
		f.description, cmd = f.description.Update(msg)
		return f, cmd
	} else if f.deadline.Focused() {
		f.deadline, cmd = f.deadline.Update(msg)
		return f, cmd
	}
	return f, cmd
}

func (f Form) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		f.title.View(),
		f.description.View(),
		f.deadline.View(),
		f.help.View(keys))
}
