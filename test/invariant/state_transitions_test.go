package invariant_test

import (
	"fmt"

	cosmossdk_io_math "cosmossdk.io/math"
	testcommon "github.com/allora-network/allora-chain/test/common"
	"github.com/stretchr/testify/require"
)

// Every function responsible for doing a state transition
// should adhere to this function signature
type StateTransitionFunc func(
	m *testcommon.TestConfig,
	actor1 Actor,
	actor2 Actor,
	amount *cosmossdk_io_math.Int,
	topicId uint64,
	data *SimulationData,
	iteration int,
)

// keep track of the name of the state transition as well as the function
type StateTransition struct {
	name string
	f    StateTransitionFunc
}

// The list of possible state transitions we can take are:
//
// create a new topic,
// fund a topic some more,
// register as a reputer,
// register as a worker,
// unregister as a reputer,
// unregister as a worker,
// stake as a reputer,
// stake in a reputer (delegate),
// unstake as a reputer,
// unstake from a reputer (undelegate),
// cancel the removal of stake (as a reputer),
// cancel the removal of delegated stake (delegator),
// collect delegator rewards,
// produce an inference (insert a bulk worker payload),
// produce reputation scores (insert a bulk reputer payload)
func allTransitions() []StateTransition {
	return []StateTransition{
		{"createTopic", createTopic},
		{"fundTopic", fundTopic},
		{"registerWorker", registerWorker},
		{"registerReputer", registerReputer},
		{"unregisterWorker", unregisterWorker},
		{"unregisterReputer", unregisterReputer},
		{"stakeAsReputer", stakeAsReputer},
		{"delegateStake", delegateStake},
		{"unstakeAsReputer", unstakeAsReputer},
		{"undelegateStake", undelegateStake},
		{"cancelStakeRemoval", cancelStakeRemoval},
		{"cancelDelegateStakeRemoval", cancelDelegateStakeRemoval},
		{"collectDelegatorRewards", collectDelegatorRewards},
		{"produceInferenceAndReputation", produceInferenceAndReputation},
	}
}

// state machine dependencies for valid transitions
//
// fundTopic: CreateTopic
// RegisterWorkerForTopic: CreateTopic
// RegisterReputerForTopic: CreateTopic
// unRegisterReputer: RegisterReputerForTopic
// unRegisterWorker: RegisterWorkerForTopic
// stakeReputer: RegisterReputerForTopic, CreateTopic
// delegateStake: CreateTopic, RegisterReputerForTopic
// unstakeReputer: stakeReputer
// unstakeDelegator: delegateStake
// cancelStakeRemoval: unstakeReputer
// cancelDelegateStakeRemoval: unstakeDelegator
// collectDelegatorRewards: delegateStake, fundTopic, InsertBulkWorkerPayload, InsertBulkReputerPayload
// InsertBulkWorkerPayload: RegisterWorkerForTopic, FundTopic
// InsertBulkReputerPayload: RegisterReputerForTopic, InsertBulkWorkerPayload
func canTransitionOccur(m *testcommon.TestConfig, data *SimulationData, transition StateTransition) bool {
	switch transition.name {
	case "unregisterWorker":
		return anyWorkersRegistered(data)
	case "unregisterReputer":
		return anyReputersRegistered(data)
	case "stakeAsReputer":
		return anyReputersRegistered(data)
	case "delegateStake":
		return anyReputersRegistered(data)
	case "unstakeAsReputer":
		return anyReputersStaked(data)
	case "undelegateStake":
		return anyDelegatorsStaked(data)
	case "collectDelegatorRewards":
		return anyDelegatorsStaked(data) && anyReputersRegistered(data)
	case "produceInferenceAndReputation":
		return findIfActiveTopics(m, data)

	// NOT YET IMPLEMENTED
	case "cancelStakeRemoval":
		return false
	case "cancelDelegateStakeRemoval":
		return false
	default:
		return true
	}
}

// is this specific combination of actors, amount, and topicId valid for the transition?
func isValidTransition(m *testcommon.TestConfig, transition StateTransition, actor1 Actor, actor2 Actor, amount *cosmossdk_io_math.Int, topicId uint64, data *SimulationData, iteration int) bool {
	switch transition.name {
	case "collectDelegatorRewards":
		// if the reputer unregisters before the delegator withdraws stake, it can be invalid for a
		// validator to collecte rewards
		if !data.isReputerRegistered(topicId, actor2) {
			iterationLog(m.T, iteration, "Transition not valid: ", transition.name, actor1, actor2, amount, topicId)
			return false
		}
		return true
	default:
		return true
	}
}

