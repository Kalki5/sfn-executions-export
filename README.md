# AWS Step Functions Executor

This repository contains a Go application that lists and describes executions for a specified AWS Step Functions state machine.

## Features

- **List Executions**: Fetch a list of executions for a given AWS Step Functions state machine.
- **Transform Input**: Use a JMESPath expression to transform the input data of each execution.
- **Web UI**: Optionally serve a web interface that shows the logs.
- **Log Fields**: Customizable logging fields based on the input transformation.

## Prerequisites

Before you begin, ensure you have met the following requirements:

- You have a working Go environment.
- You have configured AWS credentials (e.g., via `~/.aws/credentials`).

## Usage

To use this application, clone the repository and compile the program:

```sh
git clone https://github.com/YOUR_GITHUB/aws-step-functions-executor.git
cd aws-step-functions-executor
go build .
./sfn-executions-export
```
