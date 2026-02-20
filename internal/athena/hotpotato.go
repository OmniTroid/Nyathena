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
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	hotPotatoOptInDuration      = 60 * time.Second
	hotPotatoGameDuration       = 5 * time.Minute
	hotPotatoCooldown           = 5 * time.Minute
	hotPotatoMinParticipants    = 2
	hotPotatoPunishmentDuration = 10 * time.Minute
)

// hotPotatoRules is the explanation shown when a game is announced.
const hotPotatoRules = `🥔 HOT POTATO EVENT STARTING! 🥔
Type /hotpotato accept within 60 seconds to join.

📋 HOW TO PLAY:
• One random participant will secretly be given the "Hot Potato".
• The Hot Potato carrier has a 5-minute virtual timer.
• AVOID being in the same area as the carrier when the timer runs out!
• At the end of 5 minutes, everyone opted-in who is in the same area as the carrier receives a random punishment.
• If the carrier is a MODERATOR, all participants in the same area will be KICKED from the server instead.
• If the carrier fails to share their area with anyone, THEY receive the punishment themselves.
• Players who did not opt in are completely safe and unaffected.
• Only one Hot Potato game can run at a time (5-minute cooldown between games).

Good luck — and watch who you hang around with! 🔥`

// hotPotatoState holds all state for an active or pending hot potato game.
type hotPotatoState struct {
	mu           sync.Mutex
	optInActive  bool           // true during the 60-second opt-in window
	gameActive   bool           // true while the 5-minute game is running
	participants map[int]bool   // uid -> opted-in
	carrierUID   int            // uid of hot potato carrier (-1 if none)
	lastGameEnd  time.Time      // time the last game ended (for cooldown)
}

var hotPotato = hotPotatoState{
	participants: make(map[int]bool),
	carrierUID:   -1,
}

// isHotPotatoCoolingDown returns true and remaining seconds if the global cooldown is active.
func isHotPotatoCoolingDown() (bool, int) {
	hotPotato.mu.Lock()
	defer hotPotato.mu.Unlock()
	if hotPotato.lastGameEnd.IsZero() {
		return false, 0
	}
	elapsed := time.Since(hotPotato.lastGameEnd)
	if elapsed < hotPotatoCooldown {
		remaining := int((hotPotatoCooldown - elapsed).Seconds()) + 1
		return true, remaining
	}
	return false, 0
}

// cmdHotPotato handles both /hotpotato (start) and /hotpotato accept.
func cmdHotPotato(client *Client, args []string, _ string) {
	// Sub-command: /hotpotato accept
	if len(args) > 0 && args[0] == "accept" {
		hotPotatoAccept(client)
		return
	}

	// Start a new game announcement
	hotPotato.mu.Lock()

	if hotPotato.optInActive || hotPotato.gameActive {
		hotPotato.mu.Unlock()
		client.SendServerMessage("A Hot Potato game is already in progress.")
		return
	}

	if !hotPotato.lastGameEnd.IsZero() {
		elapsed := time.Since(hotPotato.lastGameEnd)
		if elapsed < hotPotatoCooldown {
			remaining := int((hotPotatoCooldown - elapsed).Seconds()) + 1
			client.SendServerMessage(fmt.Sprintf("Hot Potato is on cooldown. Please wait %d seconds.", remaining))
			hotPotato.mu.Unlock()
			return
		}
	}

	// Begin opt-in phase
	hotPotato.optInActive = true
	hotPotato.gameActive = false
	hotPotato.participants = make(map[int]bool)
	hotPotato.carrierUID = -1
	hotPotato.mu.Unlock()

	// Announce globally in OOC
	writeToAll("CT", encode(config.Name), encode(hotPotatoRules), "1")
	addToBuffer(client, "CMD", "Started Hot Potato opt-in", false)

	// Launch the opt-in timer
	go hotPotatoOptInTimer()
}

// hotPotatoAccept handles a player opting into the current game.
func hotPotatoAccept(client *Client) {
	hotPotato.mu.Lock()
	defer hotPotato.mu.Unlock()

	if !hotPotato.optInActive {
		client.SendServerMessage("There is no active Hot Potato game to join right now.")
		return
	}

	uid := client.Uid()
	if hotPotato.participants[uid] {
		client.SendServerMessage("You have already joined the Hot Potato game.")
		return
	}

	hotPotato.participants[uid] = true
	client.SendServerMessage(fmt.Sprintf("🥔 You have joined the Hot Potato game! (%d participant(s) so far)", len(hotPotato.participants)))
	writeToAll("CT", encode(config.Name), encode(fmt.Sprintf("🥔 %v joined Hot Potato! (%d participant(s))", client.OOCName(), len(hotPotato.participants))), "1")
}

