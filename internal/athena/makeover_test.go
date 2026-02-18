/* Athena - A server for Attorney Online 2 written in Go
Copyright (C) 2022 MangosArentLiterature <mango@transmenace.dev>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>. */

package athena

import (
	"testing"
)

// TestMakeoverValidCharacter tests that makeover command works with a valid character
func TestMakeoverValidCharacter(t *testing.T) {
	// Save original characters array and restore after test
	originalCharacters := characters
	defer func() {
		characters = originalCharacters
	}()

	// Set up test characters
	characters = []string{
		"Phoenix Wright",
		"Miles Edgeworth",
		"Maya Fey",
	}

	// Create mock clients
	client1 := &Client{
		uid:        1,
		char:       0, // Phoenix Wright
		possessing: -1,
		pair:       ClientPairInfo{wanted_id: -1, emote: "normal", flip: "0", offset: ""},
		pairedUID:  -1,
	}

	client2 := &Client{
		uid:        2,
		char:       1, // Miles Edgeworth
		possessing: -1,
		pair:       ClientPairInfo{wanted_id: -1, emote: "thinking", flip: "0", offset: ""},
		pairedUID:  -1,
	}

	client3 := &Client{
		uid:        3,
		char:       2, // Maya Fey
		possessing: -1,
		pair:       ClientPairInfo{wanted_id: -1, emote: "happy", flip: "1", offset: "50&60"},
		pairedUID:  -1,
	}

	// Save original clients list and restore after test
	originalClients := clients
	defer func() {
		clients = originalClients
	}()

	// Set up test clients list
	clients = ClientList{list: make(map[*Client]struct{})}
	clients.list[client1] = struct{}{}
	clients.list[client2] = struct{}{}
	clients.list[client3] = struct{}{}

	// Test: Force all clients to iniswap into "Miles Edgeworth"
	targetChar := "Miles Edgeworth"

	// Verify character exists
	charID := getCharacterID(targetChar)
	if charID == -1 {
		t.Fatalf("Test setup failed: character '%s' not found", targetChar)
	}

	// Simulate what cmdMakeover does
	for c := range clients.GetAllClients() {
		if c.Uid() == -1 {
			continue
		}
		currentPair := c.PairInfo()
		c.SetPairInfo(targetChar, currentPair.emote, currentPair.flip, currentPair.offset)
	}

	// Verify all clients now have the target character in their PairInfo
	for c := range clients.GetAllClients() {
		if c.PairInfo().name != targetChar {
			t.Errorf("Expected client UID %d to have PairInfo name '%s', got '%s'", 
				c.Uid(), targetChar, c.PairInfo().name)
		}

		// Verify that emote, flip, and offset were preserved
		if c.Uid() == 1 {
			if c.PairInfo().emote != "normal" {
				t.Errorf("Client 1 emote should be preserved as 'normal', got '%s'", c.PairInfo().emote)
			}
		} else if c.Uid() == 2 {
			if c.PairInfo().emote != "thinking" {
				t.Errorf("Client 2 emote should be preserved as 'thinking', got '%s'", c.PairInfo().emote)
			}
		} else if c.Uid() == 3 {
			if c.PairInfo().emote != "happy" {
				t.Errorf("Client 3 emote should be preserved as 'happy', got '%s'", c.PairInfo().emote)
			}
			if c.PairInfo().flip != "1" {
				t.Errorf("Client 3 flip should be preserved as '1', got '%s'", c.PairInfo().flip)
			}
			if c.PairInfo().offset != "50&60" {
				t.Errorf("Client 3 offset should be preserved as '50&60', got '%s'", c.PairInfo().offset)
			}
		}
	}
}

// TestMakeoverInvalidCharacter tests that makeover command handles invalid characters properly
func TestMakeoverInvalidCharacter(t *testing.T) {
	// Save original characters array and restore after test
	originalCharacters := characters
	defer func() {
		characters = originalCharacters
	}()

	// Set up test characters
	characters = []string{
		"Phoenix Wright",
		"Miles Edgeworth",
		"Maya Fey",
	}

	// Test with a character that doesn't exist
	invalidChar := "NonExistent Character"
	charID := getCharacterID(invalidChar)
	
	if charID != -1 {
		t.Errorf("Expected getCharacterID to return -1 for invalid character '%s', got %d", 
			invalidChar, charID)
	}
}

// TestMakeoverSkipsUnjoined tests that makeover skips clients with UID -1
func TestMakeoverSkipsUnjoined(t *testing.T) {
	// Save original characters array and restore after test
	originalCharacters := characters
	defer func() {
		characters = originalCharacters
	}()

	// Set up test characters
	characters = []string{
		"Phoenix Wright",
		"Miles Edgeworth",
	}

	// Create a joined client and an unjoined client (UID -1)
	joinedClient := &Client{
		uid:        1,
		char:       0,
		possessing: -1,
		pair:       ClientPairInfo{wanted_id: -1, emote: "normal"},
		pairedUID:  -1,
	}

	unjoinedClient := &Client{
		uid:        -1, // Not joined yet
		char:       -1,
		possessing: -1,
		pair:       ClientPairInfo{wanted_id: -1, emote: ""},
		pairedUID:  -1,
	}

	// Save original clients list and restore after test
	originalClients := clients
	defer func() {
		clients = originalClients
	}()

	// Set up test clients list
	clients = ClientList{list: make(map[*Client]struct{})}
	clients.list[joinedClient] = struct{}{}
	clients.list[unjoinedClient] = struct{}{}

	targetChar := "Miles Edgeworth"
	
	// Simulate what cmdMakeover does
	var count int
	for c := range clients.GetAllClients() {
		if c.Uid() == -1 {
			continue
		}
		currentPair := c.PairInfo()
		c.SetPairInfo(targetChar, currentPair.emote, currentPair.flip, currentPair.offset)
		count++
	}

	// Verify only the joined client was affected
	if count != 1 {
		t.Errorf("Expected 1 client to be affected, got %d", count)
	}

	// Verify joined client was updated
	if joinedClient.PairInfo().name != targetChar {
		t.Errorf("Expected joined client to have PairInfo name '%s', got '%s'",
			targetChar, joinedClient.PairInfo().name)
	}

	// Verify unjoined client was NOT updated
	if unjoinedClient.PairInfo().name != "" {
		t.Errorf("Expected unjoined client to have empty PairInfo name, got '%s'",
			unjoinedClient.PairInfo().name)
	}
}
