package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// TestPeerUnmarshalDispatch verifies that []Peer fields decode correctly and
// that the held value is recoverable via the As<Member>() accessors.
func TestPeerUnmarshalDispatch(t *testing.T) {
	payload := `{
		"id": "ng_123",
		"label": "demo",
		"lastAllocatedIp": "10.0.0.1",
		"networkIp": "10.0.0.0/24",
		"ownerId": "user_001",
		"version": 1,
		"peers": [
			{
				"type": "CleverPeer",
				"id": "peer_clever",
				"hostname": "host.example.com",
				"hv": "hv-1",
				"parentMember": "member_a",
				"publicKey": "pk-clever",
				"endpoint": {"type": "ClientEndpoint", "ngIp": "10.0.0.10"}
			},
			{
				"type": "ExternalPeer",
				"id": "peer_external",
				"hostname": "ext.example.com",
				"parentMember": "member_b",
				"publicKey": "pk-external",
				"endpoint": {"type": "ServerEndpoint", "ngTerm": "1.2.3.4:51820", "publicTerm": "5.6.7.8:51820"}
			}
		]
	}`

	var ng NetworkGroup1
	if err := json.Unmarshal([]byte(payload), &ng); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(ng.Peers) != 2 {
		t.Fatalf("expected 2 peers, got %d", len(ng.Peers))
	}

	if got := ng.Peers[0].Type(); got != CleverPeerType {
		t.Errorf("first peer Type() = %q, want %q", got, CleverPeerType)
	}
	clever, ok := ng.Peers[0].AsCleverPeer()
	if !ok {
		t.Fatalf("first peer AsCleverPeer returned ok=false")
	}
	if clever.ID != "peer_clever" {
		t.Errorf("clever.ID = %q, want peer_clever", clever.ID)
	}
	clientEnd, ok := clever.Endpoint.AsClientEndpoint()
	if !ok {
		t.Fatalf("nested endpoint AsClientEndpoint returned ok=false (Type=%q)", clever.Endpoint.Type())
	}
	if clientEnd.NgIP != "10.0.0.10" {
		t.Errorf("clientEnd.NgIP = %q, want 10.0.0.10", clientEnd.NgIP)
	}

	if got := ng.Peers[1].Type(); got != ExternalPeerType {
		t.Errorf("second peer Type() = %q, want %q", got, ExternalPeerType)
	}
	external, ok := ng.Peers[1].AsExternalPeer()
	if !ok {
		t.Fatalf("second peer AsExternalPeer returned ok=false")
	}
	if external.ID != "peer_external" {
		t.Errorf("external.ID = %q, want peer_external", external.ID)
	}

	// Wrong-variant accessor must return ok=false.
	if _, ok := ng.Peers[0].AsExternalPeer(); ok {
		t.Errorf("CleverPeer.AsExternalPeer() should return ok=false")
	}
}

// TestPeerUnmarshalEmptyPeers ensures missing/empty/null peer slices don't error.
func TestPeerUnmarshalEmptyPeers(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		wantLen int
	}{
		{
			name:    "no peers field",
			payload: `{"id":"ng_1","label":"x","lastAllocatedIp":"10.0.0.1","networkIp":"10.0.0.0/24","ownerId":"u","version":1}`,
			wantLen: 0,
		},
		{
			name:    "explicit empty array",
			payload: `{"id":"ng_1","label":"x","lastAllocatedIp":"10.0.0.1","networkIp":"10.0.0.0/24","ownerId":"u","version":1,"peers":[]}`,
			wantLen: 0,
		},
		{
			name:    "explicit null",
			payload: `{"id":"ng_1","label":"x","lastAllocatedIp":"10.0.0.1","networkIp":"10.0.0.0/24","ownerId":"u","version":1,"peers":null}`,
			wantLen: 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var ng NetworkGroup1
			if err := json.Unmarshal([]byte(tc.payload), &ng); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			if len(ng.Peers) != tc.wantLen {
				t.Fatalf("expected %d peers, got %d", tc.wantLen, len(ng.Peers))
			}
		})
	}
}

