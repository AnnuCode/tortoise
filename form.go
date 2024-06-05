package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
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
	help        help.Model
	title       textinput.Model
	description textarea.Model
	deadline    textarea.Model
	col         column
	index       int
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

func (f Form) CreateTask() Task {

	return Task{f.col.status, f.title.Value(), f.description.Value(), f.deadline.Value()}
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
			var buffer bytes.Buffer
			encoder := gob.NewEncoder(&buffer)
			if err := encoder.Encode(Task{Tasktitle: f.title.Value(), Taskdescription: f.description.Value(), Taskstatus: f.col.status, Taskdeadline: f.deadline.Value()}); err != nil {
				fmt.Printf("Error encoding custom type: %v", err)
			}

			err = client.ZAdd(ctx, "tasks", redis.Z{Score: float64(dl.Unix()), Member: buffer.Bytes()}).Err()
			if err != nil {
				fmt.Println("Error adding task:", err)
			} else {
				fmt.Println("Task added successfully!")
			}
			// Return the completed form as a message.
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
