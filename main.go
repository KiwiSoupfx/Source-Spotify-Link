package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var currToken = ""
var currTokenType = "Bearer"
var currExpiresIn = "3600"

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("root get request \n")

	io.WriteString(w, "Not where you should be!\n")

	fmt.Println(r)
}

func getAuth(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Getting your info")

	//fmt.Println(r.URL.Query().Encode())

	token := r.URL.Query().Get("access_token")
	if token != "" {
		currToken = token
		//fmt.Println(currToken)
	}
	tokenType := r.URL.Query().Get("token_type")
	if tokenType != "" {
		currTokenType = tokenType
		//fmt.Println(currTokenType)
	}
	expiresIn := r.URL.Query().Get("expires_in")
	if expiresIn != "" {
		currExpiresIn = expiresIn
		//fmt.Println(currExpiresIn)
	}

	fmt.Println("auth page get request")

}

func getSuccess(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /success get request\n")
	io.WriteString(w, "authentication success!\n")
}

func handleErrors(myError error) {
	if myError != nil {
		fmt.Println(myError)
	}
}

func getCurrentTrack(w http.ResponseWriter, r *http.Request) {
	currentTrackData := TrackData{}
	//resp, err := http.Get("https://api.spotify.com/v1/me/player/currently-playing")

	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/currently-playing", nil)
	handleErrors(err)

	client := &http.Client{}
	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + currToken},
	}
	resp, err := client.Do(req)

	jsonData, errJson := io.ReadAll(resp.Body)
	handleErrors(errJson)

	errUnmarsh := json.Unmarshal([]byte(jsonData), &currentTrackData)
	handleErrors(errUnmarsh)

	//fmt.Println(currentTrackData.Item)

	if len(currentTrackData.Item.Artists) > 0 {
		io.WriteString(w, currentTrackData.Item.Artists[0].Name+" - "+currentTrackData.Item.Name)
		sb := []byte("say \"[Spotify Bot] Currently listening to " + currentTrackData.Item.Name + " - " + currentTrackData.Item.Artists[0].Name + "\"")
		errWF := os.WriteFile("D:\\Program Files (x86)\\Steam\\steamapps\\common\\GarrysMod\\garrysmod\\cfg\\listening.cfg", sb, 0644)
		handleErrors(errWF)
	} else {
		io.WriteString(w, currentTrackData.Item.Name)
		sb := []byte("say \"[Spotify Bot] Currently listening to " + currentTrackData.Item.Name + "\"")
		errWF := os.WriteFile("D:\\Program Files (x86)\\Steam\\steamapps\\common\\GarrysMod\\garrysmod\\cfg\\listening.cfg", sb, 0644)
		handleErrors(errWF)
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/getauth", getAuth)
	http.HandleFunc("/success", getSuccess)
	http.HandleFunc("/gettrackdata", getCurrentTrack)

	errSrv := http.ListenAndServe(":8080", nil)
	if errSrv != nil {
		fmt.Println(errSrv)
	}
}

//https://accounts.spotify.com/en/authorize?client_id=c273adf519ee41afa11a5c2c2e6482b3&redirect_uri=http%3A%2F%2Flocalhost%3A8080&response_type=token&scope=user-read-currently-playing
//You would not BELIEVE the hoops i jumped through to access fragments

type TrackData struct {
	Device struct {
		ID               string `json:"id"`
		IsActive         bool   `json:"is_active"`
		IsPrivateSession bool   `json:"is_private_session"`
		IsRestricted     bool   `json:"is_restricted"`
		Name             string `json:"name"`
		Type             string `json:"type"`
		VolumePercent    int    `json:"volume_percent"`
		SupportsVolume   bool   `json:"supports_volume"`
	} `json:"device"`
	RepeatState  string `json:"repeat_state"`
	ShuffleState bool   `json:"shuffle_state"`
	Context      struct {
		Type         string `json:"type"`
		Href         string `json:"href"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		URI string `json:"uri"`
	} `json:"context"`
	Timestamp  int  `json:"timestamp"`
	ProgressMs int  `json:"progress_ms"`
	IsPlaying  bool `json:"is_playing"`
	Item       struct {
		Album struct {
			AlbumType        string   `json:"album_type"`
			TotalTracks      int      `json:"total_tracks"`
			AvailableMarkets []string `json:"available_markets"`
			ExternalUrls     struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href   string `json:"href"`
			ID     string `json:"id"`
			Images []struct {
				URL    string `json:"url"`
				Height int    `json:"height"`
				Width  int    `json:"width"`
			} `json:"images"`
			Name                 string `json:"name"`
			ReleaseDate          string `json:"release_date"`
			ReleaseDatePrecision string `json:"release_date_precision"`
			Restrictions         struct {
				Reason string `json:"reason"`
			} `json:"restrictions"`
			Type    string `json:"type"`
			URI     string `json:"uri"`
			Artists []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"artists"`
		} `json:"album"`
		Artists []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Followers struct {
				Href  string `json:"href"`
				Total int    `json:"total"`
			} `json:"followers"`
			Genres []string `json:"genres"`
			Href   string   `json:"href"`
			ID     string   `json:"id"`
			Images []struct {
				URL    string `json:"url"`
				Height int    `json:"height"`
				Width  int    `json:"width"`
			} `json:"images"`
			Name       string `json:"name"`
			Popularity int    `json:"popularity"`
			Type       string `json:"type"`
			URI        string `json:"uri"`
		} `json:"artists"`
		AvailableMarkets []string `json:"available_markets"`
		DiscNumber       int      `json:"disc_number"`
		DurationMs       int      `json:"duration_ms"`
		Explicit         bool     `json:"explicit"`
		ExternalIds      struct {
			Isrc string `json:"isrc"`
			Ean  string `json:"ean"`
			Upc  string `json:"upc"`
		} `json:"external_ids"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href       string `json:"href"`
		ID         string `json:"id"`
		IsPlayable bool   `json:"is_playable"`
		LinkedFrom struct {
		} `json:"linked_from"`
		Restrictions struct {
			Reason string `json:"reason"`
		} `json:"restrictions"`
		Name        string `json:"name"`
		Popularity  int    `json:"popularity"`
		PreviewURL  string `json:"preview_url"`
		TrackNumber int    `json:"track_number"`
		Type        string `json:"type"`
		URI         string `json:"uri"`
		IsLocal     bool   `json:"is_local"`
	} `json:"item"`
	CurrentlyPlayingType string `json:"currently_playing_type"`
	Actions              struct {
		InterruptingPlayback  bool `json:"interrupting_playback"`
		Pausing               bool `json:"pausing"`
		Resuming              bool `json:"resuming"`
		Seeking               bool `json:"seeking"`
		SkippingNext          bool `json:"skipping_next"`
		SkippingPrev          bool `json:"skipping_prev"`
		TogglingRepeatContext bool `json:"toggling_repeat_context"`
		TogglingShuffle       bool `json:"toggling_shuffle"`
		TogglingRepeatTrack   bool `json:"toggling_repeat_track"`
		TransferringPlayback  bool `json:"transferring_playback"`
	} `json:"actions"`
}
