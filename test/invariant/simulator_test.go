package invariant_test

import (
	"context"

	cosmossdk_io_math "cosmossdk.io/math"
	testcommon "github.com/allora-network/allora-chain/test/common"
)

// set up the common state for the simulator
// prior to either doing random simulation
// or manual simulation
func simulateSetUp(
	m *testcommon.TestConfig,
	numActors int,
	epochLength int,
) (
	faucet Actor,
	simulationData *SimulationData,
) {
	// fund all actors from the faucet with some amount
	// give everybody the same amount of money to start with
	actorsList := createActors(m, numActors)
	faucet = Actor{
		name: getFaucetName(m.Seed),
		addr: m.FaucetAddr,
		acc:  m.FaucetAcc,
	}
	preFundAmount, err := getPreFundAmount(m, faucet, numActors)
	if err != nil {
		m.T.Fatal(err)
	}
	err = fundActors(
		m,
		faucet,
		actorsList,
		preFundAmount,
	)
	if err != nil {
		m.T.Fatal(err)
	}
	data := SimulationData{
		epochLength:        int64(epochLength),
		actors:             actorsList,
		counts:             StateTransitionCounts{},
		registeredWorkers:  testcommon.NewRandomKeyMap[Registration, struct{}](m.Client.Rand),
		registeredReputers: testcommon.NewRandomKeyMap[Registration, struct{}](m.Client.Rand),
		reputerStakes: testcommon.NewRandomKeyMap[Registration, cosmossdk_io_math.Int](
			m.Client.Rand,
		),
		delegatorStakes: testcommon.NewRandomKeyMap[Delegation, cosmossdk_io_math.Int](
			m.Client.Rand,
		),
	}
	return faucet, &data
}

// run the outer loop of the simulator
func simulate(
	m *testcommon.TestConfig,
	maxIterations int,
	numActors int,
	epochLength int,
) {
	faucet, simulationData := simulateSetUp(m, numActors, epochLength)
	if MANUAL_SIMULATION {
		simulateManual(m, faucet, simulationData)
	} else {
		simulateAutomatic(m, faucet, simulationData, maxIterations)
	}
}

// this is a helper function
// if you need to directly simulate some activity in a test
// say, to reproduce an issue
// you can do so here and it will occur in the same way as the simulator
const MANUAL_SIMULATION = true

// this is the body of the "manual" simulation mode
// put your code here if you wish to manually send transactions
// in some specific order to test something
func simulateManual(
	m *testcommon.TestConfig,
	faucet Actor,
	simulationData *SimulationData,
) {
	iterationLog(m.T, 0, "manual simulation mode")
	reputer := pickRandomActor(m, simulationData)
	delegator := pickRandomActorExcept(m, simulationData, []Actor{reputer})
	worker := pickRandomActorExcept(m, simulationData, []Actor{reputer, delegator})
	reputerWithoutStake := pickRandomActorExcept(m, simulationData, []Actor{reputer, delegator, worker})
	amount := cosmossdk_io_math.NewInt(1e10)

	// create topic
	createTopic(m, faucet, Actor{}, nil, 0, simulationData, 0)
	// register reputer
	registerReputer(m, reputer, Actor{}, nil, 1, simulationData, 1)
	// delegate from delegator on reputer
	delegateStake(m, delegator, reputer, &amount, 1, simulationData, 2)
	// register reputer without stake
	registerReputer(m, reputerWithoutStake, Actor{}, nil, 1, simulationData, 3)
	// unregister reputer
	unregisterReputer(m, reputer, Actor{}, nil, 1, simulationData, 4)
	// fund the topic from delegator
	fundTopic(m, delegator, Actor{}, &amount, 1, simulationData, 5)
	// register worker
	registerWorker(m, worker, Actor{}, nil, 1, simulationData, 6)
	// now nobody has stake, is the topic active?
	// make sure an ABCI endblock has passed
	ctx := context.Background()
	m.Client.WaitForNextBlock(ctx)
	isActive := findIfActiveTopics(m, simulationData)
	m.T.Log("Is topic active?", isActive)
}

// this is the body of the "normal" simulation mode
func simulateAutomatic(
	m *testcommon.TestConfig,
	faucet Actor,
	simulationData *SimulationData,
	maxIterations int,
) {
	// iteration 0, always create a topic to start
	createTopic(m, faucet, Actor{}, nil, 0, simulationData, 0)

	// every iteration, pick an actor,
	// then pick a state transition function for that actor to do
	for i := 1; i < maxIterations; i++ {
		for {
			stateTransition := pickStateTransition(m, i, simulationData)
			actor1, actor2, amount, topicId := pickActorAndTopicIdForStateTransition(
				m,
				stateTransition,
				simulationData,
			)
			if isValidTransition(m, stateTransition, actor1, actor2, amount, topicId, simulationData, i) {
				stateTransition.f(m, actor1, actor2, amount, topicId, simulationData, i)
				break
			}
		}
		if i%5 == 0 {
			m.T.Log("State Transitions Summary:", simulationData.counts)
		}
	}
	m.T.Log("Final Summary:", simulationData.counts)
}
