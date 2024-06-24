## Description

On startup application declares `input-A` and `output-A` queues and starts listen on `input-A`.
Upon delivery service processes it (basically does `{"result": " + string(input) + "}`).
Processed message is being published to default exchange with routing key `output-A`.

### Black box
Client publishes 'hello' to `input-A` and receives '{"result": "hello"}' from `output-A`.


## How to

### Lint
Format the code, fix dependencies and validate the project.
```bash
make lint
```

### Run
Bootstrap RabbitMQ, run the service and start processing messages received from the queue.
```bash
make
```

## Objective

Add support for the second stream of data using `input-B` and `output-B` queues.

### Black box
1. Client publishes 'hello' to `input-A` and receives '{"result": "hello"}' from `output-A`.
2. Client publishes 'hello' to `input-B` and receives '{"result": "hello"}' from `output-B`.


### Testing
## Run in a terminal
`make`

## Publishing
Open your web browser at http://localhost:15672/.
Use: username guest and password guest.
Click on the "Queues" tab to view all queues.
Locate the input-A queue and click on its name to open its details.
Go to "Publish message" section.
In the "Payload" field, type helloA and click the "Publish message" button.
Do the same for the input-B queue using helloB in the payload.

## Consuming
Go back to the "Queues" tab.
Find the output-A queue and click on its name to open its details.
Go to "Get messages" section.
Click the "Get Message(s)" button to retrieve messages from the queue.
Do the same for the output-B queue.

## Test resut
Queue input-A, the message in output-A should be {"result": "helloA"}.
Queue input-B, the message in output-B should be {"result": "helloB"}.