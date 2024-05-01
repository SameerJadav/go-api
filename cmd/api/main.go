package main

import (
	"github.com/SameerJadav/go-api/internal/database"
	"github.com/SameerJadav/go-api/internal/logger"
	"github.com/SameerJadav/go-api/internal/server"
)

func main() {
	db, err := database.New()
	if err != nil {
		logger.Error.Fatalln(err)
	}
	defer db.Close()

	server, err := server.NewServer(db)
	if err != nil {
		logger.Error.Fatalln(err)
	}

	logger.Info.Printf("starting server on http://localhost%s", server.Addr)
	if err = server.ListenAndServe(); err != nil {
		logger.Error.Fatalln(err)
	}
}
