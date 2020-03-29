# ![logo](assets/logo.png) eventbridge-cli
[![Actions Status](https://github.com/spezam/eventbridge-cli/workflows/test/badge.svg)](https://github.com/spezam/eventbridge-cli/actions)

Amazon EventBridge is a serverless event bus that makes it easy to connect applications together using data from your own applications, integrated Software-as-a-Service (SaaS) applications, and AWS services.

Eventbridge-cli is a tool to listen to an EventBus events. Useful for debugging and event pattern testing.
```
EventBus --> EventBrige Rule --> SQS <-- poller
```

Features:
- Listen to Event Bus messages
- Filter messages by event pattern
- Authentication via profile or env variables
- Pretty JSON output
- ...

### Install from releases binary:
```
wget https://github.com/spezam/eventbridge-cli/releases/download/<version>/eventbridge-cli_<version>_darwin_amd64.tar.gz
tar xvfz eventbridge-cli_<version>_darwin_amd64.tar.gz
mv eventbridge-cli /somewhere/in/PATH
```
###or build from source:
```
go build
```

### Flags:
```
NAME:
   eventbridge-cli - AWS EventBridge cli

USAGE:
   eventbridge [global options] command [command options] [arguments...]

VERSION:
   0.1.0

AUTHOR:
   matteo ridolfi

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --profile value, -p value       AWS profile (default: "default") [$AWS_PROFILE]
   --region value, -r value        AWS region [$AWS_DEFAULT_REGION]
   --eventbusname value, -b value  EventBridge Bus Name (default: "default")
   --eventpattern value, -e value  EventBridge event pattern (default: "{\"source\": [{\"anything-but\": [\"eventbridge-cli\"]}]}")
   --prettyjson, -j                Pretty JSON output (default: false)
   --useSAM, -s                    Invokes a local serverless application in the current directory with the event from EventBridge
   --help, -h                      show help (default: false)
   --version, -v                   print the version (default: false)
```

### Usage example:
```sh
# with env variables
AWS_PROFILE=myawsprofile eventbridge-cli
AWS_PROFILE=myawsprofile AWS_DEFAULT_REGION=eu-north-1 eventbridge-cli

# with cli flags
eventbridge-cli --profile myawsprofile
eventbridge-cli --profile myawsprofile --region eu-north-1

# full example - with event pattern
eventbridge-cli -p myawsprofile -j \
	-b fishnchips-eventbus \
	-e '{"source":["gamma"],"detail":{"channel":["web"]}}'
```

![screenshot](assets/screenshot.png)

### Content-based Filtering with Event Patterns reference:
https://docs.aws.amazon.com/eventbridge/latest/userguide/content-filtering-with-event-patterns.html

Here is a summary of all the comparison operators available in EventBridge:

| Comparison | Example | Rule syntax  |
| ------------ |------------------ | --------------------|
| Null | UserID is null | "UserID": [ null ] |
| Empty | LastName is empty | "LastName": [""] |
| Equals | Name is "Alice" | "Name": [ "Alice" ] |
| And | Location is "New York" and Day is "Monday" | "Location": [ "New York" ], "Day": [ "Monday" ] |
| Or | PaymentType is "Credit" or "Debit" | "PaymentType": [ "Credit", "Debit" ] |
| Not | Weather is anything but "Raining" | "Weather": [{ "anything-but": [ "Raining" ] }] |
| Numeric (equals) | Price is 100 | "Price": [{ "numeric": [ "=", 100 ] }] |
| Numeric (range) | Price is more than 10, and less than or equal to 20 | "Price": [{ "numeric": [ ">", 10, "<=", 20 ] }] |
| Exists | ProductName exists | "ProductName": [{ "exists": true }] |
| Does not exist | ProductName does not exist | "ProductName": [{ "exists": false }] |
| Begins with | Region is in the US | "Region": [{ "prefix": "us-" }] |