// TestPeerUnmarshalUnknownDiscriminator: an unknown "type" value still
// unmarshals (we don't fail) but yields a Peer whose As<Member>() all return
// false. This is more permissive than failing hard — the SDK shouldn't break
// when the API ships a new variant before regen.
func TestPeerUnmarshalUnknownDiscriminator(t *testing.T) {
	payload := `{
		"id": "ng_1", "label": "x", "lastAllocatedIp": "10.0.0.1",
		"networkIp": "10.0.0.0/24", "ownerId": "u", "version": 1,
		"peers": [{"type": "UnknownPeer", "id": "p1"}]
	}`
	var ng NetworkGroup1
	if err := json.Unmarshal([]byte(payload), &ng); err != nil {
		t.Fatalf("unmarshal must not fail on unknown variant, got: %v", err)
	}
	if len(ng.Peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(ng.Peers))
	}
	if ng.Peers[0].Type() != "UnknownPeer" {
		t.Errorf("Type() = %q, want UnknownPeer", ng.Peers[0].Type())
	}
	if _, ok := ng.Peers[0].AsCleverPeer(); ok {
		t.Errorf("AsCleverPeer should be false for unknown variant")
	}
	if _, ok := ng.Peers[0].AsExternalPeer(); ok {
		t.Errorf("AsExternalPeer should be false for unknown variant")
	}
}

// TestPeerRoundTrip checks Marshal→Unmarshal preserves both the discriminator
// and the inner fields, using the To<Union>() method on the concrete type.
func TestPeerRoundTrip(t *testing.T) {
	cleverPeer := CleverPeer{
		ID:           "peer_a",
		Hostname:     "a.example.com",
		Hv:           "hv-a",
		ParentMember: "m1",
		PublicKey:    "pk-a",
	}.ToPeer()

	original := NetworkGroup1{
		ID:              "ng_rt",
		Label:           "rt-label",
		LastAllocatedIP: "10.0.0.1",
		NetworkIP:       "10.0.0.0/24",
		OwnerID:         "owner_rt",
		Version:         3,
		Peers:           []Peer{cleverPeer},
	}

	data, err := json.Marshal(&original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// The serialized payload should contain the discriminator we set.
	if !strings.Contains(string(data), `"type":"CleverPeer"`) {
		t.Fatalf("marshaled payload missing CleverPeer discriminator: %s", data)
	}

	var decoded NetworkGroup1
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(decoded.Peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(decoded.Peers))
	}
	if decoded.Peers[0].Type() != CleverPeerType {
		t.Errorf("decoded.Peers[0].Type() = %q, want %q", decoded.Peers[0].Type(), CleverPeerType)
	}
	rt, ok := decoded.Peers[0].AsCleverPeer()
	if !ok {
		t.Fatalf("AsCleverPeer returned ok=false")
	}
	if rt.ID != "peer_a" || rt.PublicKey != "pk-a" {
		t.Errorf("peer fields not preserved: %+v", rt)
	}
}

