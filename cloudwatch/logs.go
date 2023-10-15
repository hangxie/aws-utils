package cloudwatch

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

func DumpLog(ctx context.Context, client *cloudwatchlogs.Client, logGroup string, logStream string, output io.Writer, timeFormat string) error {
	// get start and end time
	descStreamOutput, err := client.DescribeLogStreams(
		ctx,
		&cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName:        &logGroup,
			LogStreamNamePrefix: &logStream,
			Limit:               aws.Int32(1),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to DescribeLogStreams: %v", err)
	}
	if len(descStreamOutput.LogStreams) == 0 {
		return fmt.Errorf("log stream [%s] not found", logStream)
	}
	startTime := descStreamOutput.LogStreams[0].FirstEventTimestamp
	endTime := descStreamOutput.LogStreams[0].LastEventTimestamp
	includeTimestamp := timeFormat != ""

	getLogEventsInput := cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  &logGroup,
		LogStreamName: &logStream,
		StartTime:     startTime,
		EndTime:       endTime,
		Limit:         aws.Int32(10_000),
		NextToken:     nil,
	}
	for {
		getLogEventsOutput, err := client.GetLogEvents(ctx, &getLogEventsInput)
		if err != nil {
			return fmt.Errorf("failed to GetLogEvents: %v", err)
		}
		if len(getLogEventsOutput.Events) == 0 {
			break
		}
		for _, event := range getLogEventsOutput.Events {
			if event.Message == nil {
				continue
			}
			if includeTimestamp {
				fmt.Fprintf(output, "%s ", time.UnixMilli(*event.Timestamp).Format(timeFormat))
			}
			fmt.Fprintln(output, *event.Message)
		}
		if getLogEventsOutput.NextBackwardToken == nil {
			break
		}
		getLogEventsInput.NextToken = getLogEventsOutput.NextBackwardToken
	}

	return nil
}
