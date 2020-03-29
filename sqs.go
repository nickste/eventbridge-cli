package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"os"

	"github.com/TylerBrock/colorjson"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/sqsiface"
)

type sqsClient struct {
	client sqsiface.ClientAPI

	arn       string
	queueName string
	queueURL  string
}

func newSQSClient(cfg aws.Config, accountID, queueName string) *sqsClient {
	return &sqsClient{
		client:    sqs.New(cfg),
		arn:       fmt.Sprintf("arn:aws:sqs:%s:%s:%s", cfg.Region, accountID, queueName),
		queueName: queueName,
	}
}

func (s *sqsClient) createQueue(ctx context.Context, ruleArn string) error {
	resp, err := s.client.CreateQueueRequest(&sqs.CreateQueueInput{
		QueueName: aws.String(s.queueName),
		Attributes: map[string]string{
			"Policy": fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Id": "%s",
				"Statement": [{
					"Sid": "Sid1579089564623",
					"Effect": "Allow",
					"Principal": {
						"Service": "events.amazonaws.com"
					},
					"Action": "SQS:SendMessage",
					"Resource": "%s",
					"Condition": {
						"ArnEquals": {
							"aws:SourceArn": "%s"
						}
					}
				}]
			}`, runID, s.arn, ruleArn),
		},
	}).Send(ctx)
	if err != nil {
		log.Printf("sqs.CreateQueue error: %s", err)
		return err
	}

	s.queueURL = *resp.QueueUrl
	return err
}

func (s *sqsClient) deleteQueue(ctx context.Context) error {
	_, err := s.client.DeleteQueueRequest(&sqs.DeleteQueueInput{
		QueueUrl: aws.String(s.queueURL),
	}).Send(ctx)
	if err != nil {
		log.Printf("sqs.DeleteQueue error: %s", err)
		return err
	}

	return err
}

func (s *sqsClient) pollQueue(ctx context.Context, breaker <-chan struct{}, prettyJSON bool, useSAM bool) error {
	log.Printf("polling queue %s ...", s.queueURL)
	log.Printf("press ctr+c to stop")

	for {
		// goroutine
		select {
		case <-breaker:
			log.Printf("stopping poller...")
			return nil

		default:
			resp, err := s.client.ReceiveMessageRequest(&sqs.ReceiveMessageInput{
				QueueUrl:              aws.String(s.queueURL),
				MaxNumberOfMessages:   aws.Int64(10),
				WaitTimeSeconds:       aws.Int64(5),
				MessageAttributeNames: []string{"All"},
			}).Send(ctx)
			if err != nil {
				log.Printf("sqs.ReceiveMessage error: %s", err)
				return err
			}

			if len(resp.Messages) == 0 {
				continue
			}

			entries := []sqs.DeleteMessageBatchRequestEntry{}
			for _, m := range resp.Messages {
				entries = append(entries, sqs.DeleteMessageBatchRequestEntry{
					Id:            m.MessageId,
					ReceiptHandle: m.ReceiptHandle,
				})

				if prettyJSON {
					var j map[string]interface{}
					err := json.Unmarshal([]byte(*m.Body), &j)
					if err != nil {
						return err
					}

					f := colorjson.NewFormatter()
					f.Indent = 2
					pj, _ := f.Marshal(j)

					log.Println(string(pj))
				} else {
					log.Printf("%s", *m.Body)
				}

				if useSAM {
					log.Printf("Invoking serverless application locally...\n")
					log.Printf("Returned data: \n")

					f, err := os.Create("events/event.json")
			    if err != nil {
			        fmt.Println(err)
			    }
			    f.WriteString(*m.Body)
			    
			    err = f.Close()
			    if err != nil {
			        fmt.Println(err)
			    }

					out, err := exec.Command("sam", "local", "invoke", "--event", "events/event.json").Output()
			    if err != nil {
			    	fmt.Println(err)
					}
					fmt.Printf("%s\n", out)
			  }
			}

			// cleanup messages
			_, err = s.client.DeleteMessageBatchRequest(&sqs.DeleteMessageBatchInput{
				QueueUrl: aws.String(s.queueURL),
				Entries:  entries,
			}).Send(ctx)
			if err != nil {
				log.Printf("sqs.DeleteMessageBatch error: %s", err)
			}
		}
	}
}
