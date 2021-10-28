package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/consul-template/manager"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/trento-project/trento/agent/collector"
	"github.com/trento-project/trento/agent/discovery"
	"github.com/trento-project/trento/internal/consul"
	"github.com/trento-project/trento/internal/hosts"
	"github.com/trento-project/trento/version"
)

const trentoAgentCheckId = "trentoAgent"

type Agent struct {
	config         *Config
	discoveries    []discovery.Discovery
	consul         consul.Client
	ctx            context.Context
	ctxCancel      context.CancelFunc
	templateRunner *manager.Runner
}

type Config struct {
	InstanceName    string
	ConsulConfigDir string
	DiscoveryPeriod time.Duration
	CollectorConfig *collector.Config
}

// NewAgent returns a new instance of Agent with the given configuration
func NewAgent(config *Config) (*Agent, error) {
	consulClient, err := consul.DefaultClient()
	if err != nil {
		return nil, errors.Wrap(err, "could not create a Consul client")
	}

	collectorClient, err := collector.NewCollectorClient(config.CollectorConfig)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a collector client")
	}

	templateRunner, err := NewTemplateRunner(config.ConsulConfigDir)
	if err != nil {
		return nil, errors.Wrap(err, "could not create the consul template runner")
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	agent := &Agent{
		config:    config,
		ctx:       ctx,
		ctxCancel: ctxCancel,
		consul:    consulClient,
		discoveries: []discovery.Discovery{
			// discovery.NewClusterDiscovery(consulClient, collectorClient),
			// discovery.NewSAPSystemsDiscovery(consulClient, collectorClient),
			// discovery.NewCloudDiscovery(consulClient, collectorClient),
			// discovery.NewSubscriptionDiscovery(consulClient, collectorClient),
			discovery.NewHostDiscovery(consulClient, collectorClient),
		},
		templateRunner: templateRunner,
	}
	return agent, nil
}

// Start the Agent which includes the registration against Consul Agent
func (a *Agent) Start() error {
	log.Println("Registering the agent service with Consul...")
	err := a.registerConsulService()
	if err != nil {
		return errors.Wrap(err, "could not register consul service")
	}
	log.Println("Consul service registered.")

	defer func() {
		log.Println("De-registering the agent service with Consul...")
		err := a.consul.Agent().ServiceDeregister(a.config.InstanceName)
		if err != nil {
			log.Println("An error occurred while trying to deregisterConsulService the agent service with Consul:", err)
		} else {
			log.Println("Consul service de-registered.")
		}
	}()

	var wg sync.WaitGroup

	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		log.Println("Starting Discover loop...")
		defer wg.Done()
		a.startDiscoverTicker()
		log.Println("Discover loop stopped.")
	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Println("Starting consul-template loop...")
		defer wg.Done()
		a.startConsulTemplate()
		log.Println("consul-template loop stopped.")
	}(&wg)

	storeAgentMetadata(a.consul, version.Version)

	wg.Wait()

	return nil
}

func (a *Agent) Stop() {
	a.ctxCancel()
}

func (a *Agent) registerConsulService() error {
	var err error

	discoveryTTL := a.config.DiscoveryPeriod * 2
	consulService := &consulApi.AgentServiceRegistration{
		ID:   a.config.InstanceName,
		Name: "trento-agent",
		Tags: []string{"trento"},
		Checks: consulApi.AgentServiceChecks{
			&consulApi.AgentServiceCheck{
				CheckID: trentoAgentCheckId,
				Name:    "Trento Agent",
				Notes:   "Reports the health of the Trento Agent itself",
				TTL:     discoveryTTL.String(),
				Status:  consulApi.HealthWarning,
			},
		},
	}

	err = a.consul.Agent().ServiceRegister(consulService)
	if err != nil {
		return errors.Wrap(err, "could not register the agent service with Consul")
	}

	return nil
}

// Start a Ticker loop that will iterate over the hardcoded list of Discovery backends
// and execute them. The initial run will happen relatively quickly after Agent launch
// subsequent runs are done with a 15 minute delay. The effectiveness of the discoveries
// is reported back in the "discover_cluster" Service in consul under a TTL of 60 minutes
func (a *Agent) startDiscoverTicker() {
	tick := func() {
		var output []string
		status := consulApi.HealthPassing

		for _, d := range a.discoveries {
			result, err := d.Discover()
			if err != nil {
				result = fmt.Sprintf("Error while running discovery '%s': %s", d.GetId(), err)
				status = consulApi.HealthCritical

				log.Errorln(result)
			}
			output = append(output, result)
		}

		if err := a.consul.Agent().UpdateTTL(trentoAgentCheckId, strings.Join(output, "\n\n"), status); err != nil {
			log.Errorln("An error occurred while trying to update TTL with Consul:", err)
		}
	}

	interval := a.config.DiscoveryPeriod
	repeat(tick, interval, a.ctx)
}

func repeat(tick func(), interval time.Duration, ctx context.Context) {
	// run the first tick immediately
	tick()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			tick()
		case <-ctx.Done():
			return
		}
	}
}

func storeAgentMetadata(client consul.Client, version string) error {
	metadata := hosts.Metadata{
		AgentVersion: version,
	}

	err := metadata.Store(client)
	if err != nil {
		return err
	}

	return nil
}
