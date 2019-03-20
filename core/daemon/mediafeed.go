package daemon

import (
	"fmt"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-common/util"
	"net/http"
	"time"
)

func (d *Daemon) StartMediaFeed(user model.User) error {
	d.CurrentUser = user
	d.baseReq.Set("Authorization", "Bearer "+user.Token)

	if !d.running {
		go d.mediaFeedLoop()
		d.running = true
	}

	return nil
}

func (d *Daemon) mediaFeedLoop() {
	log.Info("Daemon: media feed loop run")
	for {
		d.handleSearchRequestList()
		time.Sleep(1 * time.Second)
	}
}

func (d *Daemon) handleSearchRequestList() {
	req := d.baseReq.Clone()

	var res model.SearchRequestList
	resp, _, errs := req.Get(fmt.Sprintf("%s/v1/searches/requests?offset=%d&count=%d",
		d.config.CenterBaseURL, d.searchReqOffset, model.MaxRequestItemsPerRequestCount)).
		EndStruct(&res)
	if resp != nil && resp.StatusCode == http.StatusNotFound || len(res.Items) == 0 {
		d.searchReqOffset = 0
		log.Debug("Daemon: search request list handle: empty list")
		return
	}
	if errs != nil {
		d.searchReqOffset = 0
		log.Warn("Daemon: search request list handle failed: %+v", errs)
		return
	}

	for _, r := range res.Items {
		r := *r

		go d.handleSearchRequest(r)
	}
}

func (d *Daemon) handleSearchRequest(request model.SearchRequest) {
	db := d.DB
	req := d.baseReq.Clone()
	user := model.User{
		Username: d.CurrentUser.Token,
	}

	request.RawText = util.DecodeInStr(util.ParseInStr(request.Text))

	mediaList, err := db.SearchMedia(request, 0, 1001)
	if err != nil {
		if !db.IsMediaItemNotFound(err) {
			log.Warn("Daemon: search request handle failed: %+v", err)
		}

		return
	}
	if len(mediaList.Items) == 0 {
		log.Debug("Daemon: request skipped: %s", request.RawText)
		return
	}

	var responseList model.SearchResponseList
	for _, m := range mediaList.Items {
		m := *m

		m.CoreSideID = m.ID

		resp := model.SearchResponse{
			Owner: user,
			Media: m,
		}

		responseList.Items = append(responseList.Items, &resp)
	}

	resp, _, errs := req.Post(fmt.Sprintf("%s/v1/searches/responses?mode=%s&text=%s",
		d.config.CenterBaseURL, request.Mode, request.Text)).
		Type("json").
		SendStruct(&responseList).
		End()
	if (resp != nil && resp.StatusCode != http.StatusCreated) || errs != nil {
		log.Warn("Daemon: search request handle failed: %+v", errs)
		return
	}

	log.Debug("Daemon: response sent: %s: %d", request.Text, len(responseList.Items))
}
