package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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

	for metric := range metrics {
		// Print only non recording rules.
		// if !strings.Contains(metric, ":") {
		fmt.Println(metric)
		// }
	}

}
