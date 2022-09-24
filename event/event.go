package event

type Event struct {
	// partition keys
	// often be the timestamp with format YYMMDD
	Bucket    string `json:"bucket"`
	Workspace string `json:"workspace"`
	App       string `json:"app"`
	Type      string `json:"type"`

	// clustering keys
	// chronologically sortable id - ksuid - 1-second resolution
	Id string `json:"id"`

	// properties
	Payload      string `json:"payload"`
	CreationTime int64  `json:"creation_time"`
}

func (msg *Event) SetPartitionKeys(event *Event) {
	msg.Bucket = event.Bucket
	msg.Workspace = event.Workspace
	msg.App = event.App
	msg.Type = event.Type
}
