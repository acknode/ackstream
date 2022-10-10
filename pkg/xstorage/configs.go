package xstorage

type Configs struct {
	Hosts          []string `json:"hosts" mapstructure:"ACKSTREAM_XSTORAGE_HOSTS"`
	Keyspace       string   `json:"keyspace" mapstructure:"ACKSTREAM_XSTORAGE_KEYSPACE"`
	Table          string   `json:"table" mapstructure:"ACKSTREAM_XSTORAGE_TABLE"`
	BucketTemplate string   `json:"bucket_template" mapstructure:"ACKSTREAM_XSTORAGE_BUCKET_TEMPLATE"`
}
