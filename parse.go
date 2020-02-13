package main

import (
	"errors"
	"regexp"
	"strconv"
)

// Coder should use following naming convention for Benchmark functions
// Naming convention: Benchmark[Function_name]_[Function_argument](b *testing.B)
var defaultFunctionSignaturePattern = regexp.MustCompile(
	`Benchmark(?P<functionName>[a-zA-Z0-9/]+)_(?P<functionArguments>[_a-zA-Z0-9]+)-(?P<numberOfThreads>[0-9]+)$`,
)

type (
	// Storage for Func(Arg)=Result relations
	BenchArgSet             map[string]float64
	BenchNameSet            map[string]BenchArgSet
	parsedFunctionSignature struct {
		name            string
		arg             string
		numberOfThreads int
	}
)

// parseFunctionSignature parses function name, argument and number of threads from benchmark output.
func parseFunctionSignature(expression *regexp.Regexp, line string) (*parsedFunctionSignature, error) {
	match := expression.FindStringSubmatch(line)

	// we expect 4 columns
	if len(match) != 4 {
		return nil, errors.New("Can't parse benchmark result")
	}

	expressionCaptureGroups := make(map[string]string)

	for i, name := range expression.SubexpNames() {
		expressionCaptureGroups[name] = match[i]
	}

	_, ok := expressionCaptureGroups["functionName"]
	if !ok {
		return nil, errors.New("No `functionName` capture group in provided expression")
	}

	_, ok = expressionCaptureGroups["functionArguments"]
	if !ok {
		return nil, errors.New("No `functionArguments` capture group in provided expression")
	}

	_, ok = expressionCaptureGroups["numberOfThreads"]
	if !ok {
		return nil, errors.New("No `numberOfThreads` capture group in provided expression")
	}

	numberOfThreads, err := strconv.Atoi(expressionCaptureGroups["numberOfThreads"])
	if err != nil {
		return nil,
			errors.New(
				"Can't parse `numberOfThreads` string as integer with provided expression result",
			)
	}

	return &parsedFunctionSignature{
		name:            expressionCaptureGroups["functionName"],
		arg:             expressionCaptureGroups["functionArguments"],
		numberOfThreads: numberOfThreads,
	}, nil
}
