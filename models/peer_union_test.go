package models_test

import (
	"encoding/json"
	"testing"

	"go.clever-cloud.dev/sdk/models"
)

func TestPeerUnionUnmarshalling(t *testing.T) {
	tests := []struct {
		name         string
		jsonData     string
		expectedType string
		validate     func(t *testing.T, peer models.Peer)
	}{
		{
			name: "CleverPeer deserialization",
			jsonData: `{
				"type": "CleverPeer",
				"endpoint": {
					"type": "ClientEndpoint",
					"ngIp": "192.168.1.100"
				},
				"hostname": "app-instance-1.clever-cloud.com",
				"hv": "hv-123",
				"id": "peer_clever_001",
				"label": "Production App",
				"parentEvent": "event_001",
				"parentMember": "member_001",
				"publicKey": "wg_public_key_clever_001"
			}`,
			expectedType: "CleverPeer",
			validate: func(t *testing.T, peer models.Peer) {
				cleverPeer, ok := peer.(*models.CleverPeer)
				if !ok {
					t.Fatalf("Expected *models.CleverPeer, got %T", peer)
				}
				if cleverPeer.GetType() != "CleverPeer" {
					t.Errorf("Expected type 'CleverPeer', got '%s'", cleverPeer.GetType())
				}
				if cleverPeer.ID != "peer_clever_001" {
					t.Errorf("Expected ID 'peer_clever_001', got '%s'", cleverPeer.ID)
				}
				if cleverPeer.Hostname != "app-instance-1.clever-cloud.com" {
					t.Errorf("Expected hostname 'app-instance-1.clever-cloud.com', got '%s'", cleverPeer.Hostname)
				}
				if cleverPeer.Hv != "hv-123" {
					t.Errorf("Expected Hv 'hv-123', got '%s'", cleverPeer.Hv)
				}
				if cleverPeer.Label == nil || *cleverPeer.Label != "Production App" {
					t.Errorf("Expected label 'Production App'")
				}
			},
		},
		{
			name: "ExternalPeer deserialization",
			jsonData: `{
				"type": "ExternalPeer",
				"endpoint": {
					"type": "ClientEndpoint",
					"ngIp": "203.0.113.50"
				},
				"hostname": "vpn-gateway.external.com",
				"id": "peer_external_002",
				"label": "External VPN Gateway",
				"parentEvent": "event_002",
				"parentMember": "member_002",
				"publicKey": "wg_public_key_external_002"
			}`,
			expectedType: "ExternalPeer",
			validate: func(t *testing.T, peer models.Peer) {
				externalPeer, ok := peer.(*models.ExternalPeer)
				if !ok {
					t.Fatalf("Expected *models.ExternalPeer, got %T", peer)
				}
				if externalPeer.GetType() != "ExternalPeer" {
					t.Errorf("Expected type 'ExternalPeer', got '%s'", externalPeer.GetType())
				}
				if externalPeer.ID != "peer_external_002" {
					t.Errorf("Expected ID 'peer_external_002', got '%s'", externalPeer.ID)
				}
				if externalPeer.Hostname != "vpn-gateway.external.com" {
					t.Errorf("Expected hostname 'vpn-gateway.external.com', got '%s'", externalPeer.Hostname)
				}
				if externalPeer.Label == nil || *externalPeer.Label != "External VPN Gateway" {
					t.Errorf("Expected label 'External VPN Gateway'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test using the idiomatic json.Unmarshaler interface
			var peerUnion models.PeerUnion
			err := json.Unmarshal([]byte(tt.jsonData), &peerUnion)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if peerUnion.Peer == nil {
				t.Fatal("Unmarshalled peer is nil")
			}

			if peerUnion.Peer.GetType() != tt.expectedType {
				t.Errorf("Expected type '%s', got '%s'", tt.expectedType, peerUnion.Peer.GetType())
			}

			tt.validate(t, peerUnion.Peer)
		})
	}
}

// TestPeerUnmarshallingWithHelper tests the backwards-compatible helper function
func TestPeerUnmarshallingWithHelper(t *testing.T) {
	jsonData := `{
		"type": "CleverPeer",
		"endpoint": {
			"type": "ClientEndpoint",
			"ngIp": "192.168.1.100"
		},
		"hostname": "app-instance-1.clever-cloud.com",
		"hv": "hv-123",
		"id": "peer_clever_001",
		"parentMember": "member_001",
		"publicKey": "wg_public_key_clever_001"
	}`

	peer, err := models.UnmarshalPeerJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if peer == nil {
		t.Fatal("Unmarshalled peer is nil")
	}

	if peer.GetType() != "CleverPeer" {
		t.Errorf("Expected type 'CleverPeer', got '%s'", peer.GetType())
	}
}

func TestPeerUnionMarshalling(t *testing.T) {
	t.Run("CleverPeer marshalling via PeerUnion", func(t *testing.T) {
		label := "Test Clever Peer"
		parentEvent := "event_123"
		cleverPeer := &models.CleverPeer{
			Endpoint: models.ClientEndpoint{
				NgIP: "10.0.0.1",
			},
			Hostname:     "test.clever-cloud.com",
			Hv:           "hv-test",
			ID:           "peer_test_001",
			Label:        &label,
			ParentEvent:  &parentEvent,
			ParentMember: "member_test",
			PublicKey:    "test_public_key",
		}

		// Wrap in PeerUnion and marshal using json.Marshaler interface
		peerUnion := models.NewPeerUnion(cleverPeer)
		data, err := json.Marshal(peerUnion)
		if err != nil {
			t.Fatalf("Failed to marshal PeerUnion: %v", err)
		}

		// Verify that the "type" field is present in JSON
		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("Failed to unmarshal into map: %v", err)
		}

		typeField, ok := raw["type"]
		if !ok {
			t.Error("Expected 'type' field in marshalled JSON")
		} else if typeField != "CleverPeer" {
			t.Errorf("Expected type 'CleverPeer', got '%v'", typeField)
		}
	})

	t.Run("ExternalPeer marshalling via PeerUnion", func(t *testing.T) {
		label := "Test External Peer"
		externalPeer := &models.ExternalPeer{
			Endpoint: models.ClientEndpoint{
				NgIP: "203.0.113.1",
			},
			Hostname:     "external.example.com",
			ID:           "peer_ext_001",
			Label:        &label,
			ParentMember: "member_external",
			PublicKey:    "ext_public_key",
		}

		// Wrap in PeerUnion and marshal using json.Marshaler interface
		peerUnion := models.NewPeerUnion(externalPeer)
		data, err := json.Marshal(peerUnion)
		if err != nil {
			t.Fatalf("Failed to marshal PeerUnion: %v", err)
		}

		// Verify that the "type" field is present in JSON
		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("Failed to unmarshal into map: %v", err)
		}

		typeField, ok := raw["type"]
		if !ok {
			t.Error("Expected 'type' field in marshalled JSON")
		} else if typeField != "ExternalPeer" {
			t.Errorf("Expected type 'ExternalPeer', got '%v'", typeField)
		}
	})

	t.Run("Direct CleverPeer marshalling", func(t *testing.T) {
		// Also test that concrete types can be marshalled directly
		label := "Direct Test"
		cleverPeer := &models.CleverPeer{
			Endpoint: models.ClientEndpoint{
				NgIP: "10.0.0.2",
			},
			Hostname:     "direct.clever-cloud.com",
			Hv:           "hv-direct",
			ID:           "peer_direct_001",
			Label:        &label,
			ParentMember: "member_direct",
			PublicKey:    "direct_public_key",
		}

		data, err := json.Marshal(cleverPeer)
		if err != nil {
			t.Fatalf("Failed to marshal CleverPeer: %v", err)
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("Failed to unmarshal into map: %v", err)
		}

		typeField, ok := raw["type"]
		if !ok {
			t.Error("Expected 'type' field in marshalled JSON")
		} else if typeField != "CleverPeer" {
			t.Errorf("Expected type 'CleverPeer', got '%v'", typeField)
		}
	})
}

