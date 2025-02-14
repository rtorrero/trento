package services

import (
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/trento/test/helpers"
	"github.com/trento-project/trento/web/entities"
	"github.com/trento-project/trento/web/models"
	"gorm.io/gorm"
)

func hostsFixtures() []entities.Host {
	return []entities.Host{
		{
			AgentID:       "1",
			Name:          "host1",
			ClusterID:     "cluster_id_1",
			ClusterName:   "cluster_1",
			ClusterType:   models.ClusterTypeHANAScaleOut,
			CloudProvider: "azure",
			IPAddresses:   pq.StringArray{"10.74.1.5"},
			SAPSystemInstances: []*entities.SAPSystemInstance{
				{
					AgentID:        "1",
					ID:             "sap_system_id_1",
					SID:            "DEV",
					InstanceNumber: "00",
				},
			},
			AgentVersion: "rolling1337",
			Heartbeat: &entities.HostHeartbeat{
				AgentID:   "1",
				UpdatedAt: time.Date(2020, 11, 01, 00, 00, 00, 0, time.UTC),
			},
			Tags: []*models.Tag{{
				Value:        "tag1",
				ResourceID:   "1",
				ResourceType: models.TagHostResourceType,
			}},
		},
		{
			AgentID:       "2",
			Name:          "host2",
			ClusterID:     "cluster_id_2",
			ClusterName:   "cluster_2",
			CloudProvider: "azure",
			ClusterType:   models.ClusterTypeUnknown,
			IPAddresses:   pq.StringArray{"10.74.1.10"},
			SAPSystemInstances: []*entities.SAPSystemInstance{
				{
					AgentID:        "2",
					ID:             "sap_system_id_2",
					SID:            "QAS",
					InstanceNumber: "10",
				},
			},
			AgentVersion: "stable",
			Heartbeat: &entities.HostHeartbeat{
				AgentID:   "2",
				UpdatedAt: time.Date(2020, 11, 01, 00, 00, 00, 0, time.UTC),
			},
			Tags: []*models.Tag{{
				Value:        "tag2",
				ResourceID:   "2",
				ResourceType: models.TagHostResourceType,
			}},
		},
	}
}

type HostsServiceTestSuite struct {
	suite.Suite
	db           *gorm.DB
	tx           *gorm.DB
	hostsService *hostsService
}

func TestHostsServiceTestSuite(t *testing.T) {
	suite.Run(t, new(HostsServiceTestSuite))
}

func (suite *HostsServiceTestSuite) SetupSuite() {
	suite.db = helpers.SetupTestDatabase(suite.T())

	suite.db.AutoMigrate(&entities.Host{}, &entities.HostHeartbeat{}, &entities.SAPSystemInstance{}, &models.Tag{})
	hosts := hostsFixtures()
	err := suite.db.Create(&hosts).Error
	suite.NoError(err)
}

func (suite *HostsServiceTestSuite) TearDownSuite() {
	suite.db.Migrator().DropTable(&entities.Host{},
		&entities.HostHeartbeat{},
		&entities.SAPSystemInstance{},
		&models.Tag{})
}

func (suite *HostsServiceTestSuite) SetupTest() {
	suite.tx = suite.db.Begin()
	suite.hostsService = NewHostsService(suite.tx)
}

func (suite *HostsServiceTestSuite) TearDownTest() {
	suite.tx.Rollback()
}

