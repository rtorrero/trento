package ctl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/trento/test/helpers"
	"github.com/trento-project/trento/web/datapipeline"
	"github.com/trento-project/trento/web/entities"
	"gorm.io/gorm"
)

type CtlTestSuite struct {
	suite.Suite
	db *gorm.DB
	tx *gorm.DB
}

func TestCtlTestSuite(t *testing.T) {
	suite.Run(t, new(CtlTestSuite))
}

func (suite *CtlTestSuite) SetupSuite() {
	suite.db = helpers.SetupTestDatabase(suite.T())
}

func (suite *CtlTestSuite) SetupTest() {
	suite.tx = suite.db.Begin()
}

func (suite *CtlTestSuite) TearDownTest() {
	suite.tx.Rollback()
}

func (suite *CtlTestSuite) TestPruneEvents() {
	suite.tx.AutoMigrate(&datapipeline.DataCollectedEvent{})

	events := []datapipeline.DataCollectedEvent{
		{
			ID:            1,
			AgentID:       "agent_id",
			DiscoveryType: "test_discovery_type",
			Payload:       []byte("{}"),
			CreatedAt:     time.Now().Add(-24 * 15 * time.Hour),
		},
		{
			ID:            2,
			AgentID:       "agent_id",
			DiscoveryType: "test_discovery_type",
			Payload:       []byte("{}"),
			CreatedAt:     time.Now().Add(-24 * 10 * time.Hour),
		},
		{
			ID:            3,
			AgentID:       "agent_id",
			DiscoveryType: "test_discovery_type",
			Payload:       []byte("{}"),
			CreatedAt:     time.Now().Add(-24 * 6 * time.Hour),
		},
	}
	suite.tx.Create(events)

	pruneEvents(suite.tx, 24*10*time.Hour)

	var prunedEvents []datapipeline.DataCollectedEvent
	suite.tx.Find(&prunedEvents)

	suite.Equal(1, len(prunedEvents))
	suite.Equal(int64(3), prunedEvents[0].ID)
}

func (suite *CtlTestSuite) TestPruneChecksResults() {
	suite.tx.AutoMigrate(&entities.ChecksResult{})

	checksResults := []entities.ChecksResult{
		{
			ID:        1,
			GroupID:   "group_id",
			Payload:   []byte("{}"),
			CreatedAt: time.Now().Add(-24 * 15 * time.Hour),
		},
		{
			ID:        2,
			GroupID:   "group_id",
			Payload:   []byte("{}"),
			CreatedAt: time.Now().Add(-24 * 10 * time.Hour),
		},
		{
			ID:        3,
			GroupID:   "group_id",
			Payload:   []byte("{}"),
			CreatedAt: time.Now().Add(-24 * 6 * time.Hour),
		},
	}
	suite.tx.Create(checksResults)

	pruneChecksResults(suite.tx, 24*10*time.Hour)

	var prunedChecksResults []entities.ChecksResult
	suite.tx.Find(&prunedChecksResults)

	suite.Equal(1, len(prunedChecksResults))
	suite.Equal(int64(3), prunedChecksResults[0].ID)
}

func (suite *CtlTestSuite) TestGetLatestEvents() {
	suite.tx.AutoMigrate(&datapipeline.DataCollectedEvent{})

	events := []datapipeline.DataCollectedEvent{
		{
			ID:            1,
			AgentID:       "agent_id_1",
			DiscoveryType: "discovery_type_1",
			Payload:       []byte("{}"),
		},
		{
			ID:            2,
			AgentID:       "agent_id_1",
			DiscoveryType: "discovery_type_1",
			Payload:       []byte("{}"),
		},

		{
			ID:            3,
			AgentID:       "agent_id_2",
			DiscoveryType: "discovery_type_2",
			Payload:       []byte("{}"),
		},
		{
			ID:            4,
			AgentID:       "agent_id_2",
			DiscoveryType: "discovery_type_2",
			Payload:       []byte("{}"),
		},
		{
			ID:            5,
			AgentID:       "agent_id_2",
			DiscoveryType: "discovery_type_2",
			Payload:       []byte("{}"),
		},
		{
			ID:            6,
			AgentID:       "agent_id_1",
			DiscoveryType: "discovery_type_3",
			Payload:       []byte("{}"),
		},
		{
			ID:            7,
			AgentID:       "agent_id_2",
			DiscoveryType: "discovery_type_3",
			Payload:       []byte("{}"),
		},
	}

	err := suite.tx.Create(&events).Error
	suite.NoError(err)

	latestEvents, err := getLatestEvents(suite.tx)
	suite.NoError(err)

	suite.Equal(4, len(latestEvents))
	suite.Equal(int64(2), latestEvents[0].ID)
	suite.Equal(int64(5), latestEvents[1].ID)
	suite.Equal(int64(6), latestEvents[2].ID)
	suite.Equal(int64(7), latestEvents[3].ID)
}
