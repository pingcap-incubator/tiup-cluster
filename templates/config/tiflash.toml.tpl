default_profile = "default"
display_name = "TiFlash"
listen_host = "0.0.0.0"
mark_cache_size = 5368709120
tmp_path = "{{.DeployDir}}/tiflash/data/tmp"
path = "{{.DeployDir}}/tiflash/data/db"
tcp_port = {{.TCPPort}}
http_port = {{.HTTPPort}}

[flash]
tidb_status_addr = "{{.TiDBStatusAddrs}}"
service_addr = "{{.IP}}:{.FlashServicePort}"

[flash.flash_cluster]
cluster_manager_path = "{{.DeployDir}}/bin/tiflash/flash_cluster_manager"
log = "{{.DeployDir}}/log/tiflash_cluster_manager.log"
master_ttl = 60
refresh_interval = 20
update_rule_interval = 5

[flash.proxy]
config = "{{.DeployDir}}/conf/tiflash-learner.toml"

[status]
metrics_port = {{.MetricsPort}}

[logger]
errorlog = "/{{.DeployDir}}/log/tiflash_error.log"
log = "/{{.DeployDir}}/log/tiflash.log"
count = 20
level = "debug"
size = "1000M"

[application]
runAsDaemon = true

[raft]
pd_addr = "{{.PDAddrs}}"

[quotas]

[quotas.default]

[quotas.default.interval]
duration = 3600
errors = 0
execution_time = 0
queries = 0
read_rows = 0
result_rows = 0

[users]

[users.default]
password = ""
profile = "default"
quota = "default"

[users.default.networks]
ip = "::/0"

[users.readonly]
password = ""
profile = "readonly"
quota = "default"

[users.readonly.networks]
ip = "::/0"

[profiles]

[profiles.default]
load_balancing = "random"
max_memory_usage = 10000000000
use_uncompressed_cache = 0

[profiles.readonly]
readonly = 1

