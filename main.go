package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/jmespath/go-jmespath"
	"github.com/logdyhq/logdy-core/logdy"
)

const (
	MaxResults = 100
)

var (
	stateMachineArn = flag.String("arn", "arn:aws:states:us-east-1:<account-id>:stateMachine:<state-machine-name>", "State Machine ARN")
	maxExecutions   = flag.Int("limit", 100, "Limit the number of executions to fetch")
	inputExpression = flag.String("expression", "{ id: detail.id, type: detail.type }", "JmesPath Input expression to transform execution input")
	serve           = flag.Bool("serve", false, "Serve Web UI")
	bindAddress     = flag.String("bind", "127.0.0.1:8080", "Web UI bind host")
)

func main() {
	flag.Parse()

	var logdyLogger logdy.Logdy

	if *serve {
		logdyLogger = logdy.InitializeLogdy(logdy.Config{
			HttpPathPrefix: "/",
		}, nil)
	}

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("error: %v", err)
	}

	// Create an Amazon StepFunctions service client
	client := sfn.NewFromConfig(cfg)

	listExecutionsPaginator := sfn.NewListExecutionsPaginator(client, &sfn.ListExecutionsInput{
		MaxResults:      MaxResults,
		StateMachineArn: stateMachineArn,
	})

	pageNum := 0
	for listExecutionsPaginator.HasMorePages() && pageNum < 3 {
		if pageNum*MaxResults > *maxExecutions {
			break
		}

		listExecutionsOutput, err := listExecutionsPaginator.NextPage(context.TODO())
		if err != nil {
			log.Printf("error: %v", err)
			return
		}
		pageNum++

		var order []string

		for _, value := range listExecutionsOutput.Executions {
			describeExecutionOutput, err := client.DescribeExecution(context.TODO(), &sfn.DescribeExecutionInput{ExecutionArn: value.ExecutionArn})
			if err != nil {
				log.Printf("error: %v", err)
				return
			}

			var input interface{}
			json.Unmarshal([]byte(*describeExecutionOutput.Input), &input)
			result, err := jmespath.Search(*inputExpression, input)
			if err != nil {
				log.Printf("error: %v", err)
				return
			}

			fields := logdy.Fields{
				"Name":      *value.Name,
				"StartDate": value.StartDate.Local().String()[:23],
				"StopDate":  value.StopDate.Local().String()[:23],
				"Status":    value.Status,
			}

			for k, v := range result.(map[string]any) {
				fields[k] = v
			}

			if order == nil {
				order = []string{"Name", "StartDate", "StopDate", "Status"}
				for k, _ := range result.(map[string]any) {
					order = append(order, k)
				}

				for _, k := range order {
					fmt.Print(k)
					fmt.Print(",")
				}
				fmt.Println()
			}

			for _, k := range order {
				fmt.Print(fields[k])
				fmt.Print(",")
			}
			fmt.Println()

			if *serve {
				logdyLogger.Log(fields)
			}
		}

		if *serve {
			log.Fatal(http.ListenAndServe(*bindAddress, nil))
		}
	}
}
