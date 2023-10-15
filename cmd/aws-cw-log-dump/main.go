package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/rs/zerolog/log"

	"github.com/hangxie/aws-utils/cloudwatch"
)

type Cli struct {
	LogGroup  string `required:"" help:"CloudWatch log group name, eg /cxdl/prd/log-group-task-us-west-2"`
	LogStream string `required:"" help:"CloudWatch log stream, eg /cxdl-prd-cxca-ing-proc/cxca-container/3c80fc995893477b97d0ec666fd0bd93"`
	Timestamp bool   `help:"Prefix timestamp to each line of log"`
}

func main() {
	cli := Cli{}
	_ = kong.Parse(&cli,
		kong.Name(os.Args[0]),
		kong.Description("Dump AWS CloudWatch log stream"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}))

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get AWS session")
		os.Exit(1)
	}

	client := cloudwatchlogs.NewFromConfig(cfg)
	timeFormat := "2006-01-02T15:04:05.000Z0700"
	if !cli.Timestamp {
		timeFormat = ""
	}
	if err := cloudwatch.DumpLog(context.Background(), client, cli.LogGroup, cli.LogStream, os.Stdout, timeFormat); err != nil {
		fmt.Printf("failed to dump logs: %v", err)
		os.Exit(1)
	}
}
