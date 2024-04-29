package main

import (
	"log"
	"os"

	"github.com/SameerJadav/go-api/internal/database"
	"github.com/SameerJadav/go-api/internal/server"
)

func main() {
	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime)

	db, err := database.New()
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	server, err := server.NewServer(db)
	if err != nil {
		errorLog.Fatal(err)
	}

	infoLog.Printf("starting server on http://localhost%s", server.Addr)
	if err = server.ListenAndServe(); err != nil {
		errorLog.Fatal(err)
	}
}