func TestPeerUnionRoundTrip(t *testing.T) {
	label := "Round Trip Test"
	original := &models.CleverPeer{
		Endpoint: models.ClientEndpoint{
			NgIP: "192.168.100.1",
		},
		Hostname:     "roundtrip.clever-cloud.com",
		Hv:           "hv-roundtrip",
		ID:           "peer_roundtrip_001",
		Label:        &label,
		ParentMember: "member_roundtrip",
		PublicKey:    "roundtrip_public_key",
	}

	// Marshal
	peerUnion := models.NewPeerUnion(original)
	data, err := json.Marshal(peerUnion)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var decoded models.PeerUnion
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify
	cleverPeer, ok := decoded.Peer.(*models.CleverPeer)
	if !ok {
		t.Fatalf("Expected *models.CleverPeer, got %T", decoded.Peer)
	}

	if cleverPeer.ID != original.ID {
		t.Errorf("Expected ID '%s', got '%s'", original.ID, cleverPeer.ID)
	}
	if cleverPeer.Hostname != original.Hostname {
		t.Errorf("Expected Hostname '%s', got '%s'", original.Hostname, cleverPeer.Hostname)
	}
	if cleverPeer.Hv != original.Hv {
		t.Errorf("Expected Hv '%s', got '%s'", original.Hv, cleverPeer.Hv)
	}
}
