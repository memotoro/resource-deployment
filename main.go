package main

import (
	"log"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/memotoro/seldonio-resource-deployment/clients"
	"github.com/memotoro/seldonio-resource-deployment/readers"
	"github.com/memotoro/seldonio-resource-deployment/resources"
)

var (
	host         = kingpin.Flag("host-api", "Hostname or IP address for the kubernetes API").Default("192.168.99.100").String()
	port         = kingpin.Flag("port-api", "Port number for the kubernetes API").Default("8443").String()
	timeout      = kingpin.Flag("api-timeout", "Timeout in time.Duration for http calls to Kubernetes API").Default("5s").Duration()
	token        = kingpin.Flag("api-token", "API kubernetes token with the right permissions for the API").Default("").String()
	resourceFile = kingpin.Flag("resource-file", "Location of file with the resource to be created. Could be a file path or http url").Default("").String()
	waitingTime  = kingpin.Flag("waiting-time", "Time in seconds to wait between API call").Default("5s").Duration()
	maxAttempts  = kingpin.Flag("max-attempts", "Maximum number of re-tries before terminating the task").Default("20").Int()
	namespace    = kingpin.Flag("namespace", "Namespace for kubernetes").Default("seldon-system").String()
	verbose      = kingpin.Flag("verbose", "Verbose logging").Default("false").Bool()
)

func main() {
	kingpin.Parse()

	if *token == "" {
		log.Fatalf("--api-token is required")
	}

	if *resourceFile == "" {
		log.Fatalf("--resource-file is required")
	}

	basicAuthClient := clients.NewHTTPClient(*host, *port, *timeout, clients.Auth{Token: *token}, *verbose)

	basicClient := clients.NewHTTPClient(*host, *port, *timeout, clients.Auth{Token: ""}, *verbose)

	resourceContent, err := readers.ReadContentFile(basicClient, *resourceFile)
	if err != nil {
		log.Fatal(err)
	}
	// First call to see if the resource exists already in the cluster
	seldonDeployment, err := resources.GetResourceStatus(basicAuthClient, resourceContent, *namespace)
	if err != nil {
		log.Fatal(err)
	}

	if seldonDeployment != nil {
		log.Printf("Resource Name [%v] - Resource Kind [%v] is already in the cluster.", seldonDeployment.TypeMeta.Kind, seldonDeployment.ObjectMeta.Name)
	} else {
		seldonDeployment, err := resources.CreateResource(basicAuthClient, resourceContent, *namespace)
		if err != nil {
			log.Fatal(err)
		}

		if seldonDeployment != nil {
			log.Printf("Resource Name [%v] - Resource Kind [%v] created.", seldonDeployment.TypeMeta.Kind, seldonDeployment.ObjectMeta.Name)
		}
	}

	counter := 1

	time.Sleep(*waitingTime)

	for counter <= *maxAttempts {
		seldonDeployment, err := resources.GetResourceStatus(basicAuthClient, resourceContent, *namespace)
		if err != nil {
			log.Fatal(err)
		}

		if seldonDeployment == nil {
			counter++
		}

		switch seldonDeployment.Status.State {
		case "Available":
			log.Printf("Resource Name [%v] - Resource State [%v]. Waiting %v before deleting resource", seldonDeployment.ObjectMeta.Name, seldonDeployment.Status.State, *waitingTime)
			time.Sleep(*waitingTime)

			status, err := resources.DeleteResource(basicAuthClient, resourceContent, *namespace)
			if err != nil {
				log.Fatal(err)
			}

			if status.Status == "Success" {
				log.Printf("Resource Name [%v] was deleted", seldonDeployment.ObjectMeta.Name)
			}

			counter = *maxAttempts + 1
		default:
			log.Printf("Resource Name [%v] - Resource State [%v]. Waiting %v before attempt %d out of %d", seldonDeployment.ObjectMeta.Name, seldonDeployment.Status.State, *waitingTime, counter, *maxAttempts)
			counter++
			time.Sleep(*waitingTime)
		}
	}
}
