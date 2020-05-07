package meta

import (
	tiupmeta "github.com/pingcap-incubator/tiup/pkg/meta"
)

var _env *tiupmeta.Environment

// SetTiupEnv the gloable env used.
func SetTiupEnv(env *tiupmeta.Environment) {
	_env = env
}

// TiupEnv Get the gloable env used.
func TiupEnv() *tiupmeta.Environment {
	return _env
}
