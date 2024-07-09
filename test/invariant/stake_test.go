package invariant_test

import (
	"context"

	cosmossdk_io_math "cosmossdk.io/math"

	testcommon "github.com/allora-network/allora-chain/test/common"
	emissionstypes "github.com/allora-network/allora-chain/x/emissions/types"
	"github.com/stretchr/testify/require"
)

// stake actor as a reputer, pick a random amount to stake that is less than half their current balance
func stakeAsReputer(
	m *testcommon.TestConfig,
	actor Actor,
	_ Actor,
	amount *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(
		m.T,
		iteration,
		"staking as a reputer",
		actor,
		"in topic id",
		topicId,
		" in amount",
		amount.String(),
	)
	msg := emissionstypes.MsgAddStake{
		Sender:  actor.addr,
		TopicId: topicId,
		Amount:  *amount,
	}
	ctx := context.Background()
	txResp, err := m.Client.BroadcastTx(ctx, actor.acc, &msg)
	require.NoError(m.T, err)

	_, err = m.Client.WaitForTx(ctx, txResp.TxHash)
	require.NoError(m.T, err)

	response := &emissionstypes.MsgAddStakeResponse{}
	err = txResp.Decode(response)
	require.NoError(m.T, err)

	data.addReputerStake(topicId, actor, *amount)
	data.counts.incrementStakeAsReputerCount()
	iterationLog(
		m.T,
		iteration,
		"staked ",
		actor,
		"as a reputer in topic id ",
		topicId,
		" in amount ",
		amount.String(),
	)
}

// tell if any reputers are currently staked
func anyReputersStaked(data *SimulationData) bool {
	return data.reputerStakes.Len() > 0
}

// tell if any delegators are currently staked
func anyDelegatorsStaked(data *SimulationData) bool {
	return data.delegatorStakes.Len() > 0
}

// mark stake for removal as a reputer
// the amount will either be 1/10, 1/3, 1/2, 6/7, or the full amount of their
// current stake to be removed
func unstakeAsReputer(
	m *testcommon.TestConfig,
	actor Actor,
	_ Actor,
	amount *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(
		m.T,
		iteration,
		"unstaking as a reputer",
		actor,
		"in topic id",
		topicId,
		" in amount",
		amount.String(),
	)
	msg := emissionstypes.MsgRemoveStake{
		Sender:  actor.addr,
		TopicId: topicId,
		Amount:  *amount,
	}
	ctx := context.Background()
	txResp, err := m.Client.BroadcastTx(ctx, actor.acc, &msg)
	require.NoError(m.T, err)

	_, err = m.Client.WaitForTx(ctx, txResp.TxHash)
	require.NoError(m.T, err)

	response := &emissionstypes.MsgRemoveStakeResponse{}
	err = txResp.Decode(response)
	require.NoError(m.T, err)

	data.markStakeRemovalReputerStake(topicId, actor, amount)
	data.counts.incrementUnstakeAsReputerCount()
	iterationLog(
		m.T,
		iteration,
		"unstaked from ",
		actor,
		"as a reputer in topic id ",
		topicId,
		" in amount ",
		amount.String(),
	)
}

func cancelStakeRemoval(
	m *testcommon.TestConfig,
	actor Actor,
	_ Actor,
	_ *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(
		m.T,
		iteration,
		"cancelling stake removal as a reputer",
		actor,
		"in topic id",
		topicId,
	)
	data.counts.incrementCancelStakeRemovalCount()
}

