package main

import (
	"context"
	"errors"
	"html/template"
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
	oidc "github.com/coreos/go-oidc"
	"github.com/namsral/flag"
)

type (
	tmplData struct {
		ClientID string
		Host     string
		Room     *db.Room
		User     *db.User
		Rooms    struct {
			ID          string
			DisplayName string
		}
	}
)

var (
	_             = flag.String(flag.DefaultConfigFlagname, "config", "path to config file")
	addr          = flag.String("addr", ":8000", "http service address")
	clientID      = flag.String("client_id", "", "The Google ClientID to use")
	spotifyClient = flag.String("spotify_client_id", "", "The client ID of the Spotify application")
	spotifySecret = flag.String("spotify_secret", "", "The secret of the Spotify application")

	spotifyServer  music.SongServer
	googleVerifier *oidc.IDTokenVerifier

	errNoTracks = errors.New("radiotation: no tracks in room")
)

func main() {
	rand.Seed(time.Now().Unix())
	flag.Parse()

	if *clientID == "" || *spotifyClient == "" || *spotifySecret == "" {
		log.Fatalf("Missing a required flag, all of  --client_id, --spotify_client_id, and --spotify_secret are required.")
	}

	spotifyServer = spotify.NewSongServer("spotify.com", *spotifyClient, *spotifySecret)

	tmpls, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Can't load templates, dying: %v", err)
	}

	if err != nil {
		log.Fatalf("Can't load or generate keys, dying: %v", err)
	}

	googleProvider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		log.Fatalf("Failed to get provider for Google: %v", err)
	}
	googleVerifier = googleProvider.Verifier(&oidc.Config{
		ClientID: *clientID,
	})

	//db, err := initBoltDB()
	db, err := initInMemDB()
	if err != nil {
		log.Fatalf("Failed to initialize datastore: %v", err)
	}
	f, err := os.Open("inmemdb")
	if err != nil && !os.IsNotExist(err) {
		// A legitimate error, not just 'the file wasn't found'
		log.Fatalf("Failed to open datastore file for reading: %v", err)
	}

	if err == nil {
		if err := db.Load(f); err != nil {
			f.Close()
			log.Fatalf("Failed to load datastore: %v", err)
		}
	}
	f.Close()

	s := srv.New(db, db)
	s := &srv{
		Template: tmpls,
		db:       db,
		h: hub{
			broadcast:   make(chan []byte),
			register:    make(chan *connection),
			unregister:  make(chan *connection),
			connections: make(map[*connection]bool),
			userconns:   make(map[*db.User]*connection),
		},
	}
	go s.h.run()

	if err := servePaths(); err != nil {
		log.Fatalf("Can't serve static assets: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		f, err := os.Create("inmemdb")
		if err != nil && !os.IsExist(err) {
			// A legitimate error, not just 'the file was found'
			log.Fatalf("Failed to open datastore file for writing: %v", err)
		}
		if err := s.db.Save(f); err != nil {
			log.Fatalf("Failed to save datastore: %v", err)
		}
		if err := f.Close(); err != nil {
			log.Fatalf("Failed to close datastore file: %v", err)
		}
		os.Exit(1)
	}()

	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

type QueueResponse struct {
	Error   bool
	Message string
}

type TrackListResponse struct {
	Error   bool
	Message string
	Tracks  []music.Track
}

type TrackResponse struct {
	Error   bool
	Message string
	Track   music.Track
}

func songServer(rm *db.Room) music.SongServer {
	switch rm.MusicService {
	case db.Spotify:
		return spotifyServer
	default:
		return nil
	}
}
