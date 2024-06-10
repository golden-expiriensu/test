## Description

On startup, the application declares `input-A`, `output-A`, `input-B`, and `output-B` queues and starts listening on `input-A` and `input-B`.
Upon delivery, the service processes the message (basically does `{"result": " + string(input) + "}`).
Processed messages are being published to the default exchange with routing keys `output-A` and `output-B`.

### Black box
Client publishes 'hello' to `input-A` and receives `{"result": "hello"}` from `output-A`.
Client publishes 'hello' to `input-B` and receives `{"result": "hello"}` from `output-B`.


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

## Test Using RabbitMQ Management UI
### Publish Messages
1. Open the RabbitMQ management UI in your browser at http://localhost:15672/.
2. Log in with the default credentials (guest/guest).
3. Go to the "Queues" tab and find the input-A and input-B queues.
4. For input-A, click on the queue name, then go to the "Publish message" section.
5. Enter hello in the "Payload" field and click "Publish message".
6. Repeat the above step for input-B.
### Consume Messages
1. Go to the "Queues" tab in the RabbitMQ management UI.
2. Find the output-A queue, click on it, and go to the "Get messages" section.
3. Click "Get Message(s)" to retrieve messages.
4. Repeat the above step for output-B.
### Verification
1. For input-A, you should receive a message in output-A with the body {"result": "hello"}.
2. For input-B, you should receive a message in output-B with the body {"result": "hello"}.