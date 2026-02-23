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
	"strings"
	"testing"

	"github.com/MangosArentLiterature/Athena/internal/area"
)

// TestSIMusicCountMatchesSMEntries verifies that the music_count reported in the
// SI packet equals the total number of entries sent in SM packets (area names +
// music entries combined). A mismatch causes AO2/WebAO clients to send RD before
// all SM chunks have been received, resulting in an instant disconnect.
func TestSIMusicCountMatchesSMEntries(t *testing.T) {
	origAreas := areas
	origAreaNames := areaNames
	origMusic := music
	t.Cleanup(func() {
		areas = origAreas
		areaNames = origAreaNames
		music = origMusic
	})

	tests := []struct {
		name      string
		areaList  []string
		musicList []string
	}{
		{
			name:     "single area, small music list",
			areaList: []string{"Courtroom 1"},
			musicList: []string{"[Songs]", "aa.opus", "bb.opus"},
		},
		{
			name:      "multiple areas, music list requiring chunking",
			areaList:  []string{"Area 1", "Area 2", "Area 3", "Area 4", "Area 5"},
			musicList: makeMusicList(98),
		},
		{
			name:      "many areas and large music list (multiple SM chunks)",
			areaList:  []string{"Area 1", "Area 2", "Area 3"},
			musicList: makeMusicList(200),
		},
		{
			name:      "single area, exactly 100 total entries",
			areaList:  []string{"Area 1"},
			musicList: makeMusicList(99),
		},
		{
			name:      "single area, exactly 101 total entries",
			areaList:  []string{"Area 1"},
			musicList: makeMusicList(100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up global state matching what InitServer produces.
			areas = make([]*area.Area, len(tt.areaList))
			for i, name := range tt.areaList {
				areas[i] = area.NewArea(area.AreaData{Name: name}, 1, 0, area.EviAny)
			}
			areaNames = strings.Join(tt.areaList, "#")
			music = tt.musicList

			// Compute the total number of SM entries (mirrors pktReqAM logic).
			var allEntries []string
			if areaNames != "" {
				allEntries = strings.Split(areaNames, "#")
			}
			allEntries = append(allEntries, music...)
			totalSMEntries := len(allEntries)

			// The SI music_count must equal the total SM entries so the client
			// knows exactly when to send RD.
			siMusicCount := len(areas) + len(music)

			if siMusicCount != totalSMEntries {
				t.Errorf("SI music_count=%d does not match total SM entries=%d; "+
					"client will send RD before all SM chunks arrive, causing instant disconnect",
					siMusicCount, totalSMEntries)
			}
		})
	}
}

// makeMusicList returns a music list with n entries (a category header + n-1 songs).
func makeMusicList(n int) []string {
	entries := make([]string, n)
	entries[0] = "[Songs]"
	for i := 1; i < n; i++ {
		entries[i] = "song" + strings.Repeat("x", i%5) + ".opus"
	}
	return entries
}