// TestPeerFormat verifies fmt.Formatter dispatch:
//   - %v / %+v on a Peer renders the concrete variant (typed Go repr)
//   - flags propagate (%+v shows field names)
//   - unknown variants fall back to the raw JSON bytes
//   - empty Peer renders as "null"
func TestPeerFormat(t *testing.T) {
	t.Run("known variant uses concrete type", func(t *testing.T) {
		cp := CleverPeer{ID: "peer_x", PublicKey: "pk"}.ToPeer()
		got := fmt.Sprintf("%+v", cp)
		// %+v on a struct includes field names; we should see the concrete
		// CleverPeer fields, not byte-slice noise.
		if !strings.Contains(got, "ID:peer_x") {
			t.Errorf("%%+v should show concrete CleverPeer fields, got: %s", got)
		}
		if strings.Contains(got, "raw:[") {
			t.Errorf("raw bytes leaked through: %s", got)
		}
	})

	t.Run("unknown variant falls back to raw JSON", func(t *testing.T) {
		var p Peer
		_ = p.UnmarshalJSON([]byte(`{"type":"FuturePeer","id":"x"}`))
		got := fmt.Sprintf("%v", p)
		if !strings.Contains(got, `"FuturePeer"`) {
			t.Errorf("unknown variant should print raw JSON, got: %s", got)
		}
	})

	t.Run("empty peer prints null", func(t *testing.T) {
		var empty Peer
		if got := fmt.Sprintf("%v", empty); got != "null" {
			t.Errorf("empty peer should print as %q, got %q", "null", got)
		}
	})

	t.Run("nested union prints typed too", func(t *testing.T) {
		cp := CleverPeer{
			ID:       "p1",
			Endpoint: ClientEndpoint{NgIP: "10.0.0.10"}.ToWireguardEndpoint(),
		}.ToPeer()
		got := fmt.Sprintf("%+v", cp)
		// The nested WireguardEndpoint should also dispatch to ClientEndpoint
		// rather than printing as raw bytes.
		if !strings.Contains(got, "NgIP:10.0.0.10") {
			t.Errorf("nested union should dispatch to ClientEndpoint, got: %s", got)
		}
	})
}

// TestPeerDirectMarshal verifies that json.Marshal on a concrete member
// (without going through ToPeer or NewPeerFromX) produces a valid payload
// with the right discriminator.
func TestPeerDirectMarshal(t *testing.T) {
	cp := CleverPeer{ID: "x1", PublicKey: "pk"}
	data, err := json.Marshal(cp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(data), `"type":"CleverPeer"`) {
		t.Errorf("expected type discriminator in payload, got: %s", data)
	}

	// Even when the user has set a wrong type explicitly, MarshalJSON forces
	// the right one — wire format stays consistent.
	cp.Type = "WhateverWrong"
	data2, err := json.Marshal(cp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(data2), `"type":"CleverPeer"`) {
		t.Errorf("MarshalJSON should override wrong Type, got: %s", data2)
	}
}

// TestPeerVariantInterface verifies that concrete members satisfy the
// generated PeerVariant interface so generic helpers can accept any variant.
func TestPeerVariantInterface(t *testing.T) {
	addPeer := func(ng *NetworkGroup1, v PeerVariant) {
		ng.Peers = append(ng.Peers, v.ToPeer())
	}
	var ng NetworkGroup1
	addPeer(&ng, CleverPeer{ID: "c1", PublicKey: "pkc"})
	addPeer(&ng, ExternalPeer{ID: "e1", PublicKey: "pke"})

	if len(ng.Peers) != 2 {
		t.Fatalf("expected 2 peers, got %d", len(ng.Peers))
	}
	if ng.Peers[0].Type() != CleverPeerType {
		t.Errorf("Peers[0].Type() = %q", ng.Peers[0].Type())
	}
	if ng.Peers[1].Type() != ExternalPeerType {
		t.Errorf("Peers[1].Type() = %q", ng.Peers[1].Type())
	}
}

// TestWireguardEndpointDispatch covers a non-slice (single-value) union field.
func TestWireguardEndpointDispatch(t *testing.T) {
	clientPayload := `{
		"type": "CleverPeer",
		"id": "p1", "hostname": "h", "hv": "hv", "parentMember": "m", "publicKey": "pk",
		"endpoint": {"type": "ClientEndpoint", "ngIp": "10.0.0.99"}
	}`
	var p CleverPeer
	if err := json.Unmarshal([]byte(clientPayload), &p); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got := p.Endpoint.Type(); got != ClientEndpointType {
		t.Fatalf("Endpoint.Type() = %q, want %q", got, ClientEndpointType)
	}
	client, ok := p.Endpoint.AsClientEndpoint()
	if !ok {
		t.Fatalf("AsClientEndpoint returned ok=false")
	}
	if client.NgIP != "10.0.0.99" {
		t.Errorf("client.NgIP = %q, want 10.0.0.99", client.NgIP)
	}
}
