package storage

type Configs struct {
	Hosts    []string `json:"hosts" mapstructure:"ACKSTREAM_STORAGE_HOSTS"`
	Keyspace string   `json:"keyspace" mapstructure:"ACKSTREAM_STORAGE_KEYSPACE"`
	Table    string   `json:"table" mapstructure:"ACKSTREAM_STORAGE_TABLE"`
}