// stake as a delegator upon a reputer
// NOTE: in this case, the param actor is the reputer being staked upon,
// rather than the actor doing the staking.
func delegateStake(
	m *testcommon.TestConfig,
	delegator Actor,
	reputer Actor,
	amount *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(
		m.T,
		iteration,
		"delegating stake",
		delegator,
		"upon reputer",
		reputer,
		"in topic id",
		topicId,
		" in amount",
		amount.String(),
	)
	msg := emissionstypes.MsgDelegateStake{
		Sender:  delegator.addr,
		Reputer: reputer.addr,
		TopicId: topicId,
		Amount:  *amount,
	}
	ctx := context.Background()
	txResp, err := m.Client.BroadcastTx(ctx, delegator.acc, &msg)
	require.NoError(m.T, err)

	_, err = m.Client.WaitForTx(ctx, txResp.TxHash)
	require.NoError(m.T, err)

	registerWorkerResponse := &emissionstypes.MsgDelegateStakeResponse{}
	err = txResp.Decode(registerWorkerResponse)
	require.NoError(m.T, err)

	data.addDelegatorStake(topicId, delegator, reputer, *amount)
	data.counts.incrementDelegateStakeCount()
	iterationLog(
		m.T,
		iteration,
		"delegating stake",
		delegator,
		"upon reputer",
		reputer,
		"in topic id",
		topicId,
		" in amount",
		amount.String(),
	)
}

// undelegate a percentage of the stake that the delegator has upon the reputer, either 1/10, 1/3, 1/2, 6/7, or the full amount
func undelegateStake(
	m *testcommon.TestConfig,
	delegator Actor,
	reputer Actor,
	amount *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(
		m.T,
		iteration,
		"delegator ",
		delegator,
		" unstaking from reputer ",
		reputer,
		" in topic id ",
		topicId,
		" in amount ",
		amount.String(),
	)
	msg := emissionstypes.MsgRemoveDelegateStake{
		Sender:  delegator.addr,
		Reputer: reputer.addr,
		TopicId: topicId,
		Amount:  *amount,
	}
	ctx := context.Background()
	txResp, err := m.Client.BroadcastTx(ctx, delegator.acc, &msg)
	require.NoError(m.T, err)

	_, err = m.Client.WaitForTx(ctx, txResp.TxHash)
	require.NoError(m.T, err)

	response := &emissionstypes.MsgRemoveDelegateStakeResponse{}
	err = txResp.Decode(response)
	require.NoError(m.T, err)

	data.markStakeRemovalDelegatorStake(topicId, delegator, reputer, amount)
	data.counts.incrementUndelegateStakeCount()
	iterationLog(
		m.T,
		iteration,
		"delegator ",
		delegator,
		" unstaked from reputer ",
		reputer,
		" in topic id ",
		topicId,
		" in amount ",
		amount.String(),
	)
}

func cancelDelegateStakeRemoval(
	m *testcommon.TestConfig,
	delegator Actor,
	reputer Actor,
	_ *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(
		m.T,
		iteration,
		"cancelling stake removal as a delegator",
		delegator,
		"in topic id",
		topicId,
	)
	data.counts.incrementCancelDelegateStakeRemovalCount()
}

func collectDelegatorRewards(
	m *testcommon.TestConfig,
	delegator Actor,
	reputer Actor,
	_ *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
) {
	iterationLog(
		m.T,
		iteration,
		"delegator ",
		delegator,
		" collecting rewards for delegating on ",
		reputer,
		" in topic id ",
		topicId,
	)
	msg := emissionstypes.MsgRewardDelegateStake{
		Sender:  delegator.addr,
		TopicId: topicId,
		Reputer: reputer.addr,
	}
	ctx := context.Background()
	txResp, err := m.Client.BroadcastTx(ctx, delegator.acc, &msg)
	require.NoError(m.T, err)

	_, err = m.Client.WaitForTx(ctx, txResp.TxHash)
	require.NoError(m.T, err)

	response := &emissionstypes.MsgRewardDelegateStakeResponse{}
	err = txResp.Decode(response)
	require.NoError(m.T, err)

	data.counts.incrementCollectDelegatorRewardsCount()
	iterationLog(
		m.T,
		iteration,
		"delegator ",
		delegator,
		" collected rewards in topic id ",
		topicId,
	)
}
