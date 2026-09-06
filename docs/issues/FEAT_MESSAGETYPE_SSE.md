# Feature: Extend MessageType for SSE Error Types

> **Status:** Approved  
> **Created:** December 2025  
> **Related:** `tinysse` package SSE implementation

## Context

El paquete `tinysse` necesita tipos de error para SSE. En lugar de crear nuevos tipos, extendemos `MessageType` existente para reutilizar la infraestructura.

## Current State

```go
// messagetype.go
var Msg = struct {
    Normal  MessageType  // 0
    Info    MessageType  // 1
    Error   MessageType  // 2
    Warning MessageType  // 3
    Success MessageType  // 4
}{0, 1, 2, 3, 4}
```

## Proposed Extension

Agregar tipos específicos para errores de red/SSE:

```go
var Msg = struct {
    Normal     MessageType  // 0 - Default
    Info       MessageType  // 1 - Information
    Error      MessageType  // 2 - General error
    Warning    MessageType  // 3 - Warning
    Success    MessageType  // 4 - Success
    
    // Network/SSE specific (new)
    Connect    MessageType  // 5 - Connection error
    Auth       MessageType  // 6 - Authentication error
    Parse      MessageType  // 7 - Parse/decode error
    Timeout    MessageType  // 8 - Timeout error
    Broadcast  MessageType  // 9 - Broadcast/send error
}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
```

## Usage in TinySSE

```go
// tinysse/error.go
import . "webtyp.com/fmt"

type SSEError struct {
    Type    MessageType  // Reuse tinystring.MessageType
    Err     error
    Context any          // clientID, raw data, etc.
}

// Config callback
type Config struct {
    OnError func(SSEError)
}

// Usage
cfg.OnError = func(e SSEError) {
    switch {
    case e.Type == Msg.Auth:
        // Redirect to login
    case e.Type == Msg.Connect:
        // Show offline banner
    case e.Type.IsError():
        // General error handling
    }
}
```

## Benefits

1. **Reutilización** - No duplicar tipos
2. **Consistencia** - Mismo sistema en todo el stack
3. **Zero allocations** - `MessageType` es `uint8`
4. **Helpers existentes** - `IsError()`, `String()`, etc.

## New Helper Methods

```go
// Add to messagetype.go
func (t MessageType) IsConnect() bool   { return t == Msg.Connect }
func (t MessageType) IsAuth() bool      { return t == Msg.Auth }
func (t MessageType) IsParse() bool     { return t == Msg.Parse }
func (t MessageType) IsTimeout() bool   { return t == Msg.Timeout }
func (t MessageType) IsBroadcast() bool { return t == Msg.Broadcast }

// IsNetworkError returns true for any network-related error type
func (t MessageType) IsNetworkError() bool {
    return t == Msg.Connect || t == Msg.Auth || t == Msg.Timeout || t == Msg.Broadcast
}

// Update String() method
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
    default:
        return "Normal"
    }
}
```

## Implementation Steps

1. [ ] Extend `Msg` struct in `messagetype.go`
2. [ ] Add helper methods (`IsConnect()`, etc.)
3. [ ] Update `String()` method
4. [ ] Add tests for new types
5. [ ] Use in `tinysse` package

## Backward Compatibility

✅ **Fully compatible** - Only adds new constants, existing code unchanged.
