package daemon

import (
	"fmt"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-common/util"
	"net/http"
	"time"
)

func (d *Daemon) searchFeedLoop() {
	log.Info("Daemon: search feed loop run")
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
	if (resp != nil && resp.StatusCode == http.StatusNotFound) || len(res.Items) == 0 {
		d.searchReqOffset = 0
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

	d.searchReqOffset += uint(len(res.Items))
}

func (d *Daemon) handleSearchRequest(request model.SearchRequest) {
	db := d.DB
	req := d.baseReq.Clone()

	self := model.User{
		Username: d.CurrentUser.Username,
	}
	request.RawText = util.DecodeInStr(util.ParseInStr(request.Text))

	mediaList, err := db.SearchMedia(self, request, 0, model.MaxResponseItemsCount)
	if err != nil {
		if !db.IsMediaItemNotFound(err) {
			log.Warn("Daemon: search request handle failed: %+v", err)
		}

		return
	}
	if len(mediaList.Items) == 0 {
		log.Debug("Daemon: search request skipped: %s", request.RawText)
		return
	}

	responseList := model.SearchResponseList{
		Items: []*model.SearchResponse{},
	}
	for _, m := range mediaList.Items {
		m := *m

		m.CoreSideID = m.ID

		resp := model.SearchResponse{
			Owner: self,
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

	log.Debug("Daemon: search response sent: %s: %d", request.Text, len(responseList.Items))
}
