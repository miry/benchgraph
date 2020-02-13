package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/fatih/color"
	"golang.org/x/tools/benchmark/parse"
)

// uploadData sends data to server and expects graph url.
func uploadData(apiUrl, data, title string) (string, error) {

	resp, err := http.PostForm(apiUrl, url.Values{"data": {data}, "title": {title}})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New("Server din't return graph URL")
	}

	return string(body), nil
}

func main() {
	var oBenchNames, oBenchArgs stringList

	// graph elements will be ordered as in benchmark output by default - unless the order was specified here
	flag.Var(&oBenchNames, "obn", "comma-separated list of benchmark names")
	flag.Var(&oBenchArgs, "oba", "comma-separated list of benchmark arguments")
	title := flag.String("title", "Graph: Benchmark results in ns/op", "title of a graph")
	apiUrl := flag.String("apiurl", "http://benchgraph.codingberg.com", "url to server api")
	functionSignaturePattern := flag.String(
		"function-signature-pattern",
		defaultFunctionSignaturePattern.String(),
		fmt.Sprintf(
			"regex expression to extract function test signature, default: %s",
			defaultFunctionSignaturePattern,
		),
	)
	flag.Parse()

	var skipBenchNamesParsing, skipBenchArgsParsing bool

	if oBenchNames.Len() > 0 {
		skipBenchNamesParsing = true
	}
	if oBenchArgs.Len() > 0 {
		skipBenchArgsParsing = true
	}

	if len(*functionSignaturePattern) < 1 {
		fmt.Fprintf(os.Stderr, "empty `test-name-expression` provided")
		os.Exit(1)
	}

	functionSignaturePatternRegex, err := regexp.Compile(*functionSignaturePattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid regex `test-name-expression` provided. Err: %s", err)
		os.Exit(1)
	}

	benchResults := make(BenchNameSet)

	// parse Golang benchmark results, line by line
	scan := bufio.NewScanner(os.Stdin)
	green := color.New(color.FgGreen).SprintfFunc()
	red := color.New(color.FgRed).SprintFunc()
	for scan.Scan() {
		line := scan.Text()

		mark := green("√")

		b, err := parse.ParseLine(line)
		if err != nil {
			mark = red("?")
		}

		// read bench name and arguments
		if b != nil {
			parsedFunctionSignature, err := parseFunctionSignature(functionSignaturePatternRegex, b.Name)
			if err != nil {
				mark = red("!")
				fmt.Printf("%s %s\n", mark, line)
				continue
			}

			if !skipBenchNamesParsing && !oBenchNames.stringInList(parsedFunctionSignature.name) {
				oBenchNames.Add(parsedFunctionSignature.name)
			}

			if !skipBenchArgsParsing && !oBenchArgs.stringInList(parsedFunctionSignature.arg) {
				oBenchArgs.Add(parsedFunctionSignature.arg)
			}

			if _, ok := benchResults[parsedFunctionSignature.name]; !ok {
				benchResults[parsedFunctionSignature.name] = make(BenchArgSet)
			}

			benchResults[parsedFunctionSignature.name][parsedFunctionSignature.arg] = b.NsPerOp
		}

		fmt.Printf("%s %s\n", mark, line)
	}

	if err := scan.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading standard input: %v", err)
		os.Exit(1)
	}

	if len(benchResults) == 0 {
		fmt.Fprintf(os.Stderr, "no data to show.\n\n")
		os.Exit(1)
	}

	fmt.Println("\nWaiting for server response ...")

	data := graphData(benchResults, oBenchNames, oBenchArgs)

	graphUrl, err := uploadData(*apiUrl, string(data), *title)
	if err != nil {
		fmt.Fprintf(os.Stderr, "uploading data: %v", err)
		os.Exit(1)
	}

	fmt.Println("=========================================")
	fmt.Println(graphUrl)
	fmt.Println("\n=========================================")
}
