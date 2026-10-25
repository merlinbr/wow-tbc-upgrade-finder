// Command forever-armory serves a local, single-user Forever armory.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/browser"
	"github.com/wowsims/tbc/internal/forever/catalog"
	"github.com/wowsims/tbc/internal/forever/server"
	"github.com/wowsims/tbc/internal/forever/service"
	"github.com/wowsims/tbc/internal/forever/store"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:0", "loopback listen address")
	noBrowser := flag.Bool("no-browser", false, "do not open a browser on startup")
	dbPath := flag.String("db", "./forever-armory.db", "SQLite database path")
	flag.Parse()

	listenAddr, err := loopbackAddr(*addr)
	if err != nil {
		log.Fatal(err)
	}

	cat, err := catalog.Load()
	if err != nil {
		log.Fatalf("forever-armory: %v", err)
	}
	database, err := store.NewSQLite(*dbPath)
	if err != nil {
		log.Fatalf("forever-armory: open database %s: %v", *dbPath, err)
	}
	defer database.Close()
	svc := service.New(database, cat)

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("forever-armory: listen %s: %v", listenAddr, err)
	}
	url := "http://" + listener.Addr().String()
	fmt.Println(url)

	if !*noBrowser {
		go func() {
			if err := browser.OpenURL(url); err != nil {
				log.Printf("forever-armory: open browser: %v", err)
			}
		}()
	}

	httpServer := &http.Server{
		Handler:           server.New(cat, svc),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("forever-armory: %v", err)
	}
}

// loopbackAddr refuses any address that is not loopback.
func loopbackAddr(raw string) (string, error) {
	host, port, err := net.SplitHostPort(raw)
	if err != nil {
		return "", fmt.Errorf("invalid --addr %q: %w", raw, err)
	}
	if host == "" || host == "localhost" {
		host = "127.0.0.1"
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return "", fmt.Errorf("--addr host %q is not an IP address", host)
	}
	if !ip.IsLoopback() {
		return "", fmt.Errorf("refusing non-loopback address %q", raw)
	}
	return net.JoinHostPort(host, port), nil
}
