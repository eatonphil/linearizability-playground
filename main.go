package main

import "os"
import "log"
import "maps"
import "github.com/anishathalye/porcupine"

func visualizeTempFile(model porcupine.Model, info porcupine.LinearizationInfo) {
	file, err := os.CreateTemp("", "*.html")
	if err != nil {
		panic("failed to create temp file")
	}
	err = porcupine.Visualize(model, info, file)
	if err != nil {
		panic("visualization failed")
	}
	log.Printf("wrote visualization to %s", file.Name())
}

type kvInput struct {
	op string
	key string
	value string
}

func main() {
	kvModel := porcupine.Model{
		Init: func() interface{} {
			return map[string]string{}
		},
		// step function: takes a state, input, and output, and returns whether it
		// was a legal operation, along with a new state
		Step: func(stateInt, inputInt, outputInt interface{}) (bool, interface{}) {
			input := inputInt.(kvInput)
			output := outputInt.(map[string]string)
			state := stateInt.(map[string]string)
			if input.op == "set" {
				newState := maps.Clone(state)
				newState[input.key] = input.value
				return true, newState // always ok to execute a put
			} else if input.op == "get" {
				readCorrectValue := state[input.key] == output[input.key]
				return readCorrectValue, state // state is unchanged
			}

			panic("Unexpected operation")
		},
		Equal: func (aInt, bInt interface{}) bool {
			a := aInt.(map[string]string)
			b := bInt.(map[string]string)
			return maps.Equal(a, b)
		},
	}
	
	// Operation is {clientId, input, start timestamp, output, end timestamp}
	ops := []porcupine.Operation{
		// Write 1 from client 1 starts at T0, ends at T2, writes 100
		{0, kvInput{"set", "a", "100"}, 0, map[string]string{}, 2},
		// Write 2 from client 2 starts at T1, ends at T3, writes 200
		{1, kvInput{"set", "a", "200"}, 1, map[string]string{}, 3},
		// Read 1 from client 1 starts at T4, ends at T6, sees 100
		{0, kvInput{"get", "a", "0"}, 4, map[string]string{"a": "100"}, 6},
		// Read 2 from client 2 starts at T5, ends at T7, sees 100
		{1, kvInput{"get", "a", "0"}, 5, map[string]string{"a": "100"}, 7},
	}
	res, info := porcupine.CheckOperationsVerbose(kvModel, ops, 0)
	visualizeTempFile(kvModel, info)

	if res != porcupine.Ok {
		panic("expected operations to be linearizable")
	}
}
