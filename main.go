package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/prometheus/promql/parser"
)

func main() {
	file, err := os.Open("./dashboard_queries.out")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	queries := make([]string, 0, 1000)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		query := scanner.Text()
		query = strings.ReplaceAll(query, `\"`, `"`)
		query = strings.ReplaceAll(query, `\n`, ``)
		query = strings.ReplaceAll(query, `$__interval`, "5m")
		query = strings.ReplaceAll(query, `$interval`, "5m")
		query = strings.ReplaceAll(query, `$resolution`, "5s")

		queries = append(queries, query)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	metrics := map[string]struct{}{}

	for _, query := range queries {
		expr, err := parser.ParseExpr(query)
		if err != nil {
			log.Fatalln(query, err)
		}

		parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
			if n, ok := node.(*parser.VectorSelector); ok {
				metrics[n.Name] = struct{}{}
			}
			return nil
		})
	}

	metricCount := map[string]int{}

	csvfile, err := os.Open("./vals.csv")
	if err != nil {
		log.Fatalln(err)
	}
	r := csv.NewReader(csvfile)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		count, err := strconv.Atoi(record[1])
		if err != nil {
			log.Fatalln(err)
		}

		metricCount[record[0]] += count
	}

	// Final number of metrics required to run the queries.
	count := 0
	for metric := range metrics {
		count += metricCount[metric]
	}

	fmt.Println(count)
}

// for row in $(cat vals.json | jq -c '.data.result[]'); do
// 	_jq() {
// 	echo ${row} | jq -r ${1}
// 	}
// 	echo $(_jq '.metric.__name__ + "," + .value[1]')
// done > vals.csv
