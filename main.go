package main

import (
	"flag"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-core/core"
	"github.com/cnaize/mz-core/db/sqlite"
	"os"
	"path/filepath"
)

var (
	MZCoreVersion = ""

	loggerConfig log.Config
	coreConfig   core.Config
)

func init() {
	coreConfig.Version = MZCoreVersion

	flag.UintVar(&loggerConfig.Lvl, "log-lvl", 5, "log level")

	flag.StringVar(&coreConfig.Daemon.WorkingDir, "working-dir", ".", "working directory")
	flag.StringVar(&coreConfig.Daemon.SettingsFile, "settings-file", "mz-core-settings.json", "settings file name")
	flag.StringVar(&coreConfig.Daemon.DatabaseFile, "db-file", "mz-core-db.sql", "database file name")
	flag.StringVar(&coreConfig.Daemon.CenterBaseURL, "center-base-url", "http://localhost:11310", "central server base url")

	flag.UintVar(&coreConfig.Port, "port", 11311, "server port")
}

func main() {
	flag.Parse()
	log.Init(loggerConfig)

	if coreConfig.Version == "" {
		coreConfig.Version = os.Getenv("MZ_CORE_VERSION")
	}

	db, err := sqlite.New(filepath.Join(coreConfig.Daemon.WorkingDir, coreConfig.Daemon.DatabaseFile))
	if err != nil {
		log.Fatal("MuzeZone Core: db open failed: %+v", err)
	}

	if err := core.New(coreConfig, db).Run(); err != nil {
		log.Fatal("MuzeZone Core: core run failed %+v", err)
	}
}
