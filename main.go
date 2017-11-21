package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bcspragu/Radiotation/db"
	"github.com/bcspragu/Radiotation/music"
	"github.com/bcspragu/Radiotation/spotify"
	"github.com/bcspragu/Radiotation/srv"
	"github.com/namsral/flag"
)

var (
	_             = flag.String(flag.DefaultConfigFlagname, "config", "Path to config file")
	addr          = flag.String("addr", ":8000", "HTTP service address")
	clientID      = flag.String("client_id", "", "The Google ClientID to use")
	spotifyClient = flag.String("spotify_client_id", "", "The client ID of the Spotify application")
	spotifySecret = flag.String("spotify_secret", "", "The secret of the Spotify application")
	dev           = flag.Bool("dev", true, "If true, use development configuration")
)

func main() {
	rand.Seed(time.Now().Unix())
	flag.Parse()

	if *clientID == "" || *spotifyClient == "" || *spotifySecret == "" {
		log.Fatalf("Missing a required flag, all of  --client_id, --spotify_client_id, and --spotify_secret are required.")
	}

	idb, err := db.InitInMemDB()
	if err != nil {
		log.Fatalf("Failed to initialize datastore: %v", err)
	}

	s, err := srv.New(idb, &srv.Config{
		Dev:      *dev,
		ClientID: *clientID,
		SongServers: map[db.MusicService]music.SongServer{
			db.Spotify: spotify.NewSongServer("spotify.com", *spotifyClient, *spotifySecret),
		},
	})
	if err != nil {
		log.Fatalf("Failed to start DB: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()

	err = http.ListenAndServe(*addr, s)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

type TrackListResponse struct {
	Error   bool
	Message string
	Tracks  []music.Track
}
