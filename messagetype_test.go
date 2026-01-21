package fmt

import (
	"testing"
)

func TestStringTypeDetection(t *testing.T) {
	t.Run("Empty string", func(t *testing.T) {
		msg, msgType := Convert("").StringType()
		if msgType != Msg.Normal {
			t.Errorf("Expected Normal for empty string, got %v", msgType)
		}
		if msg != "" {
			t.Errorf("Expected empty string, got %q", msg)
		}
	})

	t.Run("Error keywords", func(t *testing.T) {
		errorKeywords := []string{
			"This is an error message",
			"Operation failed",
			"exit status 1",
			"variable undeclared",
			"function undefined",
			"fatal exception",
		}
		for _, keyword := range errorKeywords {
			msg, msgType := Convert(keyword).StringType()
			if msgType != Msg.Error {
				t.Errorf("Expected Error for keyword %q, got %v", keyword, msgType)
			}
			if msg != keyword {
				t.Errorf("Expected message to be unchanged, got %q", msg)
			}
		}
	})

	t.Run("Success keywords", func(t *testing.T) {
		successKeywords := []string{
			"Success! Operation completed",
			"success",
			"Operation completed",
			"Build successful",
			"Task done",
		}
		for _, keyword := range successKeywords {
			msg, msgType := Convert(keyword).StringType()
			if msgType != Msg.Success {
				t.Errorf("Expected Success for keyword %q, got %v", keyword, msgType)
			}
			if msg != keyword {
				t.Errorf("Expected message to be unchanged, got %q", msg)
			}
		}
	})

	t.Run("Info keywords", func(t *testing.T) {
		infoKeywords := []string{
			"Info: Starting process",
			"... initializing ...",
			"starting up",
			"initializing system",
		}
		for _, keyword := range infoKeywords {
			_, msgType := Convert(keyword).StringType()
			if msgType != Msg.Info {
				t.Errorf("Expected Info for keyword %q, got %v", keyword, msgType)
			}
		}
	})

	t.Run("Warning keywords", func(t *testing.T) {
		warningKeywords := []string{
			"Warning: disk space low",
			"warn user",
		}
		for _, keyword := range warningKeywords {
			_, msgType := Convert(keyword).StringType()
			if msgType != Msg.Warning {
				t.Errorf("Expected Warning for keyword %q, got %v", keyword, msgType)
			}
		}
	})

	t.Run("Debug keywords", func(t *testing.T) {
		debugKeywords := []string{
			"debug: something happening",
			"[debug] status",
			"DEBUG: uppercase",
			"DeBuG: Mixed Case",
		}
		for _, keyword := range debugKeywords {
			_, msgType := Convert(keyword).StringType()
			if msgType != Msg.Debug {
				t.Errorf("Expected Debug for keyword %q, got %v", keyword, msgType)
			}
		}
	})

	t.Run("Normal message", func(t *testing.T) {
		_, msgType := Convert("Hello world").StringType()
		if msgType != Msg.Normal {
			t.Errorf("Expected Normal for generic message, got %v", msgType)
		}
	})
}

func TestSSERelatedTypes(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		check    func(MessageType) bool
		expected string
	}{
		{"Connect", Msg.Connect, func(t MessageType) bool { return t.IsConnect() }, "Connect"},
		{"Auth", Msg.Auth, func(t MessageType) bool { return t.IsAuth() }, "Auth"},
		{"Parse", Msg.Parse, func(t MessageType) bool { return t.IsParse() }, "Parse"},
		{"Timeout", Msg.Timeout, func(t MessageType) bool { return t.IsTimeout() }, "Timeout"},
		{"Broadcast", Msg.Broadcast, func(t MessageType) bool { return t.IsBroadcast() }, "Broadcast"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(tt.msgType) {
				t.Errorf("Expected check for %s to be true", tt.name)
			}
			if tt.msgType.String() != tt.expected {
				t.Errorf("Expected String() for %s to be %q, got %q", tt.name, tt.expected, tt.msgType.String())
			}
		})
	}

	t.Run("IsNetworkError", func(t *testing.T) {
		networkTypes := []MessageType{Msg.Connect, Msg.Auth, Msg.Timeout, Msg.Broadcast}
		for _, nt := range networkTypes {
			if !nt.IsNetworkError() {
				t.Errorf("Expected %v to be a network error", nt)
			}
		}

		nonNetworkTypes := []MessageType{Msg.Normal, Msg.Info, Msg.Error, Msg.Warning, Msg.Success, Msg.Parse, Msg.Debug}
		for _, nnt := range nonNetworkTypes {
			if nnt.IsNetworkError() {
				t.Errorf("Expected %v NOT to be a network error", nnt)
			}
		}
	})

	t.Run("IsDebug", func(t *testing.T) {
		if !Msg.Debug.IsDebug() {
			t.Error("Expected Msg.Debug.IsDebug() to be true")
		}
		if Msg.Info.IsDebug() {
			t.Error("Expected Msg.Info.IsDebug() to be false")
		}
	})
}