// pickStateTransition picks a random state transition to take and returns which one it picked.
func pickStateTransition(
	m *testcommon.TestConfig,
	iteration int,
	data *SimulationData,
) StateTransition {
	transitions := allTransitions()
	for {
		randIndex := m.Client.Rand.Intn(len(transitions))
		selectedTransition := transitions[randIndex]
		if canTransitionOccur(m, data, selectedTransition) {
			return selectedTransition
		} else {
			iterationLog(m.T, iteration, "Transition not possible: ", selectedTransition.name)
		}
	}
}

// pickRandomActor picks a random actor from the list of actors in the simulation data
func pickRandomActor(m *testcommon.TestConfig, data *SimulationData) Actor {
	return data.actors[m.Client.Rand.Intn(len(data.actors))]
}

// pickRandomActorExcept picks a random actor from the list of actors in the simulation data
// and panics if it can't find one after 5 tries that is not the same as the given actor
func pickRandomActorExcept(m *testcommon.TestConfig, data *SimulationData, actor Actor) Actor {
	count := 0
	for ; count < 5; count++ {
		randomActor := pickRandomActor(m, data)
		if randomActor != actor {
			return randomActor
		}
	}
	panic(
		fmt.Sprintf(
			"could not find a random actor that is not the same as the given actor after %d tries",
			count,
		),
	)
}

// pickActorAndTopicIdForStateTransition picks random actors
// able to take the state transition and returns which one it picked.
// if the transition requires only one actor (the majority) the second is empty
func pickActorAndTopicIdForStateTransition(
	m *testcommon.TestConfig,
	transition StateTransition,
	data *SimulationData,
) (actor1 Actor, actor2 Actor, amount *cosmossdk_io_math.Int, topicId uint64) {
	switch transition.name {
	case "unregisterWorker":
		worker, topicId := data.pickRandomRegisteredWorker()
		return worker, Actor{}, nil, topicId
	case "unregisterReputer":
		reputer, topicId := data.pickRandomRegisteredReputer()
		return reputer, Actor{}, nil, topicId
	case "stakeAsReputer":
		reputer, topicId := data.pickRandomRegisteredReputer()
		amount, err := pickRandomBalanceLessThanHalf(m, reputer)
		require.NoError(m.T, err)
		return reputer, Actor{}, &amount, topicId
	case "delegateStake":
		reputer, topicId := data.pickRandomRegisteredReputer()
		delegator := pickRandomActorExcept(m, data, reputer)
		amount, err := pickRandomBalanceLessThanHalf(m, delegator)
		require.NoError(m.T, err)
		return delegator, reputer, &amount, topicId
	case "unstakeAsReputer":
		reputer, topicId := data.pickRandomStakedReputer()
		amount := data.pickPercentOfStakeByReputer(m.Client.Rand, topicId, reputer)
		return reputer, Actor{}, &amount, topicId
	case "undelegateStake":
		delegator, reputer, topicId := data.pickRandomStakedDelegator()
		amount := data.pickPercentOfStakeByDelegator(m.Client.Rand, topicId, delegator, reputer)
		return delegator, reputer, &amount, topicId
	case "collectDelegatorRewards":
		delegator, reputer, topicId := data.pickRandomStakedDelegator()
		return delegator, reputer, nil, topicId
	case "produceInferenceAndReputation":
		topicId := getActiveTopicId(m)
		worker, err := data.pickRandomWorkerRegisteredInTopic(m.Client.Rand, topicId)
		require.NoError(m.T, err)
		reputer, err := data.pickRandomReputerStakedInTopic(m.Client.Rand, topicId)
		require.NoError(m.T, err)
		return worker, reputer, nil, topicId
	default:
		randomTopicId, err := pickRandomTopicId(m)
		require.NoError(m.T, err)
		randomActor1 := pickRandomActor(m, data)
		randomActor2 := pickRandomActor(m, data)
		amount, err := pickRandomBalanceLessThanHalf(m, randomActor1)
		require.NoError(m.T, err)
		return randomActor1, randomActor2, &amount, randomTopicId
	}
}
