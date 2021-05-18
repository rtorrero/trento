package agent

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strconv"

	consulAgent "github.com/hashicorp/consul/agent"
	configAgent "github.com/hashicorp/consul/agent/config"
	"github.com/hashicorp/consul/agent/connect"
	"github.com/hashicorp/consul/sdk/freeport"
	"github.com/hashicorp/go-uuid"
)

func NodeID() string {
	id, err := uuid.GenerateUUID()
	if err != nil {
		panic(err)
	}
	return id
}

func randomPortsSource() (data string) {
	ports := freeport.MustTake(5)

	return `
		ports = {
			dns = ` + strconv.Itoa(ports[0]) + `
			http = ` + strconv.Itoa(8500) + `
			https = ` + strconv.Itoa(-1) + `
			serf_lan = ` + strconv.Itoa(ports[1]) + `
			serf_wan = ` + strconv.Itoa(ports[2]) + `
			server = ` + strconv.Itoa(ports[3]) + `
			grpc = ` + strconv.Itoa(ports[4]) + `
		}
	`
}

func TestConfigHCL(nodeID string) string {
	return fmt.Sprintf(`
		bind_addr = "127.0.0.1"
		advertise_addr = "127.0.0.1"
		datacenter = "dc1"
		bootstrap = false
		bootstrap_expect = 1
		server = true
		node_id = "%[1]s"
		node_name = "Node-%[1]s"
		connect {
			enabled = true
			ca_config {
				cluster_id = "%[2]s"
			}
		}
		performance {
			raft_multiplier = 1
		}`, nodeID, connect.TestClusterID,
	)
}

func NewConsulAgent(hcl string) (*consulAgent.Agent, error) {
	portsConfig := randomPortsSource()
	log.Println(portsConfig)
	d := filepath.ToSlash("./consul-agent-data")
	hclDataDir := fmt.Sprintf(`data_dir = "%s"`, d)
	testHCLConfig := TestConfigHCL(NodeID())

	loader := func(source configAgent.Source) (*configAgent.RuntimeConfig, []string, error) {
		opts := configAgent.BuilderOpts{
			HCL: []string{testHCLConfig, portsConfig, hcl, hclDataDir},
		}
		overrides := []configAgent.Source{
			configAgent.FileSource{
				Name:   "test-overrides",
				Format: "hcl",
				Data:   ""},
			configAgent.DefaultConsulSource(),
			configAgent.DevConsulSource(),
		}
		cfg, warnings, err := configAgent.Load(opts, source, overrides...)
		if cfg != nil {
			cfg.Telemetry.Disable = true
		}
		return cfg, warnings, err
	}
	bd, err := consulAgent.NewBaseDeps(loader, log.Writer())

	if err != nil {
		return nil, fmt.Errorf("failed to create base deps: %w", err)
	}

	return consulAgent.New(bd)
}

func (a *Agent) startConsulAgent() error {
	go a.consulAgent.Start(context.Background())
	defer a.stopConsulAgent()
	for {
		select {
		case <-a.ctx.Done():
			return nil
		}
	}
}

func (a *Agent) stopConsulAgent() {
	log.Println("Stopping consul-agent")
	a.consulAgent.Leave()
	a.consulAgent.ShutdownAgent()
	a.consulAgent.ShutdownEndpoints()
}
