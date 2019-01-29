package main // import github.com/bcspragu/Radiotation/cmd/server

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	firebase "firebase.google.com/go"
	"github.com/bcspragu/Radiotation/spotify"
	"github.com/bcspragu/Radiotation/sqldb"
	"github.com/bcspragu/Radiotation/srv"
	"github.com/namsral/flag"
	"google.golang.org/api/option"
)

var (
	addr          = flag.String("addr", ":8000", "HTTP service address")
	clientID      = flag.String("client_id", "", "The Google ClientID to use")
	fcmKey        = flag.String("fcm_key", "", "The Firebase Cloud Messaging Key to use")
	spotifyClient = flag.String("spotify_client_id", "", "The client ID of the Spotify application")
	spotifySecret = flag.String("spotify_secret", "", "The secret of the Spotify application")
	dev           = flag.Bool("dev", true, "If true, use development configuration")
	frontendGlob  = flag.String("frontend_glob", "", "The location to find the frontend HTML files.")
	staticDir     = flag.String("static_dir", "", "The location to find the static frontend files.")
	projectID     = flag.String("project_id", "", "The Firebase/GCP project ID to authenticate with.")
	creds         = flag.String("service_accoutn_creds", "", "The location of the JSON-formatted service account credentials.")
	dbPath        = flag.String("db_path", "", "The location to store/load the SQLite database.")
)

func main() {
	rand.Seed(time.Now().Unix())
	flag.Parse()

	if *clientID == "" || *spotifyClient == "" || *spotifySecret == "" {
		log.Fatalf("Missing a required flag, all of  --client_id, --spotify_client_id, and --spotify_secret are required.")
	}

	db, err := sqldb.New(*dbPath, sqldb.CryptoRandSource{})
	if err != nil {
		log.Fatalf("Failed to initialize datastore: %v", err)
	}

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: *projectID,
	}, option.WithCredentialsFile(*creds))
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	auth, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth: %v", err)
	}

	s, err := srv.New(db, &srv.Config{
		Dev:        *dev,
		ClientID:   *clientID,
		FCMKey:     *fcmKey,
		AuthClient: auth,

		SongServer:   spotify.NewSongServer("spotify.com", *spotifyClient, *spotifySecret),
		FrontendGlob: *frontendGlob,
		StaticDir:    *staticDir,
	})
	if err != nil {
		log.Fatalf("Failed to start DB: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		db.Close()
		os.Exit(1)
	}()

	err = http.ListenAndServe(*addr, s)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
