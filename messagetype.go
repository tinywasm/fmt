package fmt

// MessageType represents the classification of message types in the system.
type MessageType uint8

// Msg exposes the MessageType constants for external use, following fmt naming convention.
// Msg exposes the MessageType constants for external use, following fmt naming convention.
var Msg = struct {
	Normal  MessageType
	Info    MessageType
	Error   MessageType
	Warning MessageType
	Success MessageType

	// Network/SSE specific (new)
	Connect   MessageType // Connection error
	Auth      MessageType // Authentication error
	Parse     MessageType // Parse/decode error
	Timeout   MessageType // Timeout error
	Broadcast MessageType // Broadcast/send error
	Debug     MessageType // / Debug message

	// Pub/Sub & Request/Response (new)
	Event    MessageType
	Request  MessageType
	Response MessageType
}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}

// Helper methods for MessageType
func (t MessageType) IsNormal() bool   { return t == Msg.Normal }
func (t MessageType) IsInfo() bool     { return t == Msg.Info }
func (t MessageType) IsError() bool    { return t == Msg.Error }
func (t MessageType) IsWarning() bool  { return t == Msg.Warning }
func (t MessageType) IsSuccess() bool  { return t == Msg.Success }
func (t MessageType) IsDebug() bool    { return t == Msg.Debug }
func (t MessageType) IsEvent() bool    { return t == Msg.Event }
func (t MessageType) IsRequest() bool  { return t == Msg.Request }
func (t MessageType) IsResponse() bool { return t == Msg.Response }

// Network/SSE helper methods
func (t MessageType) IsConnect() bool   { return t == Msg.Connect }
func (t MessageType) IsAuth() bool      { return t == Msg.Auth }
func (t MessageType) IsParse() bool     { return t == Msg.Parse }
func (t MessageType) IsTimeout() bool   { return t == Msg.Timeout }
func (t MessageType) IsBroadcast() bool { return t == Msg.Broadcast }

// IsNetworkError returns true for any network-related error type
func (t MessageType) IsNetworkError() bool {
	return t == Msg.Connect || t == Msg.Auth || t == Msg.Timeout || t == Msg.Broadcast
}

func (t MessageType) String() string {
	switch t {
	case Msg.Info:
		return "Info"
	case Msg.Error:
		return "Error"
	case Msg.Warning:
		return "Warning"
	case Msg.Success:
		return "Success"
	case Msg.Connect:
		return "Connect"
	case Msg.Auth:
		return "Auth"
	case Msg.Parse:
		return "Parse"
	case Msg.Timeout:
		return "Timeout"
	case Msg.Broadcast:
		return "Broadcast"
	case Msg.Debug:
		return "Debug"
	case Msg.Event:
		return "Event"
	case Msg.Request:
		return "Request"
	case Msg.Response:
		return "Response"
	default:
		return "Normal"
	}
}

// Pre-compiled patterns for efficient buffer matching
var (
	errorPatterns = [][]byte{
		[]byte("error"), []byte("failed"), []byte("exit status 1"),
		[]byte("undeclared"), []byte("undefined"), []byte("fatal"),
	}
	warningPatterns = [][]byte{
		[]byte("warning"), []byte("warn"),
	}
	debugPatterns = [][]byte{
		[]byte("debug"),
	}
	successPatterns = [][]byte{
		[]byte("success"), []byte("completed"), []byte("successful"), []byte("done"),
	}
	infoPatterns = [][]byte{
		[]byte("info"), []byte("starting"), []byte("initializing"),
	}
)

// StringType returns the string from BuffOut and its detected MessageType, then auto-releases the Conv
func (c *Conv) StringType() (string, MessageType) {
	// Get string content FIRST (before detection modifies buffer)
	out := c.GetString(BuffOut)
	// Detect type from BuffOut content
	msgType := c.detectMessageTypeFromBuffer(BuffOut)
	// Auto-release
	c.putConv()
	return out, msgType
}

// detectMessageTypeFromBuffer analyzes the buffer content and returns the detected MessageType (zero allocations)
func (c *Conv) detectMessageTypeFromBuffer(dest BuffDest) MessageType {
	// 1. Copy content directly to work buffer using swapBuff (zero allocations)
	c.swapBuff(dest, BuffWork)
	// 2. Convert to lowercase in work buffer using existing method
	c.changeCase(true, BuffWork)
	// 3. Direct buffer pattern matching - NO Contains() allocations
	if c.bufferContainsPattern(BuffWork, errorPatterns) {
		return Msg.Error
	}
	if c.bufferContainsPattern(BuffWork, warningPatterns) {
		return Msg.Warning
	}
	if c.bufferContainsPattern(BuffWork, successPatterns) {
		return Msg.Success
	}
	if c.bufferContainsPattern(BuffWork, infoPatterns) {
		return Msg.Info
	}
	if c.bufferContainsPattern(BuffWork, debugPatterns) {
		return Msg.Debug
	}
	return Msg.Normal
}
