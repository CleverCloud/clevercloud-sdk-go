package models_test

import (
	"encoding/json"
	"fmt"

	"go.clever-cloud.dev/sdk/models"
)

// Example_peerTypeSwitch shows the canonical way to consume a Peer received
// from the API: dispatch on Type(), call the matching As<Member>(), and have
// a default branch for forward-compatibility with future variants.
func Example_peerTypeSwitch() {
	payload := []byte(`{
		"id": "ng_demo",
		"label": "demo",
		"lastAllocatedIp": "10.0.0.1",
		"networkIp": "10.0.0.0/24",
		"ownerId": "user_001",
		"version": 1,
		"peers": [
			{
				"type": "CleverPeer",
				"id": "peer_a",
				"hostname": "a.example.com",
				"hv": "hv-1",
				"parentMember": "m1",
				"publicKey": "pk-a",
				"endpoint": {"type": "ClientEndpoint", "ngIp": "10.0.0.10"}
			},
			{
				"type": "ExternalPeer",
				"id": "peer_b",
				"hostname": "b.example.com",
				"parentMember": "m2",
				"publicKey": "pk-b",
				"endpoint": {"type": "ServerEndpoint", "ngTerm": "1.2.3.4:51820", "publicTerm": "5.6.7.8:51820"}
			},
			{
				"type": "FuturePeer",
				"id": "peer_c"
			}
		]
	}`)

	var ng models.NetworkGroup1
	if err := json.Unmarshal(payload, &ng); err != nil {
		fmt.Println("decode error:", err)
		return
	}

	for _, p := range ng.Peers {
		switch p.Type() {
		case models.CleverPeerType:
			// AsCleverPeer is a checked accessor — ok==false if the union
			// holds a different variant or the payload is malformed.
			cp, ok := p.AsCleverPeer()
			if !ok {
				fmt.Println("CleverPeer: failed to decode")
				continue
			}
			fmt.Printf("CleverPeer id=%s endpoint=%s\n", cp.ID, cp.Endpoint.Type())

		case models.ExternalPeerType:
			ep, ok := p.AsExternalPeer()
			if !ok {
				fmt.Println("ExternalPeer: failed to decode")
				continue
			}
			fmt.Printf("ExternalPeer id=%s endpoint=%s\n", ep.ID, ep.Endpoint.Type())

		case "":
			// Empty/null peer in the payload — treat as a no-op.
			fmt.Println("(empty peer)")

		default:
			// A new variant the SDK doesn't know about yet — log and move on
			// rather than crash. Use p.Type() to surface which one it was.
			fmt.Printf("unknown variant %q (raw payload still available via Marshal)\n", p.Type())
		}
	}

	// Output:
	// CleverPeer id=peer_a endpoint=ClientEndpoint
	// ExternalPeer id=peer_b endpoint=ServerEndpoint
	// unknown variant "FuturePeer" (raw payload still available via Marshal)
}

// Example_peerConstruction shows how to build a NetworkGroup1 payload with
// peers using the To<Union>() method — the discriminator is set for you.
func Example_peerConstruction() {
	ng := models.NetworkGroup1{
		ID:              "ng_demo",
		Label:           "demo",
		LastAllocatedIP: "10.0.0.1",
		NetworkIP:       "10.0.0.0/24",
		OwnerID:         "user_001",
		Version:         1,
		Peers: []models.Peer{
			models.CleverPeer{ID: "peer_a", PublicKey: "pk-a"}.ToPeer(),
			models.ExternalPeer{ID: "peer_b", PublicKey: "pk-b"}.ToPeer(),
		},
	}

	for _, p := range ng.Peers {
		fmt.Println(p.Type())
	}

	// Output:
	// CleverPeer
	// ExternalPeer
}
