package main

import (
	"sync"

	"jeremiah.smtp/client"
	"jeremiah.smtp/server"
)

func main() {
	serverReady := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(3)

	
	go server.SetupHTTPServer(&wg)
	go server.SetupSMTPServer(&wg, &serverReady)
	go client.RunClient(&wg, &serverReady)

	wg.Wait()

	println("Ran all commands successfully")
}
