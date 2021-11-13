package main

import (
	"log"

	"github.com/AkinoKaede/naruse/config"
	"github.com/AkinoKaede/naruse/dispatcher"
)

func ReloadConfig(path string) {
	log.Println("reloading config")
	mPortDispatcher.Lock()
	defer mPortDispatcher.Unlock()

	config, err := config.BuildConfig(path)
	if err != nil {
		log.Printf("failed to reload config: %v", err)
		return
	}

	dispatchers, err := config.Build()
	if err != nil {
		log.Printf("failed to build dispatchers: %v", err)
		return
	}

	newConfigPortSet := make(map[int]struct{})
	for _, d := range dispatchers {
		newConfigPortSet[d.Port] = struct{}{}
		t, ok := mPortDispatcher.Map[d.Port]

		switch {
		case ok && t.ListenAddr == d.ListenAddr:
			t.UpdateValidator(d.Validator)
		case ok && t.ListenAddr != d.ListenAddr:
			t.Close()
			fallthrough
		default:
			groupWG.Add(1)
			go func(d *dispatcher.Dispatcher) {
				err := listenDispatcher(d)
				if err != nil {
					mPortDispatcher.Lock()
					// error but listening
					if _, ok := mPortDispatcher.Map[d.Port]; ok {
						log.Fatalln(err)
					}
					mPortDispatcher.Unlock()
				}
				groupWG.Done()
			}(d)
		}

	}

	for port := range mPortDispatcher.Map {
		if _, ok := newConfigPortSet[port]; !ok {
			t := mPortDispatcher.Map[port]
			delete(mPortDispatcher.Map, port)
			t.Close()
		}
	}

	log.Println("reloaded config")
}
