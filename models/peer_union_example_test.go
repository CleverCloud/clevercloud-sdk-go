package models_test

import (
	"encoding/json"
	"fmt"

	"go.clever-cloud.dev/sdk/models"
)

// ExamplePeerUnion_unmarshal demonstrates how to unmarshal a Peer union type from JSON
func ExamplePeerUnion_unmarshal() {
	jsonData := `{
		"type": "CleverPeer",
		"endpoint": {
			"type": "ClientEndpoint",
			"ngIp": "192.168.1.100"
		},
		"hostname": "app.clever-cloud.com",
		"hv": "hv-paris-1",
		"id": "peer_001",
		"parentMember": "member_001",
		"publicKey": "public_key_001"
	}`

	// Use the standard json.Unmarshaler interface
	var peerUnion models.PeerUnion
	if err := json.Unmarshal([]byte(jsonData), &peerUnion); err != nil {
		panic(err)
	}

	// Access the concrete Peer
	peer := peerUnion.Peer
	fmt.Printf("Type: %s\n", peer.GetType())

	// Type assertion if you need specific fields
	if cleverPeer, ok := peer.(*models.CleverPeer); ok {
		fmt.Printf("Hostname: %s\n", cleverPeer.Hostname)
		fmt.Printf("Hv: %s\n", cleverPeer.Hv)
	}

	// Output:
	// Type: CleverPeer
	// Hostname: app.clever-cloud.com
	// Hv: hv-paris-1
}

// ExamplePeerUnion_marshal demonstrates how to marshal a Peer union type to JSON
func ExamplePeerUnion_marshal() {
	// Create a concrete peer
	cleverPeer := &models.CleverPeer{
		Endpoint: models.ClientEndpoint{
			NgIP: "10.0.0.1",
		},
		Hostname:     "test.clever-cloud.com",
		Hv:           "hv-test",
		ID:           "peer_test",
		ParentMember: "member_test",
		PublicKey:    "test_key",
	}

	// Wrap in PeerUnion and marshal using standard json.Marshaler
	peerUnion := models.NewPeerUnion(cleverPeer)
	data, err := json.Marshal(peerUnion)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	// Output will include: "type":"CleverPeer"
}

// ExamplePeerUnion_slice demonstrates handling a slice of Peer union types
// This is useful when deserializing NetworkGroup.Peers
func ExamplePeerUnion_slice() {
	// Example JSON with mixed peer types
	jsonData := `[
		{
			"type": "CleverPeer",
			"endpoint": {"type": "ClientEndpoint", "ngIp": "10.0.0.1"},
			"hostname": "app1.clever-cloud.com",
			"hv": "hv-1",
			"id": "peer_1",
			"parentMember": "member_1",
			"publicKey": "key_1"
		},
		{
			"type": "ExternalPeer",
			"endpoint": {"type": "ClientEndpoint", "ngIp": "10.0.0.2"},
			"hostname": "external.example.com",
			"id": "peer_2",
			"parentMember": "member_2",
			"publicKey": "key_2"
		}
	]`

	// Unmarshal as slice of PeerUnion
	var peers []models.PeerUnion
	if err := json.Unmarshal([]byte(jsonData), &peers); err != nil {
		panic(err)
	}

	// Process each peer
	for i, peerUnion := range peers {
		peer := peerUnion.Peer
		fmt.Printf("Peer %d: Type=%s\n", i+1, peer.GetType())

		switch p := peer.(type) {
		case *models.CleverPeer:
			fmt.Printf("  CleverPeer: %s (Hv: %s)\n", p.Hostname, p.Hv)
		case *models.ExternalPeer:
			fmt.Printf("  ExternalPeer: %s\n", p.Hostname)
		}
	}

	// Output:
	// Peer 1: Type=CleverPeer
	//   CleverPeer: app1.clever-cloud.com (Hv: hv-1)
	// Peer 2: Type=ExternalPeer
	//   ExternalPeer: external.example.com
}

// ExamplePeerUnion_backwardsCompatible shows the backwards-compatible helper function
func ExamplePeerUnion_backwardsCompatible() {
	jsonData := `{
		"type": "ExternalPeer",
		"endpoint": {"type": "ClientEndpoint", "ngIp": "203.0.113.1"},
		"hostname": "vpn.external.com",
		"id": "peer_ext",
		"parentMember": "member_ext",
		"publicKey": "ext_key"
	}`

	// Using the helper function (backwards compatible)
	peer, err := models.UnmarshalPeerJSON([]byte(jsonData))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Type: %s\n", peer.GetType())

	if externalPeer, ok := peer.(*models.ExternalPeer); ok {
		fmt.Printf("Hostname: %s\n", externalPeer.Hostname)
	}

	// Output:
	// Type: ExternalPeer
	// Hostname: vpn.external.com
}
