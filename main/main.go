package main

import (
	"log"
	"os"
	"sync"

	"github.com/AkinoKaede/naruse"
	"github.com/AkinoKaede/naruse/config"
	"github.com/AkinoKaede/naruse/dispatcher"

	"github.com/urfave/cli/v2"
)

type MapPortDispatcher map[int]*dispatcher.Dispatcher

type SyncMapPortDispatcher struct {
	sync.Mutex
	Map MapPortDispatcher
}

func NewSyncMapPortDispatcher() *SyncMapPortDispatcher {
	return &SyncMapPortDispatcher{Map: make(MapPortDispatcher)}
}

var (
	groupWG         sync.WaitGroup
	mPortDispatcher = NewSyncMapPortDispatcher()
)

func listenDispatcher(d *dispatcher.Dispatcher) error {
	mPortDispatcher.Lock()

	if _, ok := mPortDispatcher.Map[d.Port]; !ok {
		mPortDispatcher.Map[d.Port] = d
	}

	mPortDispatcher.Unlock()

	ch := make(chan error, 1)
	go func() {
		err := d.Listen()
		ch <- err
	}()

	return <-ch
}

func listenDispatchers(dispatchers []*dispatcher.Dispatcher) {
	mPortDispatcher.Lock()
	for _, d := range dispatchers {
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
	mPortDispatcher.Unlock()
}

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Show current version of Naruse.",
	}

	app := &cli.App{
		Name:    "Naruse",
		Version: naruse.Version(),
		Usage:   naruse.Usage(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c", "conf"},
				Usage:    "Config file for Naruse.",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "suppress-timestamps",
				Usage: "Do not include timestamps in log",
			},
		},
		Action: func(c *cli.Context) error {
			if c.Bool("suppress-timestamps") {
				log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
			}

			path := c.String("config")

			go signalHandler(path)

			config, err := config.BuildConfig(path)
			if err != nil {
				return err
			}

			dispatchers, err := config.Build()
			if err != nil {
				return err
			}
			listenDispatchers(dispatchers)
			groupWG.Wait()
			return nil
		},
		ExitErrHandler: func(_ *cli.Context, err error) {
			log.Println(err)
		},
	}

	app.Run(os.Args)
}
