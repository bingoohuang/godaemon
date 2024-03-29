package main

import (
	"flag"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/bingoohuang/godaemon"
)

var signal = flag.String("s", "", `Send signal to the daemon:
  quit — graceful shutdown
  stop — fast shutdown
  reload — reloading the configuration file`)

func main() {
	flag.Parse()
	godaemon.AddCommand(godaemon.StringFlag(signal, "quit"), syscall.SIGQUIT, termHandler)
	godaemon.AddCommand(godaemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	godaemon.AddCommand(godaemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadHandler)

	cntxt := &godaemon.Context{
		PidFileName: "sample.pid",
		PidFilePerm: 0o644,
		LogFileName: "sample.log",
		LogFilePerm: 0o640,
		WorkDir:     "./",
		Umask:       0o27,
		Args:        []string{"[go-daemon sample]"},
	}

	if len(godaemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon: %s", err.Error())
		}
		godaemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Println("- - - - - - - - - - - - - - -")
	log.Println("daemon started")

	go worker()

	err = godaemon.ServeSignals()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}

	log.Println("daemon terminated")
}

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

func worker() {
LOOP:
	for {
		time.Sleep(time.Second) // this is work to be done by worker.
		select {
		case <-stop:
			break LOOP
		default:
		}
	}
	done <- struct{}{}
}

func termHandler(sig os.Signal) error {
	log.Println("terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return godaemon.ErrStop
}

func reloadHandler(sig os.Signal) error {
	log.Println("configuration reloaded")
	return nil
}
