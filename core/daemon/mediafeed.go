package daemon

import (
	"github.com/cnaize/mz-common/model"
)

func (d *Daemon) StartMediaFeed(user model.User) error {
	d.config.CurrentUser = user

	go d.mediaFeedLoop()

	return nil
}

func (d *Daemon) mediaFeedLoop() {
	
}