// hotPotatoOptInTimer waits for the opt-in window, then starts the game if enough players joined.
func hotPotatoOptInTimer() {
	time.Sleep(hotPotatoOptInDuration)

	hotPotato.mu.Lock()

	if !hotPotato.optInActive {
		// Game was cancelled externally
		hotPotato.mu.Unlock()
		return
	}

	hotPotato.optInActive = false

	// Collect participant UIDs that are still connected
	var validUIDs []int
	for uid := range hotPotato.participants {
		if _, err := getClientByUid(uid); err == nil {
			validUIDs = append(validUIDs, uid)
		}
	}

	if len(validUIDs) < hotPotatoMinParticipants {
		hotPotato.lastGameEnd = time.Now().UTC()
		hotPotato.mu.Unlock()
		writeToAll("CT", encode(config.Name), encode(fmt.Sprintf("🥔 Hot Potato cancelled: not enough participants (%d/%d required).", len(validUIDs), hotPotatoMinParticipants)), "1")
		return
	}

	// Pick a random carrier
	carrierUID := validUIDs[rand.Intn(len(validUIDs))]
	hotPotato.carrierUID = carrierUID
	hotPotato.gameActive = true
	hotPotato.mu.Unlock()

	// Notify all participants; carrier gets a secret DM
	writeToAll("CT", encode(config.Name), encode(fmt.Sprintf("🔥 THE HOT POTATO GAME HAS BEGUN! %d players are in. One of them is carrying the Hot Potato... The timer has started! Avoid anyone suspicious for the next 5 minutes!", len(validUIDs))), "1")

	carrierClient, err := getClientByUid(carrierUID)
	if err == nil {
		carrierClient.SendServerMessage("🥔🔥 YOU have the Hot Potato! Try to be in the same area as other participants when the timer expires. You have 5 minutes!")
	}

	// Launch the game timer
	go hotPotatoGameTimer(carrierUID)
}

// hotPotatoGameTimer waits for the game duration, then resolves the game.
func hotPotatoGameTimer(carrierUID int) {
	time.Sleep(hotPotatoGameDuration)

	hotPotato.mu.Lock()

	if !hotPotato.gameActive {
		// Game ended externally
		hotPotato.mu.Unlock()
		return
	}

	hotPotato.gameActive = false
	hotPotato.optInActive = false
	hotPotato.lastGameEnd = time.Now().UTC()

	// Snapshot participants before releasing lock
	participantsCopy := make(map[int]bool)
	for uid, v := range hotPotato.participants {
		participantsCopy[uid] = v
	}
	hotPotato.mu.Unlock()

	carrierClient, carrierErr := getClientByUid(carrierUID)

	// Resolve game: find opted-in players who share the same area as the carrier
	var affected []*Client
	if carrierErr == nil {
		carrierArea := carrierClient.Area()
		for uid := range participantsCopy {
			if uid == carrierUID {
				continue
			}
			c, err := getClientByUid(uid)
			if err != nil {
				continue
			}
			if c.Area() == carrierArea {
				affected = append(affected, c)
			}
		}
	}

	if len(affected) == 0 {
		// Carrier failed to affect anyone – carrier gets punished
		writeToAll("CT", encode(config.Name), encode("⏰ HOT POTATO TIMER EXPIRED! The carrier was alone — they get punished! 🥔💀"), "1")
		if carrierErr == nil {
			pType := randomHotPotatoPunishment()
			carrierClient.AddPunishment(pType, hotPotatoPunishmentDuration, "Hot Potato: solo carrier penalty")
			carrierClient.SendServerMessage(fmt.Sprintf("💀 You had the Hot Potato and nobody was nearby — you've been punished with '%v'!", pType.String()))
			addToBuffer(carrierClient, "HOTPOTATO", fmt.Sprintf("Carrier self-punished with %v (no victims)", pType.String()), false)
		}
		return
	}

	// Carrier is a mod → kick everyone in the same area
	isMod := carrierErr == nil && carrierClient.Authenticated()

	if isMod {
		writeToAll("CT", encode(config.Name), encode(fmt.Sprintf("⏰ HOT POTATO TIMER EXPIRED! The carrier was a MODERATOR — %d participant(s) in the same area are being KICKED! 🔨", len(affected))), "1")
		for _, c := range affected {
			c.SendPacket("KK", "Hot Potato: you were caught in the same area as a moderator carrying the Hot Potato!")
			c.conn.Close()
		}
		if carrierErr == nil {
			addToBuffer(carrierClient, "HOTPOTATO", fmt.Sprintf("Mod carrier kicked %d participant(s)", len(affected)), false)
		}
	} else {
		// Normal carrier → random punishment for everyone in same area
		writeToAll("CT", encode(config.Name), encode(fmt.Sprintf("⏰ HOT POTATO TIMER EXPIRED! %d participant(s) were caught in the same area as the carrier and received random punishments! 🥔💥", len(affected))), "1")
		for _, c := range affected {
			pType := randomHotPotatoPunishment()
			c.AddPunishment(pType, hotPotatoPunishmentDuration, "Hot Potato punishment")
			c.SendServerMessage(fmt.Sprintf("💥 You were caught with the Hot Potato carrier! You've been punished with '%v' for 10 minutes.", pType.String()))
			if carrierErr == nil {
				addToBuffer(carrierClient, "HOTPOTATO", fmt.Sprintf("Punished UID %d with %v", c.Uid(), pType.String()), false)
			}
		}
	}
}

// randomHotPotatoPunishment returns a random punishment type suitable for Hot Potato.
func randomHotPotatoPunishment() PunishmentType {
	pool := []PunishmentType{
		PunishmentBackward,
		PunishmentStutterstep,
		PunishmentElongate,
		PunishmentUppercase,
		PunishmentLowercase,
		PunishmentRobotic,
		PunishmentAlternating,
		PunishmentUwu,
		PunishmentPirate,
		PunishmentCaveman,
		PunishmentDrunk,
		PunishmentHiccup,
		PunishmentConfused,
		PunishmentParanoid,
		PunishmentMumble,
		PunishmentSubtitles,
	}
	return pool[rand.Intn(len(pool))]
}
