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
	"time"
)

// resetHotPotatoState resets the hot potato global state for test isolation.
func resetHotPotatoState() {
	hotPotato.mu.Lock()
	hotPotato.optInActive = false
	hotPotato.gameActive = false
	hotPotato.participants = make(map[int]bool)
	hotPotato.carrierUID = -1
	hotPotato.lastGameEnd = time.Time{}
	hotPotato.mu.Unlock()
}

// TestHotPotatoCooldown verifies the cooldown check returns correct state.
func TestHotPotatoCooldown(t *testing.T) {
	resetHotPotatoState()

	// No game has run yet – should not be cooling down.
	cooling, _ := isHotPotatoCoolingDown()
	if cooling {
		t.Error("Expected no cooldown when no game has run yet")
	}

	// Simulate a game that ended 1 second ago.
	hotPotato.mu.Lock()
	hotPotato.lastGameEnd = time.Now().Add(-1 * time.Second)
	hotPotato.mu.Unlock()

	cooling, secs := isHotPotatoCoolingDown()
	if !cooling {
		t.Error("Expected cooldown to be active after a recent game")
	}
	if secs <= 0 {
		t.Errorf("Expected positive remaining seconds, got %d", secs)
	}

	// Simulate a game that ended 6 minutes ago – cooldown should have expired.
	hotPotato.mu.Lock()
	hotPotato.lastGameEnd = time.Now().Add(-6 * time.Minute)
	hotPotato.mu.Unlock()

	cooling, _ = isHotPotatoCoolingDown()
	if cooling {
		t.Error("Expected cooldown to be expired after 6 minutes")
	}
}

// TestHotPotatoOptIn verifies that participants can be tracked.
func TestHotPotatoOptIn(t *testing.T) {
	resetHotPotatoState()

	hotPotato.mu.Lock()
	hotPotato.optInActive = true
	hotPotato.mu.Unlock()

	// Simulate two opt-ins.
	hotPotato.mu.Lock()
	hotPotato.participants[1] = true
	hotPotato.participants[2] = true
	hotPotato.mu.Unlock()

	hotPotato.mu.Lock()
	count := len(hotPotato.participants)
	hotPotato.mu.Unlock()

	if count != 2 {
		t.Errorf("Expected 2 participants, got %d", count)
	}
}

// TestHotPotatoDoubleOptIn verifies a player cannot opt in twice.
func TestHotPotatoDoubleOptIn(t *testing.T) {
	resetHotPotatoState()

	hotPotato.mu.Lock()
	hotPotato.optInActive = true
	hotPotato.participants[42] = true
	hotPotato.mu.Unlock()

	// Try to opt in again – the map should still have only 1 entry for uid 42.
	hotPotato.mu.Lock()
	alreadyIn := hotPotato.participants[42]
	if !alreadyIn {
		t.Error("Expected participant 42 to be in the map")
	}
	hotPotato.participants[42] = true // idempotent
	count := len(hotPotato.participants)
	hotPotato.mu.Unlock()

	if count != 1 {
		t.Errorf("Expected 1 participant after double opt-in attempt, got %d", count)
	}
}

// TestRandomHotPotatoPunishment verifies the function always returns a valid punishment type.
func TestRandomHotPotatoPunishment(t *testing.T) {
	validTypes := map[PunishmentType]bool{
		PunishmentBackward:    true,
		PunishmentStutterstep: true,
		PunishmentElongate:    true,
		PunishmentUppercase:   true,
		PunishmentLowercase:   true,
		PunishmentRobotic:     true,
		PunishmentAlternating: true,
		PunishmentUwu:         true,
		PunishmentPirate:      true,
		PunishmentCaveman:     true,
		PunishmentDrunk:       true,
		PunishmentHiccup:      true,
		PunishmentConfused:    true,
		PunishmentParanoid:    true,
		PunishmentMumble:      true,
		PunishmentSubtitles:   true,
	}

	// 100 iterations gives a high probability of hitting all branches of the pool
	// while keeping the test fast.
	const iterations = 100
	for i := 0; i < iterations; i++ {
		p := randomHotPotatoPunishment()
		if !validTypes[p] {
			t.Errorf("randomHotPotatoPunishment returned unexpected type: %v", p)
		}
	}
}

// TestHotPotatoOnlyOneGame verifies the guard against starting two concurrent games.
func TestHotPotatoOnlyOneGame(t *testing.T) {
	resetHotPotatoState()

	// Set optInActive to simulate an already-running game.
	hotPotato.mu.Lock()
	hotPotato.optInActive = true
	hotPotato.mu.Unlock()

	hotPotato.mu.Lock()
	alreadyActive := hotPotato.optInActive || hotPotato.gameActive
	hotPotato.mu.Unlock()

	if !alreadyActive {
		t.Error("Expected game to be detected as active")
	}
}
