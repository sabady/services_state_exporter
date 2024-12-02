package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics
var (
	serviceDesiredState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "docker_swarm_service_desired_state",
			Help: "Desired state of Docker Swarm services (1 for active, 0 for inactive)",
		},
		[]string{"service_name"},
	)

	serviceRunningCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "docker_swarm_service_running_count",
			Help: "Number of running tasks for each Docker Swarm service",
		},
		[]string{"service_name"},
	)
)

func init() {
	prometheus.MustRegister(serviceDesiredState)
	prometheus.MustRegister(serviceRunningCount)
}

func monitorServices(cli *client.Client, interval time.Duration) {
	for {
		// Get a list of all services
		services, err := cli.ServiceList(context.Background(), types.ServiceListOptions{})
		if err != nil {
			log.Printf("Error retrieving services: %v\n", err)
			time.Sleep(interval)
			continue
		}

		for _, service := range services {
			// Reset metrics for this service
			serviceDesiredState.WithLabelValues(service.Spec.Name).Set(0)
			serviceRunningCount.WithLabelValues(service.Spec.Name).Set(0)

			// Set desired state metric
			if service.Spec.TaskTemplate.Placement != nil {
				serviceDesiredState.WithLabelValues(service.Spec.Name).Set(1.0)
			}

			// Create filter for tasks in the service
			taskFilters := filters.NewArgs()
			taskFilters.Add("service", service.ID)

			// Get the tasks for the service
			tasks, err := cli.TaskList(context.Background(), types.TaskListOptions{
				Filters: taskFilters,
			})
			if err != nil {
				log.Printf("Error retrieving tasks for service %s: %v\n", service.Spec.Name, err)
				continue
			}

			// Count running tasks
			runningCount := 0
			for _, task := range tasks {
				if task.Status.State == "running" {
					runningCount++
				}
			}

			// Update the running tasks count metric for this service
			serviceRunningCount.WithLabelValues(service.Spec.Name).Set(float64(runningCount))
		}

		time.Sleep(interval)
	}
}

func main() {
	// Initialize Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error initializing Docker client: %v\n", err)
	}

	// Monitoring interval
	interval := 10 * time.Second
	go monitorServices(cli, interval)

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	port := os.Getenv("EXPORTER_PORT")
	if port == "" {
		port = "9180"
	}
	log.Printf("Starting server on :%s/metrics\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
