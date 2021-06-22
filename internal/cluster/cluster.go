package cluster

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	// Reusing the Prometheus Ha Exporter cibadmin xml parser here
	"github.com/ClusterLabs/ha_cluster_exporter/collector/pacemaker/cib"
	"github.com/ClusterLabs/ha_cluster_exporter/collector/pacemaker/crmmon"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/trento-project/trento/internal"
)

const (
	cibAdmPath             string = "/usr/sbin/cibadmin"
	crmmonAdmPath          string = "/usr/sbin/crm_mon"
	corosyncKeyPath        string = "/etc/corosync/authkey"
	clusterNameProperty    string = "cib-bootstrap-options-cluster-name"
	stonithEnabled         string = "cib-bootstrap-options-stonith-enabled"
	stonithResourceMissing string = "notconfigured"
	stonithAgent           string = "stonith:"
	sbdFencingAgentName    string = "external/sbd"
)

type Cluster struct {
	Cib    cib.Root    `mapstructure:"cib,omitempty"`
	Crmmon crmmon.Root `mapstructure:"crmmon,omitempty"`
	SBD    SBD         `mapstructure:"sbd,omitempty"`
	Id     string      `mapstructure:"id"`
	Name   string      `mapstructure:"name,omitempty"`
}

func NewCluster() (Cluster, error) {
	var cluster = Cluster{}

	cibParser := cib.NewCibAdminParser(cibAdmPath)

	cibConfig, err := cibParser.Parse()
	if err != nil {
		return cluster, err
	}

	cluster.Cib = cibConfig

	crmmonParser := crmmon.NewCrmMonParser(crmmonAdmPath)

	crmmonConfig, err := crmmonParser.Parse()
	if err != nil {
		return cluster, err
	}

	cluster.Crmmon = crmmonConfig

	// Set MD5-hashed key based on the corosync auth key
	cluster.Id, err = getCorosyncAuthkey(corosyncKeyPath)
	if err != nil {
		return cluster, err
	}

	// Handle not named clusters
	/*for _, prop := range cibConfig.Configuration.CrmConfig.ClusterProperties {
		if prop.Id == clusterNameProperty {
			cluster.Name = prop.Value
		}
	}*/

	//if cluster.Name == "" {
	cluster.Name = getName()
	//}

	if cluster.IsFencingSBD() {
		sbdData, err := NewSBD(cluster.Id, SBDPath, SBDConfigPath)
		if err != nil {
			return cluster, err
		}

		cluster.SBD = sbdData
	}

	return cluster, nil
}

func getCorosyncAuthkey(corosyncKeyPath string) (string, error) {
	kp, err := internal.Md5sum(corosyncKeyPath)
	log.Println(kp)
	return kp, err
}

func getName() string {
	words := flag.Int("words", 2, "The number of words in the pet name")
	separator := flag.String("separator", "-", "The separator between words in the pet name")
	rand.Seed(time.Now().UTC().UnixNano())

	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	return petname.Generate(*words, *separator)
}

func (c *Cluster) IsDc() bool {
	host, _ := os.Hostname()

	for _, nodes := range c.Crmmon.Nodes {
		if nodes.Name == host {
			return nodes.DC
		}
	}

	return false
}

func (c *Cluster) IsFencingEnabled() bool {
	for _, prop := range c.Cib.Configuration.CrmConfig.ClusterProperties {
		if prop.Id == stonithEnabled {
			b, err := strconv.ParseBool(prop.Value)
			if err != nil {
				return false
			}
			return b
		}
	}

	return false
}

func (c *Cluster) FencingResourceExists() bool {
	f := c.FencingType()

	return f != stonithResourceMissing
}

func (c *Cluster) FencingType() string {
	for _, resource := range c.Crmmon.Resources {
		if strings.HasPrefix(resource.Agent, stonithAgent) {
			return strings.Split(resource.Agent, ":")[1]
		}
	}
	return stonithResourceMissing
}

func (c *Cluster) IsFencingSBD() bool {
	f := c.FencingType()

	return f == sbdFencingAgentName
}
