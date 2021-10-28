package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/trento/agent/collector"
	"github.com/trento-project/trento/internal/cloud"
	"github.com/trento-project/trento/internal/consul"
	"github.com/trento-project/trento/internal/hosts"
	"github.com/trento-project/trento/version"
	"github.com/zcalusic/sysinfo"
)

const HostDiscoveryId string = "host_discovery"

type HostDiscovery struct {
	id        string
	discovery BaseDiscovery
}

type DiscoveredHost struct {
	HostIpAddresses []string
	HostName        string
	SlesVersion     string
	CPUCount        int
	SocketCount     int
	TotalMemory     int
	CloudProvider   string
	AgentVersion    string
}

// SLES version
// number of CPUs per host
// number of sockets per host
// amount of memory (RAM) per host
// cloud provider / bare metal environment (AWS, Azure, Google, private data center etc - where easy identifiable) per host

func NewHostDiscovery(consulClient consul.Client, collectorClient collector.Client) HostDiscovery {
	d := HostDiscovery{}
	d.id = HostDiscoveryId
	d.discovery = NewDiscovery(consulClient, collectorClient)
	return d
}

func (h HostDiscovery) GetId() string {
	return h.id
}

// Execute one iteration of a discovery and store the result in the Consul KVStore.
func (h HostDiscovery) Discover() (string, error) {

	var si sysinfo.SysInfo

	si.GetSysInfo()

	data, err := json.MarshalIndent(&si, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))

	ipAddresses, err := getHostIpAddresses()
	if err != nil {
		return "", err
	}

	metadata := hosts.Metadata{
		HostIpAddresses: ipAddresses,
	}
	err = metadata.Store(h.discovery.consulClient)
	if err != nil {
		return "", err
	}

	// SLES version
	// number of CPUs per host
	// number of sockets per host
	// amount of memory (RAM) per host
	// cloud provider / bare metal environment (AWS, Azure, Google, private data center etc - where easy identifiable) per host

	// TODO: this needs to be redesigned to capture what we are discovering about a Host

	ipAddressesList, err := getHostIpAddressesList()

	if err != nil {
		log.Debugf("Error while getting HostIpAddresses: %s", err)
		return "", err
	}

	cloudProvider, err := cloud.IdentifyCloudProvider()

	if err != nil {
		log.Debugf("Error while identifying cloud provider: %s", err)
		return "", err
	}

	host := DiscoveredHost{
		ipAddressesList,
		h.discovery.host,
		si.OS.Version,
		int(si.CPU.Cpus) * int(si.CPU.Cores), // What do they expect with CPU count?
		int(si.CPU.Cpus),
		int(si.Memory.Size),
		cloudProvider,
		version.Version,
	}

	err = h.discovery.collectorClient.Publish(h.id, host)
	if err != nil {
		log.Debugf("Error while sending host discovery to data collector: %s", err)
		return "", err
	}

	return fmt.Sprintf("Host with name: %s successfully discovered", h.discovery.host), nil
}

func getHostIpAddresses() (string, error) {
	ips, err := getHostIpAddressesList()

	if err != nil {
		return "", err
	}

	return strings.Join(ips, ","), nil
}

func getHostIpAddressesList() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	ipAddrList := make([]string, 0)

	for _, inter := range interfaces {
		addrs, err := inter.Addrs()
		if err != nil {
			continue
		}

		for _, ipaddr := range addrs {
			ipv4Addr, _, _ := net.ParseCIDR(ipaddr.String())
			ipAddrList = append(ipAddrList, ipv4Addr.String())
		}
	}

	return ipAddrList, nil
}

func getHostSLESRelease() (string, error) {
	return "banana", nil
}
