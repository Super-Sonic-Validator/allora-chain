package invariant_test

import (
	"context"

	cosmossdk_io_math "cosmossdk.io/math"
	testcommon "github.com/allora-network/allora-chain/test/common"
	emissionstypes "github.com/allora-network/allora-chain/x/emissions/types"
	"github.com/stretchr/testify/require"
)

// determine if this state transition is worth trying based on our knowledge of the state
func anyWorkersRegistered(data *SimulationData) bool {
	return data.registeredWorkers.Len() > 0
}

// determine if this state transition is worth trying based on our knowledge of the state
func anyReputersRegistered(data *SimulationData) bool {
	return data.registeredReputers.Len() > 0
}

// register actor as a new worker in topicId
func registerWorker(
	m *testcommon.TestConfig,
	actor Actor,
	_ Actor,
	_ *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(m.T, iteration, "registering ", actor, "as worker in topic id", topicId)
	ctx := context.Background()
	txResp, err := m.Client.BroadcastTx(ctx, actor.acc, &emissionstypes.MsgRegister{
		Sender:       actor.addr,
		Owner:        actor.addr, // todo pick random other actor
		LibP2PKey:    getLibP2pKeyName(actor),
		MultiAddress: getMultiAddressName(actor),
		IsReputer:    false,
		TopicId:      topicId,
	})
	require.NoError(m.T, err)

	_, err = m.Client.WaitForTx(ctx, txResp.TxHash)
	require.NoError(m.T, err)

	registerWorkerResponse := &emissionstypes.MsgRegisterResponse{}
	err = txResp.Decode(registerWorkerResponse)
	require.NoError(m.T, err)
	require.True(m.T, registerWorkerResponse.Success)

	data.addWorkerRegistration(topicId, actor)
	data.counts.incrementRegisterWorkerCount()
	iterationLog(m.T, iteration, "registered ", actor, "as worker in topic id ", topicId)
}

// unregister actor from being a worker in topic topicId
func unregisterWorker(
	m *testcommon.TestConfig,
	actor Actor,
	_ Actor,
	_ *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(m.T, iteration, "unregistering ", actor, "as worker in topic id", topicId)
	ctx := context.Background()
	txResp, err := m.Client.BroadcastTx(ctx, actor.acc, &emissionstypes.MsgRemoveRegistration{
		Sender:    actor.addr,
		TopicId:   topicId,
		IsReputer: false,
	})
	require.NoError(m.T, err)

	_, err = m.Client.WaitForTx(ctx, txResp.TxHash)
	require.NoError(m.T, err)

	removeRegistrationResponse := &emissionstypes.MsgRemoveRegistrationResponse{}
	err = txResp.Decode(removeRegistrationResponse)
	require.NoError(m.T, err)
	require.True(m.T, removeRegistrationResponse.Success)

	data.removeWorkerRegistration(topicId, actor)
	data.counts.incrementUnregisterWorkerCount()
	iterationLog(m.T, iteration, "unregistered ", actor, "as worker in topic id ", topicId)
}

// register actor as a new reputer in topicId
func registerReputer(
	m *testcommon.TestConfig,
	actor Actor,
	_ Actor,
	_ *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(m.T, iteration, "registering ", actor, "as reputer in topic id", topicId)
	ctx := context.Background()
	txResp, err := m.Client.BroadcastTx(ctx, actor.acc, &emissionstypes.MsgRegister{
		Sender:       actor.addr,
		Owner:        actor.addr, // todo pick random other actor
		LibP2PKey:    getLibP2pKeyName(actor),
		MultiAddress: getMultiAddressName(actor),
		IsReputer:    true,
		TopicId:      topicId,
	})
	require.NoError(m.T, err)

	_, err = m.Client.WaitForTx(ctx, txResp.TxHash)
	require.NoError(m.T, err)

	registerWorkerResponse := &emissionstypes.MsgRegisterResponse{}
	err = txResp.Decode(registerWorkerResponse)
	require.NoError(m.T, err)
	require.True(m.T, registerWorkerResponse.Success)

	data.addReputerRegistration(topicId, actor)
	data.counts.incrementRegisterReputerCount()
	iterationLog(m.T, iteration, "registered ", actor, "as reputer in topic id ", topicId)
}

// unregister actor as a reputer in topicId
func unregisterReputer(
	m *testcommon.TestConfig,
	actor Actor,
	_ Actor,
	_ *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(m.T, iteration, "unregistering ", actor, "as reputer in topic id", topicId)
	ctx := context.Background()
	txResp, err := m.Client.BroadcastTx(ctx, actor.acc, &emissionstypes.MsgRemoveRegistration{
		Sender:    actor.addr,
		TopicId:   topicId,
		IsReputer: true,
	})
	require.NoError(m.T, err)

	_, err = m.Client.WaitForTx(ctx, txResp.TxHash)
	require.NoError(m.T, err)

	removeRegistrationResponseMsg := &emissionstypes.MsgRemoveRegistrationResponse{}
	err = txResp.Decode(removeRegistrationResponseMsg)
	require.NoError(m.T, err)
	require.True(m.T, removeRegistrationResponseMsg.Success)

	data.removeReputerRegistration(topicId, actor)
	data.counts.incrementUnregisterReputerCount()
	iterationLog(m.T, iteration, "unregistered ", actor, "as reputer in topic id ", topicId)
}
