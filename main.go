package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/bbars/bridget/internal/event"
	"github.com/bbars/bridget/internal/handler"
)

func main() {
	bind := "0.0.0.0:0"
	flag.StringVar(&bind, "b", bind, "http bind address")
	flag.Parse()

	subs := event.NewSubscriptions()
	http.Handle("GET /subscribe/{"+handler.WildcardPath+"...}", handler.Subscribe{Subscriptions: subs})
	http.Handle("/emit/{"+handler.WildcardPath+"...}", handler.Emit{Subscriptions: subs})
	http.Handle("GET /manage/list", handler.ManageList{Subscriptions: subs})
	http.Handle("POST /manage/kick/{"+handler.WildcardId+"}", handler.ManageKick{Subscriptions: subs})

	if lis, err := net.Listen("tcp", bind); err != nil {
		log.Fatalf("failed to listen socket: %v", err)
	} else {
		log.Printf("listening on %s\n", lis.Addr())
		log.Printf("request %q or %q\n",
			"GET /subscribe/{"+handler.WildcardPath+"...}",
			"/emit/{"+handler.WildcardPath+"...}",
		)
		if err = http.Serve(lis, http.DefaultServeMux); err != nil {
			log.Fatal(err)
		}
	}
}
