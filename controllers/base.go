package controllers

// EventType event  type
type EventType int

const (
	// EVENTADD type add
	EVENTADD EventType = iota
	// EVENTUPDATE type update
	EVENTUPDATE
	// EVENTDELETE type delete
	EVENTDELETE
)
