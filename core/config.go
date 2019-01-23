package core

import (
	"github.com/cnaize/mz-core/core/daemon"
)

type Config struct {
	Port   uint
	Daemon daemon.Config
}
