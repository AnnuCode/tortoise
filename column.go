package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/redis/go-redis/v9"
)

const APPEND = -1

type column struct {
	focus  bool
	status status
	list   list.Model
	height int
	width  int
}

func (c *column) Focus() {
	c.focus = true
}

func (c *column) Blur() {
	c.focus = false
}

func (c *column) Focused() bool {
	return c.focus
}

func newColumn(status status) column {
	var focus bool
	if status == todo {
		focus = true
	}
	defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	defaultList.SetShowHelp(false)
	return column{focus: focus, status: status, list: defaultList}
}

// Init does initial setup for the column.
func (c column) Init() tea.Cmd {
	return nil
}

// Update handles all the I/O for columns.
func (c column) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.setSize(msg.Width, msg.Height)
		c.list.SetSize(msg.Width/margin, msg.Height/2)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Edit):
			if len(c.list.VisibleItems()) != 0 {
				task := c.list.SelectedItem().(Task)
				f := NewForm(task.Tasktitle, task.Taskdescription, task.Taskdeadline)
				f.index = c.list.Index()
				f.col = c
				return f.Update(nil)
			}
		case key.Matches(msg, keys.New):
			f := newDefaultForm()
			f.index = APPEND
			f.col = c
			return f.Update(nil)
		case key.Matches(msg, keys.Delete):
			return c, c.DeleteCurrent()
		case key.Matches(msg, keys.Enter):
			return c, c.MoveToNext()
		}
	}
	c.list, cmd = c.list.Update(msg)
	return c, cmd
}

func (c column) View() string {
	return c.getStyle().Render(c.list.View())
}

func (c *column) DeleteCurrent() tea.Cmd {
	if len(c.list.VisibleItems()) > 0 {
		c.list.RemoveItem(c.list.Index())
	}

	var cmd tea.Cmd
	c.list, cmd = c.list.Update(nil)
	return cmd
}

func (c *column) Set(i int, t Task) tea.Cmd {
	if i != APPEND {
		return c.list.SetItem(i, t)
	}
	return c.list.InsertItem(APPEND, t)
}

func (c *column) setSize(width, height int) {
	c.width = width / margin
}

func (c *column) getStyle() lipgloss.Style {
	if c.Focused() {
		return lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Height(c.height).
			Width(c.width)
	}
	return lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.HiddenBorder()).
		Height(c.height).
		Width(c.width)
}

type moveMsg struct {
	Task
}

func (c *column) MoveToNext() tea.Cmd {
	var task Task
	var ok bool
	// If nothing is selected, the SelectedItem will return Nil.
	if task, ok = c.list.SelectedItem().(Task); !ok {
		return nil
	}
	// move item
	c.list.RemoveItem(c.list.Index())

	// remove task from db before updating its status
	RemoveTask(task)

	// status update
	task.Taskstatus = c.status.getNext()

	// add the updated task to db
	AddTask(task)

	// refresh list
	var cmd tea.Cmd
	c.list, cmd = c.list.Update(nil)

	return tea.Sequence(cmd, func() tea.Msg { return moveMsg{task} })
}

func RemoveTask(task Task) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(Task{Tasktitle: task.Tasktitle, Taskdescription: task.Taskdescription, Taskstatus: task.Taskstatus, Taskdeadline: task.Taskdeadline}); err != nil {
		fmt.Printf("Error encoding custom type: %v", err)
	}

	err := client.ZRem(ctx, "tasks", buffer.Bytes()).Err()
	if err != nil {
		fmt.Printf("encountered an error in removing: %v", task)
	}
}

func AddTask(task Task) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(Task{Tasktitle: task.Tasktitle, Taskdescription: task.Taskdescription, Taskstatus: task.Taskstatus, Taskdeadline: task.Taskdeadline}); err != nil {
		fmt.Printf("Error encoding custom type: %v", err)
	}

	dl, err := time.Parse("2006-01-02", task.Taskdeadline)
	if err != nil {
		fmt.Println("Error parsing deadline:", err)
	}

	err = client.ZAdd(ctx, "tasks", redis.Z{Score: float64(dl.Unix()), Member: buffer.Bytes()}).Err()
	if err != nil {
		fmt.Println("Error adding task:", err)
	}
}
