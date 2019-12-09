package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/kardianos/service"
)

type program struct {
	exit    chan struct{}
	service service.Service
	Config  *service.Config
	cmd     *exec.Cmd
}

func main() {

	versionFlag := flag.Bool("version", false, "Print current Version")

	svcFlag := flag.String("service", "", "Control the system service")

	flag.Parse()

	if *versionFlag {
		fmt.Println("Version: 1.0\n")
		os.Exit(0)
	}
	/*
		configPath, err := getConfigPath()
		if err != nil {
			log.Fatal(err)
		}
		config, err := getConfig(configPath)
		if err != nil {
			log.Fatal(err)
		}
	*/

	svcConfig := &service.Config{
		Name:        "wremonservice",
		DisplayName: "Wre Monitor Service",
		Description: "Webriposte Monitor",
	}

	prg := &program{
		exit:   make(chan struct{}),
		Config: svcConfig,
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	prg.service = s

	errs := make(chan error, 5)
	//log.Error(errs)

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}
	err = s.Run()
	if err != nil {
		log.Panic(err)
	}
}

// Start should not block. Do the actual work async.
func (p *program) Start(s service.Service) error {
	
	/*-----------------------------------------------------------*/
  
	MetricsRun()

	/*-----------------------------------------------------------*/

	return nil
}

//Stop - service stop function
func (p *program) Stop(s service.Service) error {
	close(p.exit)
	
	if service.Interactive() {
		os.Exit(0)
	}
	return nil
}