func (suite *HostsServiceTestSuite) TestHostsService_GetAll() {
	timeSince = func(_ time.Time) time.Duration {
		return time.Duration(0)
	}

	hosts, err := suite.hostsService.GetAll(nil, nil)
	suite.NoError(err)

	suite.ElementsMatch(models.HostList{
		{
			ID:            "1",
			Name:          "host1",
			Health:        "passing",
			IPAddresses:   []string{"10.74.1.5"},
			CloudProvider: "azure",
			ClusterID:     "cluster_id_1",
			ClusterName:   "cluster_1",
			ClusterType:   models.ClusterTypeHANAScaleOut,
			AgentVersion:  "rolling1337",
			SAPSystems: []*models.SAPSystem{
				{
					ID:  "sap_system_id_1",
					SID: "DEV",
					Instances: []*models.SAPSystemInstance{
						{
							InstanceNumber: "00",
							SID:            "DEV",
							ClusterName:    "cluster_1",
							ClusterID:      "cluster_id_1",
							HostID:         "1",
							Hostname:       "host1",
						},
					},
				},
			},
			Tags: []string{"tag1"},
		},
		{
			ID:            "2",
			Name:          "host2",
			Health:        "passing",
			IPAddresses:   []string{"10.74.1.10"},
			CloudProvider: "azure",
			ClusterID:     "cluster_id_2",
			ClusterName:   "cluster_2",
			ClusterType:   models.ClusterTypeUnknown,
			AgentVersion:  "stable",
			SAPSystems: []*models.SAPSystem{
				{
					ID:  "sap_system_id_2",
					SID: "QAS",
					Instances: []*models.SAPSystemInstance{
						{
							InstanceNumber: "10",
							SID:            "QAS",
							ClusterName:    "cluster_2",
							ClusterID:      "cluster_id_2",
							HostID:         "2",
							Hostname:       "host2",
						},
					},
				},
			},
			Tags: []string{"tag2"},
		},
	}, hosts)
}

func (suite *HostsServiceTestSuite) TestHostsService_GetAll_Filters() {
	timeSince = func(_ time.Time) time.Duration {
		return time.Duration(0)
	}

	hosts, _ := suite.hostsService.GetAll(&HostsFilter{
		Tags:   []string{"tag1"},
		SIDs:   []string{"DEV"},
		Health: []string{"passing", "unknown"},
	}, nil)
	suite.Equal(1, len(hosts))
	suite.Equal("1", hosts[0].ID)
}

func (suite *HostsServiceTestSuite) TestHostsService_GetByID() {
	host, _ := suite.hostsService.GetByID("1")
	suite.Equal("host1", host.Name)
}

func (suite *HostsServiceTestSuite) TestHostsService_GetByID_NotFound() {
	host, err := suite.hostsService.GetByID("13")
	suite.NoError(err)
	suite.Nil(host)
}

func (suite *HostsServiceTestSuite) TestHostsService_GetAllBySAPSystemID() {
	hosts, _ := suite.hostsService.GetAllBySAPSystemID("sap_system_id_2")
	suite.Equal(1, len(hosts))
	suite.Equal("2", hosts[0].ID)
}

func (suite *HostsServiceTestSuite) TestHostsService_GetHostsCount() {
	count, err := suite.hostsService.GetCount()

	suite.NoError(err)
	suite.Equal(2, count)
}

func (suite *HostsServiceTestSuite) TestHostsService_GetAllTags() {
	hosts, _ := suite.hostsService.GetAllTags()
	suite.EqualValues([]string{"tag1", "tag2"}, hosts)
}

func (suite *HostsServiceTestSuite) TestHostsService_GetAllSIDs() {
	hosts, _ := suite.hostsService.GetAllSIDs()
	suite.ElementsMatch([]string{"DEV", "QAS"}, hosts)
}

func (suite *HostsServiceTestSuite) TestHostsService_Heartbeat() {
	err := suite.hostsService.Heartbeat("1")
	suite.NoError(err)

	var heartbeat entities.HostHeartbeat
	suite.tx.First(&heartbeat)
	suite.Equal("1", heartbeat.AgentID)
}

func (suite *HostsServiceTestSuite) TestHostsService_computeHealth() {
	host := hostsFixtures()[0]

	timeSince = func(_ time.Time) time.Duration {
		return time.Duration(0)
	}
	suite.Equal(models.HostHealthPassing, computeHealth(&host))

	timeSince = func(_ time.Time) time.Duration {
		return time.Duration(HeartbeatTreshold + 1)
	}
	suite.Equal(models.HostHealthCritical, computeHealth(&host))

	host.Heartbeat = nil
	suite.Equal(models.HostHealthUnknown, computeHealth(&host))
}
