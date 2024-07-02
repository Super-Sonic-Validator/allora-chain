package invariant_test

import (
	"context"
	"fmt"

	cosmossdk_io_math "cosmossdk.io/math"
	alloraMath "github.com/allora-network/allora-chain/math"
	testcommon "github.com/allora-network/allora-chain/test/common"
	emissionstypes "github.com/allora-network/allora-chain/x/emissions/types"
	"github.com/stretchr/testify/require"
)

// Use actor to create a new topic
func createTopic(
	m *testcommon.TestConfig,
	actor Actor,
	_ Actor,
	_ *cosmossdk_io_math.Int,
	_ uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(m.T, iteration, actor, "creating new topic")
	createTopicRequest := &emissionstypes.MsgCreateNewTopic{
		Creator:         actor.addr,
		Metadata:        fmt.Sprintf("Created topic iteration %d", iteration),
		LossLogic:       "bafybeid7mmrv5qr4w5un6c64a6kt2y4vce2vylsmfvnjt7z2wodngknway",
		LossMethod:      "loss-calculation-eth.wasm",
		InferenceLogic:  "bafybeigx43n7kho3gslauwtsenaxehki6ndjo3s63ahif3yc5pltno3pyq",
		InferenceMethod: "allora-inference-function.wasm",
		EpochLength:     data.epochLength,
		GroundTruthLag:  0,
		DefaultArg:      "ETH",
		PNorm:           alloraMath.NewDecFromInt64(3),
		AlphaRegret:     alloraMath.NewDecFromInt64(1),
		AllowNegative:   true,
		Epsilon:         alloraMath.MustNewDecFromString("0.0001"),
	}

	ctx := context.Background()
	txResp, err := m.Client.BroadcastTx(ctx, actor.acc, createTopicRequest)
	require.NoError(m.T, err)

	_, err = m.Client.WaitForTx(ctx, txResp.TxHash)
	require.NoError(m.T, err)

	createTopicResponse := &emissionstypes.MsgCreateNewTopicResponse{}
	err = txResp.Decode(createTopicResponse)
	require.NoError(m.T, err)

	data.counts.incrementCreateTopicCount()
	iterationLog(m.T, iteration, actor, " created topic ", createTopicResponse.TopicId)
}

// use actor to fund topic, picked randomly
func fundTopic(
	m *testcommon.TestConfig,
	actor Actor,
	_ Actor,
	amount *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(m.T, iteration, actor, "funding topic in amount ", amount)
	fundTopicRequest := &emissionstypes.MsgFundTopic{
		Sender:  actor.addr,
		TopicId: topicId,
		Amount:  *amount,
	}

	ctx := context.Background()
	txResp, err := m.Client.BroadcastTx(ctx, actor.acc, fundTopicRequest)
	require.NoError(m.T, err)

	_, err = m.Client.WaitForTx(ctx, txResp.TxHash)
	require.NoError(m.T, err)

	data.counts.incrementFundTopicCount()
	iterationLog(m.T, iteration, actor, " funded topic ", topicId)
}
