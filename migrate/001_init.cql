CREATE KEYSPACE IF NOT EXISTS {{.Keyspace}} WITH replication = {'class': 'NetworkTopologyStrategy', 'replication_factor' : 1};
CREATE TABLE IF NOT EXISTS %s.%s (
    bucket TEXT, workspace TEXT, app TEXT, type TEXT, id TEXT,
    data BLOB, timestamps BIGINT,
    PRIMARY KEY ((bucket, workspace, app, type), id))
WITH CLUSTERING ORDER BY (id DESC);
