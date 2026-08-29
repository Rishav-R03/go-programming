package main

import (
	"dynamicstory"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := flag.Int("port", 3000, "the port to start the dynamic story")
	filename := flag.String("file", "story.json", "the file with story and options")
	flag.Parse()

	fmt.Printf("Loading the story from %s...\n", *filename)
	f, err := os.Open(*filename)
	if err != nil {
		log.Fatalf("failed to open file : %v", err)
	}
	defer f.Close()
	story, err := dynamicstory.JsonStory(f)
	if err != nil {
		log.Fatalf("Failed to parse story json: %v", err)
	}
	h := dynamicstory.NewHandler(story)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Starting server: http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, h))
}
