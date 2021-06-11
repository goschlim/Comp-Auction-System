package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"goLive/frontend/common"
	"goLive/frontend/logs"

	"github.com/gorilla/mux"
)

var addr = ":5221"

// RunServer handles connection routing request.
func RunServer() {
	router := mux.NewRouter()

	router.Handle("/favicon.ico", http.NotFoundHandler())

	assets := "/assets/"
	fileServer := http.FileServer(http.Dir("." + assets))
	router.PathPrefix(assets).Handler(http.StripPrefix(assets, fileServer))

	router.HandleFunc("/", index)
	router.HandleFunc("/signup", signup)
	router.HandleFunc("/login", login)
	router.HandleFunc("/logout", logout)
	router.HandleFunc("/profile", profile)

	router.HandleFunc("/itemSell", itemSell)
	router.HandleFunc("/itemEdit", itemEdit)

	// for production, remove localhost
	if common.ConstDebug {
		addr = "localhost" + addr
	}
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	fmt.Println("Listening at port", addr)

	go func() {
		err := server.ListenAndServeTLS("ssl/cert.pem", "ssl/key.pem")
		if err != http.ErrServerClosed {
			logs.Fatalf("<ListenAndServeTLS>: %v\n", err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// save data
		if common.ConstDebug {
			logs.SaveChecksum(logs.HashPath, logs.LogPath)
		}

		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		logs.Infof("<server shutdown failure>: %+v\n", err)
	}
}
