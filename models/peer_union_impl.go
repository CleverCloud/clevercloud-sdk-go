// Manual implementation for Peer union type deserialization
// This file is NOT generated and should be maintained manually

package models

import (
	"encoding/json"
	"fmt"
)

// PeerUnion is a wrapper type that implements json.Marshaler and json.Unmarshaler
// for the Peer interface union type.
type PeerUnion struct {
	Peer Peer
}

// NewPeerUnion creates a new PeerUnion wrapping the given Peer
func NewPeerUnion(p Peer) PeerUnion {
	return PeerUnion{Peer: p}
}

// MarshalJSON implements json.Marshaler for PeerUnion
func (p PeerUnion) MarshalJSON() ([]byte, error) {
	if p.Peer == nil {
		return []byte("null"), nil
	}
	// Delegate to the concrete type's MarshalJSON
	return json.Marshal(p.Peer)
}

// UnmarshalJSON implements json.Unmarshaler for PeerUnion
func (p *PeerUnion) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		p.Peer = nil
		return nil
	}

	// Detect discriminator
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal peer envelope: %w", err)
	}

	// Direct unmarshalling based on discriminator
	switch envelope.Type {
	case CleverPeerType:
		var cleverPeer CleverPeer
		if err := json.Unmarshal(data, &cleverPeer); err != nil {
			return fmt.Errorf("failed to unmarshal CleverPeer: %w", err)
		}
		p.Peer = &cleverPeer
		return nil

	case ExternalPeerType:
		var externalPeer ExternalPeer
		if err := json.Unmarshal(data, &externalPeer); err != nil {
			return fmt.Errorf("failed to unmarshal ExternalPeer: %w", err)
		}
		p.Peer = &externalPeer
		return nil

	default:
		return fmt.Errorf("unknown peer type: %s (expected %s or %s)",
			envelope.Type, CleverPeerType, ExternalPeerType)
	}
}

// UnmarshalPeerJSON is a convenience function for backwards compatibility
// Prefer using PeerUnion with json.Unmarshal directly
func UnmarshalPeerJSON(data []byte) (Peer, error) {
	var p PeerUnion
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return p.Peer, nil
}

// MarshalJSON implements json.Marshaler for CleverPeer
// Ensures the Type field is always set correctly
func (c CleverPeer) MarshalJSON() ([]byte, error) {
	type Alias CleverPeer
	return json.Marshal(&struct {
		*Alias
		Type string `json:"type"`
	}{
		Alias: (*Alias)(&c),
		Type:  CleverPeerType,
	})
}

// UnmarshalJSON implements json.Unmarshaler for CleverPeer
// Validates and sets the Type field, and handles WireguardEndpoint union
func (c *CleverPeer) UnmarshalJSON(data []byte) error {
	// Use an auxiliary type to unmarshal everything except Endpoint
	type Alias CleverPeer
	aux := &struct {
		Endpoint json.RawMessage `json:"endpoint"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal the Endpoint field separately as it's a union type
	var endpointUnion WireguardEndpointUnion
	if err := json.Unmarshal(aux.Endpoint, &endpointUnion); err != nil {
		return fmt.Errorf("failed to unmarshal endpoint: %w", err)
	}
	c.Endpoint = endpointUnion.Endpoint

	// Validate or set Type
	if c.Type == "" {
		c.Type = CleverPeerType
	} else if c.Type != CleverPeerType {
		return fmt.Errorf("invalid type for CleverPeer: expected %s, got %s",
			CleverPeerType, c.Type)
	}

	return nil
}

// MarshalJSON implements json.Marshaler for ExternalPeer
// Ensures the Type field is always set correctly
func (e ExternalPeer) MarshalJSON() ([]byte, error) {
	type Alias ExternalPeer
	return json.Marshal(&struct {
		*Alias
		Type string `json:"type"`
	}{
		Alias: (*Alias)(&e),
		Type:  ExternalPeerType,
	})
}

// UnmarshalJSON implements json.Unmarshaler for ExternalPeer
// Validates and sets the Type field, and handles WireguardEndpoint union
func (e *ExternalPeer) UnmarshalJSON(data []byte) error {
	// Use an auxiliary type to unmarshal everything except Endpoint
	type Alias ExternalPeer
	aux := &struct {
		Endpoint json.RawMessage `json:"endpoint"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal the Endpoint field separately as it's a union type
	var endpointUnion WireguardEndpointUnion
	if err := json.Unmarshal(aux.Endpoint, &endpointUnion); err != nil {
		return fmt.Errorf("failed to unmarshal endpoint: %w", err)
	}
	e.Endpoint = endpointUnion.Endpoint

	// Validate or set Type
	if e.Type == "" {
		e.Type = ExternalPeerType
	} else if e.Type != ExternalPeerType {
		return fmt.Errorf("invalid type for ExternalPeer: expected %s, got %s",
			ExternalPeerType, e.Type)
	}

	return nil
}
