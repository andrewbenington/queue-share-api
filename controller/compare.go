package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/history"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/exp/maps"
)

type CompareTracksResp struct {
	StreamsByURI map[string]map[uuid.UUID]int64 `json:"streams_by_uri"`
	RanksByURI   map[string]map[uuid.UUID]int64 `json:"ranks_by_uri"`
	TrackData    map[string]db.TrackData        `json:"track_data"`
	FriendData   map[uuid.UUID]*db.User         `json:"friend_data"`
}

func (c *StatsController) UserCompareFriendTopTracks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, ok := ctx.Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	userUUID, err := userUUIDFromRequest(r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	start, end := getStartAndEndTimes(r)
	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	friends, err := db.New(db.Service().DB).UserGetFriends(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filter := getFilterParams(r)
	filter.Max = 50
	userStreamsByURI, userRanksByURI, _, err := history.CalcTrackStreamsAndRanks(ctx, userUUID, filter, start, end, nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	streamsByURI := map[string]map[uuid.UUID]int64{}
	ranksByURI := map[string]map[uuid.UUID]int64{}

	for uri, userStreams := range userStreamsByURI {
		uriStreams := map[uuid.UUID]int64{}
		uriStreams[userUUID] = userStreams

		streamsByURI[uri] = uriStreams
	}

	for uri, userRanks := range userRanksByURI {
		uriRanks := map[uuid.UUID]int64{}
		uriRanks[userUUID] = userRanks

		ranksByURI[uri] = uriRanks
	}

	resp := CompareTracksResp{
		StreamsByURI: map[string]map[uuid.UUID]int64{},
		RanksByURI:   map[string]map[uuid.UUID]int64{},
		FriendData:   map[uuid.UUID]*db.User{},
	}

	for _, friend := range friends {
		friendStreamsByURI, friendRanksByURI, _, err := history.CalcTrackStreamsAndRanks(ctx, friend.ID, filter, start, end, nil, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for uri, friendStreams := range friendStreamsByURI {
			if uriStreams, ok := streamsByURI[uri]; ok {
				uriStreams[friend.ID] = friendStreams
			}
		}

		for uri, friendRanks := range friendRanksByURI {
			if uriRanks, ok := ranksByURI[uri]; ok {
				uriRanks[friend.ID] = friendRanks
			}
		}

		resp.FriendData[friend.ID] = friend
	}

	trackIDs := map[string]bool{}

	// Only include tracks with stats from at least 1 friend if true
	sharedOnly := strings.EqualFold(r.URL.Query().Get("shared_only"), "true")

	for uri, streamsByUser := range streamsByURI {
		if !sharedOnly || len(streamsByUser) > 1 {
			resp.StreamsByURI[uri] = streamsByUser
			trackIDs[service.IDFromURIMust(uri)] = true
		}
	}
	for uri, ranksByUser := range ranksByURI {
		if !sharedOnly || len(ranksByUser) > 1 {
			resp.RanksByURI[uri] = ranksByUser
		}
	}

	trackResults, err := service.GetTracks(ctx, spClient, maps.Keys(trackIDs))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.TrackData = trackResults

	json.NewEncoder(w).Encode(resp)
}

type CompareArtistsResp struct {
	StreamsByURI  map[string]map[uuid.UUID]int64         `json:"streams_by_uri"`
	RanksByURI    map[string]map[uuid.UUID]int64         `json:"ranks_by_uri"`
	ArtistData    map[string]spotify.FullArtist          `json:"artist_data"`
	FriendData    map[uuid.UUID]*db.User                 `json:"friend_data"`
	FriendStreams map[uuid.UUID][]*history.ArtistStreams `json:"friend_streams"`
}

func (c *StatsController) UserCompareFriendTopArtists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, ok := ctx.Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	userUUID, err := userUUIDFromRequest(r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	friends, err := db.New(db.Service().DB).UserGetFriends(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	start, end := getStartAndEndTimes(r)
	filter := getFilterParams(r)
	filter.Max = 50

	userStreamsByURI, userRanksByURI, _, err := history.CalcArtistStreamsAndRanks(ctx, userUUID, filter, start, end, nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	streamsByURI := map[string]map[uuid.UUID]int64{}
	ranksByURI := map[string]map[uuid.UUID]int64{}

	for uri, userStreams := range userStreamsByURI {
		uriStreams := map[uuid.UUID]int64{}
		uriStreams[userUUID] = userStreams

		streamsByURI[uri] = uriStreams
	}

	for uri, userRanks := range userRanksByURI {
		uriRanks := map[uuid.UUID]int64{}
		uriRanks[userUUID] = userRanks

		ranksByURI[uri] = uriRanks
	}

	resp := CompareArtistsResp{
		StreamsByURI:  map[string]map[uuid.UUID]int64{},
		RanksByURI:    map[string]map[uuid.UUID]int64{},
		FriendData:    map[uuid.UUID]*db.User{},
		FriendStreams: map[uuid.UUID][]*history.ArtistStreams{},
	}

	for _, friend := range friends {
		friendStreamsByURI, friendRanksByURI, streamList, err := history.CalcArtistStreamsAndRanks(ctx, friend.ID, filter, start, end, nil, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for uri, friendStreams := range friendStreamsByURI {
			if uriStreams, ok := streamsByURI[uri]; ok {
				uriStreams[friend.ID] = friendStreams
			}
		}

		for uri, friendRanks := range friendRanksByURI {
			if uriRanks, ok := ranksByURI[uri]; ok {
				uriRanks[friend.ID] = friendRanks
			}
		}

		resp.FriendData[friend.ID] = friend
		resp.FriendStreams[friend.ID] = streamList
	}

	artistIDs := map[string]bool{}

	// Only include tracks with stats from at least 1 friend
	for uri, streams := range streamsByURI {
		if len(streams) > 1 {
			resp.StreamsByURI[uri] = streams
			artistIDs[service.IDFromURIMust(uri)] = true
		}
	}

	for uri, streams := range ranksByURI {
		if len(streams) > 1 {
			resp.RanksByURI[uri] = streams
		}
	}

	artistResults, err := service.GetArtists(ctx, spClient, maps.Keys(artistIDs))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.ArtistData = artistResults

	json.NewEncoder(w).Encode(resp)
}

func getStartAndEndTimes(r *http.Request) (time.Time, time.Time) {

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	end := now

	startParam := r.URL.Query().Get("start")
	startUnix, err := strconv.ParseInt(startParam, 10, 64)
	if err == nil {
		start = time.Unix(startUnix, 0)
	}

	endParam := r.URL.Query().Get("end")
	endUnix, err := strconv.ParseInt(endParam, 10, 64)
	if err == nil {
		end = time.Unix(endUnix, 0)
	}

	return start, end
}
