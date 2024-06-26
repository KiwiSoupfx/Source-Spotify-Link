package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var currToken = ""

// var currTokenType = "Bearer"
// var currExpiresIn = "3600" //in seconds //pick back up later

var currCode = ""
var currRefreshToken = ""
var currErrors = 0
var alreadyChecking = false

/* Env vars section*/
var clientId = "" //ideally immutable
var clientSecret = ""
var maxErrors = 20
var cfgTargetPath = ""
var customMsg = ""

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Println(time.Now().Format(time.StampMilli), "root get request")
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	code := r.URL.Query().Get("code")
	if code != "" {
		currCode = code
	}
	http.Redirect(w, r, "http://localhost:8080/panel?client_id="+clientId, http.StatusTemporaryRedirect)
	initAuth("authorization_code")
	//Click the "Ok" to start the process of getting track data automatically
}

func errorLimitCheck() bool {
	if maxErrors == -1 {
		return false
	} //If we are ignoring errors and just hoping that they aren't important
	if currErrors >= maxErrors {
		return true
	}
	return false
}

func handleErrors(myError error) {
	if myError != nil {
		fmt.Println(time.Now().Format(time.StampMilli), myError)
		currErrors += 1
		if errorLimitCheck() {
			log.Fatalln("Fatal Error: Too many errors, exitting") //Prevents you from excessively spamming api. You'll still spam enough to get them to cut you off but at least you'll stop
		}
	}
}

func getCurrentTrack() (string, string, time.Duration) {
	if currCode == "" {
		return "", "", time.Duration(0).Abs()
	}
	currentTrackData := TrackData{}

	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/currently-playing", nil)
	handleErrors(err)

	client := &http.Client{}
	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + currToken},
	}
	resp, _ := client.Do(req)

	jsonData, errJson := io.ReadAll(resp.Body)
	handleErrors(errJson)

	errUnmarsh := json.Unmarshal([]byte(jsonData), &currentTrackData)
	handleErrors(errUnmarsh)

	artists := ""

	for _, artist := range currentTrackData.Item.Artists {
		if len(artists) > 0 {
			artists += ", "
		}
		artists += artist.Name
	}

	if len(currentTrackData.Item.Artists) > 0 {
		updatedCustomMsg := strings.Replace(customMsg, "{SongName}", currentTrackData.Item.Name, -1)
		updatedCustomMsg = strings.Replace(updatedCustomMsg, "{Artists}", artists, -1)
		sb := []byte(updatedCustomMsg)
		errWF := os.WriteFile(cfgTargetPath, sb, 0644)
		handleErrors(errWF)
	}

	if len(currentTrackData.Item.Artists) == 0 && currentTrackData.Item.Name == "" {
		initAuth("refresh_token") //don't spam the api if we can't read data every 0-0ms
	}
	timeLeft := time.Duration(currentTrackData.Item.DurationMs-currentTrackData.ProgressMs) * time.Millisecond

	//fmt.Println("milliseconds left: ", timeLeft) //debug to make sure we're not hitting the api too much
	fmt.Println(time.Now().Format(time.StampMilli), "Getting track data")

	return currentTrackData.Item.Name, artists, timeLeft
}

func repeatCheckTrackData(w http.ResponseWriter, _ *http.Request) {
	if currCode == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if alreadyChecking {return}
	alreadyChecking = true
	for {
		songName, _, timeLeft := getCurrentTrack()
		if timeLeft == 0 && songName == "" { //Handle no song playing
			timeLeft = 10 * time.Second //Don't hit the api too much. Maybe go higher
		}
		if timeLeft < 1*time.Second {
			timeLeft = 2 * time.Second
		} //Prevent it from getting stuck loading the next song and trying to load the song data 40 times

		time.Sleep(timeLeft)
	}
}

func initAuth(grantType string) {
	authData := AuthData{}
	form := url.Values{}

	if grantType == "authorization_code" { // do code stuff
		form.Add("code", currCode)
	} else if grantType == "refresh_token" {
		form.Add("refresh_token", currRefreshToken)
	}

	form.Add("redirect_uri", "http://localhost:8080")
	form.Add("grant_type", grantType)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(form.Encode()))
	handleErrors(err)
	encodedClientData := b64.StdEncoding.EncodeToString([]byte(clientId + ":" + clientSecret))
	client := &http.Client{}
	req.Header = http.Header{
		"Content-Type":  {"application/x-www-form-urlencoded"},
		"Authorization": {"Basic " + encodedClientData},
	}

	resp, _ := client.Do(req)

	jsonData, errJson := io.ReadAll(resp.Body)
	handleErrors(errJson)

	errUnmarsh := json.Unmarshal([]byte(jsonData), &authData)
	handleErrors(errUnmarsh)

	if authData.AccessToken != "" {
		currToken = authData.AccessToken
	}

	if authData.RefreshToken != "" {
		currRefreshToken = authData.RefreshToken
	}
}

func displayTrackData(w http.ResponseWriter, r *http.Request) {
	trackName, trackArtists, timeLeft := "", "", time.Duration(0).Abs()
	if currCode != "" {
		trackName, trackArtists, timeLeft = getCurrentTrack()
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}

	//return json so it's easier to reformat in js
	w.Header().Set("Content-Type", "application/json")
	//Also use statuses properly* so we know if there's a problem

	responseTrackData := &ResponseTrackData{
		TrackName:    trackName,
		ArtistsNames: trackArtists,
		TimeLeft:     timeLeft.String(),
	}

	json.NewEncoder(w).Encode(responseTrackData)
}

func main() {
	//Alright, time to get yucky
	handleErrors(godotenv.Load())

	clientId = os.Getenv("client_id")
	clientSecret = os.Getenv("client_secret")
	cfgTargetPath = os.Getenv("escaped_cfg_file_path")
	customMsg = os.Getenv("custom_message")

	maxErrorsStr, errError := strconv.ParseInt(os.Getenv("max_errors"), 10, 0)
	handleErrors(errError)
	maxErrors = int(maxErrorsStr) //lol???

	http.HandleFunc("/gettrackdata", displayTrackData)
	http.Handle("/panel/", http.StripPrefix("/panel/", http.FileServer(http.Dir("./public/panel"))))
	http.HandleFunc("/repeatcheck", repeatCheckTrackData)
	http.HandleFunc("/", getRoot)

	fmt.Println(time.Now().Format(time.StampMilli), "Started.")
	errSrv := http.ListenAndServe("localhost:8080", nil)
	handleErrors(errSrv)
}

type ResponseTrackData struct {
	TrackName    string `json:"track_name"`
	ArtistsNames string `json:"artists"`
	TimeLeft     string `json:"time_left"`
}

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

type AuthData struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}
