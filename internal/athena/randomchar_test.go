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

	"github.com/MangosArentLiterature/Athena/internal/area"
)

// TestGetRandomFreeChar verifies that getRandomFreeChar returns a free character
// ID from the client's area, matching the behaviour expected when WebAO sends
// CC#0#-1#% (random character button).
func TestGetRandomFreeChar(t *testing.T) {
	origChars := characters
	t.Cleanup(func() { characters = origChars })

	characters = []string{"Phoenix Wright", "Miles Edgeworth", "Maya Fey", "Franziska von Karma"}

	t.Run("returns free character when some are taken", func(t *testing.T) {
		a := area.NewArea(area.AreaData{}, len(characters), 0, area.EviAny)
		// Take characters 0 and 2.
		a.AddChar(0)
		a.AddChar(2)

		client := &Client{
			uid:        1,
			char:       -1,
			possessing: -1,
			pair:       ClientPairInfo{wanted_id: -1},
		}
		client.SetArea(a)

		id := getRandomFreeChar(client)
		if id != 1 && id != 3 {
			t.Errorf("getRandomFreeChar returned %d, want 1 or 3 (free characters)", id)
		}
	})

	t.Run("returns -1 when all characters are taken", func(t *testing.T) {
		a := area.NewArea(area.AreaData{}, len(characters), 0, area.EviAny)
		// Take all characters.
		for i := range characters {
			a.AddChar(i)
		}

		client := &Client{
			uid:        1,
			char:       -1,
			possessing: -1,
			pair:       ClientPairInfo{wanted_id: -1},
		}
		client.SetArea(a)

		id := getRandomFreeChar(client)
		if id != -1 {
			t.Errorf("getRandomFreeChar returned %d, want -1 (no free characters)", id)
		}
	})

	t.Run("returns the only free character when one is available", func(t *testing.T) {
		a := area.NewArea(area.AreaData{}, len(characters), 0, area.EviAny)
		// Take all except character 2.
		a.AddChar(0)
		a.AddChar(1)
		a.AddChar(3)

		client := &Client{
			uid:        1,
			char:       -1,
			possessing: -1,
			pair:       ClientPairInfo{wanted_id: -1},
		}
		client.SetArea(a)

		id := getRandomFreeChar(client)
		if id != 2 {
			t.Errorf("getRandomFreeChar returned %d, want 2 (only free character)", id)
		}
	})

	t.Run("returns -1 when character list is empty", func(t *testing.T) {
		origCharsInner := characters
		t.Cleanup(func() { characters = origCharsInner })
		characters = []string{}

		a := area.NewArea(area.AreaData{}, 0, 0, area.EviAny)
		client := &Client{
			uid:        1,
			char:       -1,
			possessing: -1,
			pair:       ClientPairInfo{wanted_id: -1},
		}
		client.SetArea(a)

		id := getRandomFreeChar(client)
		if id != -1 {
			t.Errorf("getRandomFreeChar returned %d, want -1 (empty character list)", id)
		}
	})
}
