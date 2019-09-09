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

// Get all DD monitors from source and target accounts
func getAllMonitors(sourceClient *datadog.Client, targetClient *datadog.Client) ([]datadog.Monitor, []datadog.Monitor, error) {
	sourceMonitors, err := sourceClient.GetMonitors()
	if err != nil {
		return nil, nil, err
	}

	targetMonitors, err := targetClient.GetMonitors()

	return sourceMonitors, targetMonitors, err
}

func isMonitorExists(monitor datadog.Monitor, targetMonitors []datadog.Monitor) bool {
	for _, targetMonitor := range targetMonitors {
		if targetMonitor.GetName() == monitor.GetName() {
			log.Fatalln("monitor already exists on the target account")
			return true
		}
	}

	return false
}

// Check if a monitor isn't already exits (by it's Name) on the target datadog account, and create monitor
func createMonitor(client *datadog.Client, targetMonitors []datadog.Monitor, monitor datadog.Monitor, wg *sync.WaitGroup) error {
	defer wg.Done()

	if !isMonitorExists(monitor, targetMonitors) {
		_, err := client.CreateMonitor(&monitor)
		if err != nil {
			return err
		}
		log.Println(monitor.GetName() + " Monitor has been created successfully")
	}

	return nil
}

func main() {
	sourceCreds, targetCreds := getCredentials()
	clientSource := datadog.NewClient(sourceCreds.ddAPI, sourceCreds.ddAPP)
	clientTarget := datadog.NewClient(targetCreds.ddAPI, targetCreds.ddAPP)
	sourceMonitors, targetMonitors, err := getAllMonitors(clientSource, clientTarget)
	if err != nil {
		log.Fatalf("Coudldn't load monitors for datadog account: %s\n", err)
	}

	var wg sync.WaitGroup

	// Iterate over monitor tags and look for for the tag in 'SharedTag'
	for _, monitor := range sourceMonitors {
		for _, tag := range monitor.Tags {
			if tag == SharedTag {
				wg.Add(1)
				go createMonitor(clientTarget, targetMonitors, monitor, &wg)
			}
			continue
		}
	}
	// Wait for all the monitors creation to complete
	wg.Wait()

	log.Println("Finished, good bye")
}
