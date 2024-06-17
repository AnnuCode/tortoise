package main

type Task struct {
	Taskstatus      status
	Tasktitle       string
	Taskdescription string
	Taskdeadline    string
	Taskid          int
}

func NewTask(status status, title, description, deadline string) Task {
	return Task{Taskstatus: status, Tasktitle: title, Taskdescription: description, Taskdeadline: deadline}
}

func (t *Task) Next() {
	if t.Taskstatus == done {
		t.Taskstatus = todo
	} else {
		t.Taskstatus++
	}
}

// implement the list.Item interface
func (t Task) FilterValue() string {
	return t.Tasktitle
}

func (t Task) Title() string {
	return t.Tasktitle
}

func (t Task) Description() string {
	return t.Taskdescription
}
func (t Task) Deadline() string {
	return t.Taskdeadline
}
