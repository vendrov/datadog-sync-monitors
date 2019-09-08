package main

import (
	"log"
	"os"
	"sync"

	"github.com/zorkian/go-datadog-api"
)

// SharedTag is the tag that Only monitors who own it will be synced among accounts"
const SharedTag = "devops:common"

type credentials struct {
	ddAPI string
	ddAPP string
}

func getCredentials() (sourceCreds, targetCreds credentials) {
	ddAPI, isSet := os.LookupEnv("DD_API")
	if isSet != true {
		log.Fatalf("DD_API can't be found")
	}

	ddApp, isSet := os.LookupEnv("DD_APP")
	if isSet != true {
		log.Fatalf("DD_APP can't be found")
	} else {
		sourceCreds = credentials{ddAPI, ddApp}
	}

	ddAPIT, isSet := os.LookupEnv("DD_APIT")
	if isSet != true {
		log.Fatalf("DD_APIT can't be found")
	}

	ddAppT, isSet := os.LookupEnv("DD_APPT")
	if isSet != true {
		log.Fatalf("DD_APPT can't be found")
	} else {
		targetCreds = credentials{ddAPIT, ddAppT}
	}

	return
}

func createMonitor(client *datadog.Client, monitor datadog.Monitor, wg *sync.WaitGroup) error {
	defer wg.Done()
	_, err := client.CreateMonitor(&monitor)
	if err != nil {
		return err
	}

	log.Println(monitor.GetName() + " Monitor has been created successfully")
	return nil
}

func main() {
	sourceCreds, targetCreds := getCredentials()
	clientSource := datadog.NewClient(sourceCreds.ddAPI, sourceCreds.ddAPP)
	clientTarget := datadog.NewClient(targetCreds.ddAPI, targetCreds.ddAPP)
	var wg sync.WaitGroup

	monitors, err := clientSource.GetMonitors()
	if err != nil {
		log.Fatalf("fatal: %s\n", err)
	}

	// Iterate over monitor tags and look for
	for _, monitor := range monitors {
		for _, tag := range monitor.Tags {
			if tag == SharedTag {
				wg.Add(1)
				go createMonitor(clientTarget, monitor, &wg)

			}
			continue
		}
	}
	// Wait for all the monitors creation to complete
	wg.Wait()

	log.Println("Finished, good bye")
}
