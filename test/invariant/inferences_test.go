package invariant_test

import (
	cosmossdk_io_math "cosmossdk.io/math"
	testcommon "github.com/allora-network/allora-chain/test/common"
)

func produceInferenceAndReputation(
	m *testcommon.TestConfig,
	actor1 Actor,
	actor2 Actor,
	amount *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(m.T, iteration, "producing inference and reputation for topic id", topicId)
	data.counts.incrementProduceInferenceAndReputationCount()
	iterationLog(m.T, iteration, "produced inference for topic id", topicId)
}
