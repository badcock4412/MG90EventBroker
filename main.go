package main

import (
	"github.com/badcock4412/SSEBroker"
	"github.com/badcock4412/mg90"
	"fmt"
	"log"
	"time"
	"net/http"
	"runtime"
)

func main() {

	broker := SSEBroker.NewServer()

	go func() {
		for {
			MG,err := MG90.NewMG90("172.22.0.1")
			
			if err != nil {
				log.Println("Could not find MG-90")
				time.Sleep( 5 * time.Second )
				continue
			}
			
			log.Printf("Found MG-90 %s\n",MG.GetId())
			
			go MG.ListenBeacon(15000)
			
			switch runtime.GOOS {
				case "windows":
					go MG.StartMonitor(3 * time.Second)
				case "linux":
					go MG.PingMonitor()
			}	
			
			broker.NewSubscriberNotifier <- []byte(fmt.Sprintf("{ \"event\": \"hello\", \"message\": \"%s\" }",MG.GetId()))
			
			keepGoing := true
			
			for keepGoing {
				select {
					case e := <-MG.Events.LostConnection:
						log.Printf("Lost MG-90 %s: %s\n",MG.GetId(), e)
						broker.Notifier <- []byte(fmt.Sprintf("{ \"event\": \"goodbye\", \"message\":\"%s\"}",e))
						keepGoing = false
					case e := <-MG.Events.GPIO:
						log.Printf("Detected change in GPIO%d\n",e.Channel)
						t := time.Now()
						broker.Notifier <- []byte(fmt.Sprintf("{ \"event\": \"gpio\", \"gpio\":%d, \"value\":%d, \"time\":\"%s\"}",e.Channel,e.NewValue,t.Format("2006-01-02T15:04:05.000-07:00")))
				}
			}
			log.Printf("Closing MG-90 %s\n",MG.GetId())
			MG.Close()
		}
	}()

	log.Fatal("HTTP server error: ", http.ListenAndServe("localhost:3000", broker))
}