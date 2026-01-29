// Manual implementation for WireguardEndpoint union type
// This file is NOT generated and should be maintained manually

package models

import (
	"encoding/json"
	"fmt"
)

// WireguardEndpointUnion is a wrapper type that implements json.Marshaler and json.Unmarshaler
// for the WireguardEndpoint interface union type.
type WireguardEndpointUnion struct {
	Endpoint WireguardEndpoint
}

// NewWireguardEndpointUnion creates a new WireguardEndpointUnion wrapping the given WireguardEndpoint
func NewWireguardEndpointUnion(e WireguardEndpoint) WireguardEndpointUnion {
	return WireguardEndpointUnion{Endpoint: e}
}

// MarshalJSON implements json.Marshaler for WireguardEndpointUnion
func (w WireguardEndpointUnion) MarshalJSON() ([]byte, error) {
	if w.Endpoint == nil {
		return []byte("null"), nil
	}
	// Delegate to the concrete type's MarshalJSON
	return json.Marshal(w.Endpoint)
}

// UnmarshalJSON implements json.Unmarshaler for WireguardEndpointUnion
func (w *WireguardEndpointUnion) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		w.Endpoint = nil
		return nil
	}

	// Detect discriminator
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal wireguard endpoint envelope: %w", err)
	}

	// Direct unmarshalling based on discriminator
	switch envelope.Type {
	case ClientEndpointType:
		var clientEndpoint ClientEndpoint
		if err := json.Unmarshal(data, &clientEndpoint); err != nil {
			return fmt.Errorf("failed to unmarshal ClientEndpoint: %w", err)
		}
		w.Endpoint = clientEndpoint
		return nil

	case ServerEndpointType:
		var serverEndpoint ServerEndpoint
		if err := json.Unmarshal(data, &serverEndpoint); err != nil {
			return fmt.Errorf("failed to unmarshal ServerEndpoint: %w", err)
		}
		w.Endpoint = serverEndpoint
		return nil

	default:
		return fmt.Errorf("unknown wireguard endpoint type: %s (expected %s or %s)",
			envelope.Type, ClientEndpointType, ServerEndpointType)
	}
}

// UnmarshalWireguardEndpointJSON is a convenience function for backwards compatibility
// Prefer using WireguardEndpointUnion with json.Unmarshal directly
func UnmarshalWireguardEndpointJSON(data []byte) (WireguardEndpoint, error) {
	var w WireguardEndpointUnion
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, err
	}
	return w.Endpoint, nil
}

// MarshalJSON implements json.Marshaler for ClientEndpoint
// Ensures the Type field is always set correctly
func (c ClientEndpoint) MarshalJSON() ([]byte, error) {
	type Alias ClientEndpoint
	return json.Marshal(&struct {
		*Alias
		Type string `json:"type"`
	}{
		Alias: (*Alias)(&c),
		Type:  ClientEndpointType,
	})
}

// UnmarshalJSON implements json.Unmarshaler for ClientEndpoint
// Validates and sets the Type field
func (c *ClientEndpoint) UnmarshalJSON(data []byte) error {
	type Alias ClientEndpoint
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Validate or set Type
	if c.Type == "" {
		c.Type = ClientEndpointType
	} else if c.Type != ClientEndpointType {
		return fmt.Errorf("invalid type for ClientEndpoint: expected %s, got %s",
			ClientEndpointType, c.Type)
	}

	return nil
}

// MarshalJSON implements json.Marshaler for ServerEndpoint
// Ensures the Type field is always set correctly
func (s ServerEndpoint) MarshalJSON() ([]byte, error) {
	type Alias ServerEndpoint
	return json.Marshal(&struct {
		*Alias
		Type string `json:"type"`
	}{
		Alias: (*Alias)(&s),
		Type:  ServerEndpointType,
	})
}

// UnmarshalJSON implements json.Unmarshaler for ServerEndpoint
// Validates and sets the Type field
func (s *ServerEndpoint) UnmarshalJSON(data []byte) error {
	type Alias ServerEndpoint
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Validate or set Type
	if s.Type == "" {
		s.Type = ServerEndpointType
	} else if s.Type != ServerEndpointType {
		return fmt.Errorf("invalid type for ServerEndpoint: expected %s, got %s",
			ServerEndpointType, s.Type)
	}

	return nil
}
