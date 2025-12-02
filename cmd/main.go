package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	core "AvitoTest/internal/core/team"
	rep "AvitoTest/internal/repository/team"
	handlTeam "AvitoTest/internal/server/team"

	corePR "AvitoTest/internal/core/pullrequest"
	repPR "AvitoTest/internal/repository/pullrequest"
	handlPR "AvitoTest/internal/server/pullrequest"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	ctx := context.Background()
	conn, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5432/avito_db?sslmode=disable")
	if err != nil {
		log.Fatalf("failed connecto to db %v", err)
	}
	defer conn.Close()

	err = conn.Ping(ctx)
	if err != nil {
		log.Fatalf("failed ping to db %v", err)
	}

	repo := rep.NewTeam(conn)
	core := core.NewTeam(repo)
	handlerTeam := handlTeam.NewTeamsHandler(core)

	repoPR := repPR.NewPullReq(conn)
	corePR := corePR.NewPullReq(repoPR)
	handlerPR := handlPR.NewTeamsHandler(corePR)

	mux := http.NewServeMux()
	handlerTeam.ReqisterRoutes(mux)
	handlerPR.ReqisterRoutes(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	gshutdown := make(chan struct{})
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("error starting server%v", err)
		}
	}()

	log.Println("Server started: ", server.Addr)
	<-gshutdown
}
