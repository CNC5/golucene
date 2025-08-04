package main

import (
	"golucene/httpComponents"
	"log"
	"net/http"
	"sync"
)

func main() {
	server := httpComponents.NewIndicesServer()
	var waitGroup sync.WaitGroup
	log.Println("Server start")
	waitGroup.Add(2)
	go func() {
		http.ListenAndServe(":9999", server.IndicesServerMuxPointer)
		waitGroup.Done()
	}()
	go func() {
		http.ListenAndServe("localhost:9001", server.MetricsServerMuxPointer)
		waitGroup.Done()
	}()
	waitGroup.Wait()
}
