# Message Type Detection

fmt provides automatic message type classification to help identify the nature of text content. The system detects common message types like errors, warnings, success messages, and information using zero-allocation buffer-based pattern matching.

```go
// Using StringType() (zero allocations)
message, msgType := Convert("Operation failed").StringType()
// message: "Operation failed", msgType: Msg.Error

// Real example - Progress callback with message classification
// Note: Requires import "github.com/tinywasm/fmt/lang" for Translate
progressCallback := func(msgs ...any) {
    message, msgType := lang.Translate(msgs...).StringType()
    if msgType.IsError() {
        handleError(message)
    } else {
        logMessage(message, msgType)
    }
}

// Message type constants available via Msg struct
if msgType.IsError() {
    // Handle error case
}

// Available message types:
// Msg.Normal    - Default type for general content
// Msg.Info      - Information messages
// Msg.Error     - Error messages and failures
// Msg.Warning   - Warning and caution messages
// Msg.Success   - Success and completion messages
// Msg.Debug     - Debugging messages
//
// Pub/Sub & Request/Response:
// Msg.Event     - Asynchronous events
// Msg.Request   - Synchronous requests
// Msg.Response  - Request responses
//
// Network/SSE specific:
// Msg.Connect   - Connection error
// Msg.Auth      - Authentication error
// Msg.Parse     - Parse/decode error
// Msg.Timeout   - Timeout error
// Msg.Broadcast - Broadcast/send error

// Zero allocations - reuses existing conversion buffers
// Perfect for logging, UI status messages, and error handling
```
