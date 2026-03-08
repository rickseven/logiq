package main

import (
	"flag"
	"log"

	"github.com/rickseven/logiq/internal/interfaces/mcp"
	"github.com/rickseven/logiq/internal/infrastructure/plugin"
	"github.com/rickseven/logiq/plugins/flutter"
	"github.com/rickseven/logiq/plugins/vue"
)

func main() {
	port := flag.Int("port", 8080, "Port to run the MCP server on")
	flag.Parse()

	plugin.Register(flutter.New())
	plugin.Register(vue.New())

	server := mcp.NewServer(*port)
	if err := server.Start(); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
