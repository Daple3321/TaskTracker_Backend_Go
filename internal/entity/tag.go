package entity

// Represents a tag.
// Related to tags table in database
type Tag struct {
	Id     int    `json:"id"`
	UserId int    `json:"userId"`
	Name   string `json:"name"`
}

// Connects tag with task.
// Related to task_tag table in database
type TaskTag struct {
	TaskId int
	TagId  int
}
