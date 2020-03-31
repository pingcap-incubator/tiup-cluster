log-file = "{{.DeployDir}}/log/tiflash_tikv.log"

[readpool]

[readpool.coprocessor]

[readpool.storage]

[server]
engine-addr = "{{.IP}}:{{.FlashServicePort}}"
addr = "0.0.0.0:{{.FlashProxyPort}}"
advertise-addr = "{{.IP}}:{{.FlashProxyPort}}"
status-addr = "{{.IP}}:{{.FlashProxyStatusPort}}"

[storage]
data-dir = "{{.DeployDir}}/tiflash/data/flash"

[pd]

[metric]

[raftstore]

[coprocessor]

[rocksdb]
wal-dir = ""

[rocksdb.defaultcf]

[rocksdb.lockcf]

[rocksdb.writecf]

[raftdb]

[raftdb.defaultcf]

[security]
ca-path = ""
cert-path = ""
key-path = ""

[import]