package main

import (
	"flag"
	"github.com/bingoohuang/godaemon"
	"log"
	"os"
	"syscall"
	"time"
)

var signal = flag.String("s", "", `Send signal to the daemon:
  stop — shutdown`)

const (
	logFileName = "sample.log"
	pidFileName = "sample.pid"
)

func main() {
	flag.Parse()
	godaemon.AddCommand(godaemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)

	cntxt := &godaemon.Context{
		PidFileName: pidFileName,
		PidFilePerm: 0o644,
		LogFileName: logFileName,
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

	setupLog()

	go worker()

	err = godaemon.ServeSignals()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
	log.Println("daemon terminated")
}

func setupLog() {
	lf, err := NewLogFile(logFileName, os.Stderr)
	if err != nil {
		log.Fatalf("Unable to create log file: %s", err.Error())
	}

	log.SetOutput(lf)
	// rotate log every 30 seconds.
	rotateLogSignal := time.Tick(30 * time.Second)
	go func() {
		for {
			<-rotateLogSignal
			if err := lf.Rotate(); err != nil {
				log.Fatalf("Unable to rotate log: %s", err.Error())
			}
		}
	}()
}

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

func worker() {
LOOP:
	for {
		// spam to log every one second (as payload).
		log.Print("+ ", time.Now().Unix())
		time.Sleep(time.Second)
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
