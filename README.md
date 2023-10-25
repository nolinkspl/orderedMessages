# orderedMessages

### Description
The app transfer messages from WebSocket connection wss://test-ws.skns.dev/raw-messages to another WebSocket connection wss://test-ws.skns.dev/ordered-messages/naguslaev. Transfer is processing in efficient way with using Priority Queue to sort messages with custom buffer size.

### Install and run:
- install Docker
- run command in the repository root: `docker build -t loader_app . && docker run loader_app`

### Realization explanation:
The code contains two services MessageLoader and MessageSender, to get messages and send it respectively. Each service cointains goroutines to process messages and PriorityQueue to store and sort. The Priority Queue helps us to find elements quickly with O(ln n) time complexity and O(C) space complexity because the size is limited.
Here is using Priority Queue with Mutex, Mutex allow use goroutines without racing condition.

### Configuration
Current configuration is optimal to handle messages without errors.
You can find it in constants in `src/services/message_sender.go`.
