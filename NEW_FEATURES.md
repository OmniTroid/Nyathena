# New Features Documentation

## Feature 1: Stacking Punishments

### Overview
Multiple different punishment types can now be applied to a single user simultaneously. The effects combine sequentially, creating interesting and varied punishment experiences.

### Usage
**Individual Commands**: Apply punishments one at a time using existing commands (e.g., `/uppercase`, `/backward`, `/uwu`)

**Stack Command**: Apply multiple punishments at once:
```
/stack <punishment1> <punishment2> [<punishment3>...] [-d duration] [-r reason] <uid1>,<uid2>...
```

### Examples
```
/stack uppercase backward -d 15m -r "Being silly" 5,7,9
/stack uwu pirate robotic -d 1h 12
```

### Notes
- Different punishment types stack and apply sequentially
- Adding the same punishment type twice will replace the first one
- Punishments are applied in order of addition for consistent results

---

## Feature 2: Punishment Tournament Mode

### Overview
A voluntary game mode where users compete with random punishment effects. Participants try to communicate effectively while under punishment effects. The user with the most messages wins!

### Commands

**Start Tournament** (Requires MUTE permission):
```
/tournament start
```

**Join Tournament** (Any user):
```
/join-tournament
```
- Automatically applies 2-3 random punishments to the participant
- Punishments have no expiration during the tournament

**View Status** (Requires MUTE permission):
```
/tournament status
```
Shows:
- Tournament duration
- Number of participants
- Leaderboard sorted by message count
- Time each participant has been in the tournament

**Stop Tournament** (Requires MUTE permission):
```
/tournament stop
```
- Announces winner (highest message count)
- Removes all punishments from winner
- Clears tournament state

### Participation Flow
1. Admin starts tournament with `/tournament start`
2. Users voluntarily join with `/join-tournament`
3. Random punishments are applied (2-3 per user)
4. Users send IC messages (counted automatically)
5. Admin checks leaderboard with `/tournament status`
6. Admin ends tournament with `/tournament stop`
7. Winner announced to all users

### Notes
- Tournament mode is server-wide
- Only one tournament can be active at a time
- Message counts are tracked automatically in IC messages
- Random punishments include: backward, stutterstep, elongate, uppercase, lowercase, robotic, alternating, uwu, pirate, confused, drunk, hiccup

---

## Technical Implementation Details

### Thread Safety
- Tournament state uses mutex locks for thread safety
- Safe for concurrent access by multiple clients

### State Management
- Tournament state stored in `tournamentParticipants` map
- Initialized in `InitServer()`

### Message Processing
- Tournament message counting happens in `pktIC` during IC message processing

### Testing
All features include comprehensive tests:
- Stacking punishment tests
- Punishment replacement tests
- Type conversion tests
- Sequential punishment application tests
- Tournament participant creation tests

---

## Feature 3: Coinflip Challenge System

### Overview
A player-vs-player coinflip challenge system that allows any user to challenge another player to a coinflip battle. This replaces the removed `/copycats` and `/randomname` punishment commands with a more interactive social feature.

### Command
```
/coinflip <heads|tails>
```

### How It Works

**Starting a Challenge:**
1. Player1 types `/coinflip heads` or `/coinflip tails`
2. System announces: "Player1 has chosen heads and is ready to coinflip! Type /coinflip tails to battle them!"
3. Challenge remains active for 30 seconds

**Accepting a Challenge:**
1. Player2 types `/coinflip` with the opposite choice (if Player1 chose heads, Player2 must choose tails)
2. System randomly flips a virtual coin (50/50 chance)
3. Winner is announced to the area: "⚔️ COINFLIP BATTLE! Player1 (heads) vs Player2 (tails) - The coin landed on heads! 🎉 Player1 WINS! 🎉"
4. Challenge is cleared

### Examples

**Basic Usage:**
```
Player1: /coinflip heads
Server: Player1 has chosen heads and is ready to coinflip! Type /coinflip tails to battle them!
Player2: /coinflip tails
Server: ⚔️ COINFLIP BATTLE! Player1 (heads) vs Player2 (tails) - The coin landed on tails! 🎉 Player2 WINS! 🎉
```

**Invalid Choice:**
```
Player1: /coinflip coin
Server: Invalid choice. Use: heads or tails.
```

**Same Player Attempting to Accept Own Challenge:**
```
Player1: /coinflip heads
Server: Player1 has chosen heads and is ready to coinflip! Type /coinflip tails to battle them!
Player1: /coinflip tails
Server: You cannot accept your own coinflip challenge!
```

**Wrong Choice (Must Pick Opposite):**
```
Player1: /coinflip heads
Server: Player1 has chosen heads and is ready to coinflip! Type /coinflip tails to battle them!
Player2: /coinflip heads
Server: You must pick the opposite choice! The challenger picked heads, so you must pick tails.
```

**Challenge Expiration:**
```
Player1: /coinflip heads
Server: Player1 has chosen heads and is ready to coinflip! Type /coinflip tails to battle them!
[... 31 seconds pass ...]
Player2: /coinflip tails
Server: Previous coinflip expired. Player2 has chosen tails and is ready to coinflip! Type /coinflip heads to battle them!
```

### Notes
- Available to all users (no special permissions required)
- Challenges are area-specific (one active challenge per area)
- Challenges expire after 30 seconds if not accepted
- Players cannot accept their own challenges
- Must choose the opposite side from the challenger
- Results are logged in the game buffer for both players

### Removed Commands
This feature replaces:
- `/copycats` - Moderator-only punishment command that modified messages
- `/randomname` - Moderator-only punishment command that changed names

### Testing
Comprehensive tests include:
- `oppositeChoice` helper function validation
- Winner determination logic (4 scenarios)
- Choice validation tests
- Edge case handling

---

## Feature 4: Configurable Rate Limiting

### Overview
A lightweight, resource-efficient spam prevention system that automatically kicks users who exceed a configurable message rate limit. This prevents server abuse while allowing legitimate users to communicate normally.

### Configuration
Add these settings to your `config.toml` under the `[Server]` section:

```toml
# Rate limiting: Maximum number of messages (IC, OOC, music) a player can send within the time window.
# Players who exceed this limit will be automatically kicked from the server.
# This helps prevent spam and resource exhaustion. Set to 0 to disable rate limiting.
# Default: 20 messages
message_rate_limit = 20

# Rate limiting: Time window in seconds for counting messages.
# For example, with message_rate_limit=20 and message_rate_limit_window=10,
# players can send up to 20 messages every 10 seconds.
# Default: 10 seconds
message_rate_limit_window = 10
```

### How It Works

**Sliding Window Approach:**
- Tracks timestamps of recent messages per client
- Automatically cleans up old timestamps outside the window
- Memory-efficient: releases unused memory through GC

**Applies to:**
- IC (In-Character) messages
- OOC (Out-of-Character) messages
- Music changes

**When Triggered:**
- User receives message: "You have been kicked for spamming."
- Kick is logged: `Client (IPID:xxx UID:x) kicked for exceeding rate limit`
- Connection is closed

### Examples

**Default Configuration (Recommended):**
```toml
message_rate_limit = 20
message_rate_limit_window = 10
```
Allows 20 messages per 10 seconds (average 2 messages/second)

**Strict Configuration:**
```toml
message_rate_limit = 10
message_rate_limit_window = 5
```
Allows 10 messages per 5 seconds (average 2 messages/second)

**Lenient Configuration:**
```toml
message_rate_limit = 50
message_rate_limit_window = 30
```
Allows 50 messages per 30 seconds (average 1.67 messages/second)

**Disabled:**
```toml
message_rate_limit = 0
message_rate_limit_window = 10
```
Rate limiting completely disabled

### Technical Details

**Resource Efficiency:**
- Uses simple slice of timestamps (not a complex data structure)
- Automatic cleanup of expired timestamps
- Thread-safe with mutex protection
- Minimal CPU overhead per message

**Memory Management:**
- Expired timestamps are removed from memory
- Empty slices release underlying arrays to GC
- Memory usage scales with active users, not message volume

**Previous Implementation:**
- Removed hardcoded 10 requests/second global rate limiter
- New implementation is per-client and configurable

### Testing
Comprehensive test coverage includes:
- Disabled rate limiting functionality
- Basic rate limiting enforcement
- Sliding window behavior verification
- Concurrent access safety (race condition free)
- Memory efficiency and cleanup verification
- All tests pass with race detector

### Security
- CodeQL scan: No vulnerabilities detected
- No race conditions detected
- Thread-safe implementation

### Notes
- Default values (20 messages per 10 seconds) are suitable for most servers
- Legitimate rapid-fire conversations are still possible within limits
- Spammers attempting flood attacks are automatically kicked
- Setting to 0 disables rate limiting entirely (not recommended for public servers)
- Changes require server restart to take effect

---

## Feature 5: Per-Area Logging

### Overview
A comprehensive logging system that creates separate log folders for each area with daily-rotating log files. This allows server administrators to track all activity in specific areas for moderation, investigation, and record-keeping purposes.

### Configuration
Add this setting to your `config.toml` under the `[Logging]` section:

```toml
# Enable per-area folder logging with daily rotation.
# When enabled, each area gets its own folder under logs/ with daily log files.
# Format: logs/AreaName/AreaName-YYYY-MM-DD.txt
# Each log line includes: [HH:MM:SS] | ACTION | CHARACTER | IPID | HDID | SHOWNAME | OOC_NAME | MESSAGE
enable_area_logging = false
```

### Directory Structure
When enabled, the logging system creates the following structure:

```
logs/
├── Lobby/
│   ├── Lobby-2026-02-18.txt
│   ├── Lobby-2026-02-17.txt
│   └── Lobby-2026-02-16.txt
├── Courtroom/
│   ├── Courtroom-2026-02-18.txt
│   └── Courtroom-2026-02-17.txt
└── Defense Attorney/
    └── Defense Attorney-2026-02-18.txt
```

### Log Format
Each log entry includes comprehensive information about the action:

```
[HH:MM:SS] | ACTION | CHARACTER | IPID | HDID | SHOWNAME | OOC_NAME | MESSAGE
```

**Example Entries:**
```
[15:04:05] | IC | Phoenix Wright | abc123def | hash789 | Phoenix | JohnDoe | "Objection!"
[15:04:06] | OOC | Spectator | xyz456abc | hash456 | Spectator | JaneDoe | "nice moves"
[15:04:07] | AREA | Phoenix Wright | abc123def | hash789 | Phoenix | JohnDoe | "Joined area."
[15:04:10] | MUSIC | Maya Fey | def789ghi | hash123 | Maya | Player3 | "Changed music to Trial.mp3."
[15:04:15] | CMD | Miles Edgeworth | ghi012jkl | hash321 | Edgeworth | Moderator1 | "Set BG to courtroom."
```

### Action Types Logged
- **IC** - In-character messages
- **OOC** - Out-of-character messages
- **AREA** - Area join/leave events
- **MUSIC** - Music changes
- **CMD** - Moderator commands (background changes, area settings, etc.)
- **AUTH** - Authentication events
- **MOD** - Moderator actions
- **JUD** - Judge actions (testimony, verdicts)
- **EVI** - Evidence changes

### Features

**Daily File Rotation:**
- New log file created each day automatically
- Files named with ISO date format (YYYY-MM-DD)
- No manual intervention required

**Thread-Safe Operation:**
- Per-area mutex locks prevent race conditions
- Multiple users in the same area write safely
- No performance impact under normal load

**Special Character Handling:**
- Area names with special characters are sanitized
- Slashes, colons, and other problematic characters replaced with underscores
- Examples: "Area/Test" → "Area_Test", "Room:1" → "Room_1"

**Zero Performance Impact When Disabled:**
- No overhead when `enable_area_logging = false`
- Files only created when feature is enabled

**Automatic Directory Creation:**
- Area log directories created on server startup
- New areas automatically get their directories
- No manual filesystem setup required

### Usage Examples

**Enable Area Logging:**
```toml
[Logging]
enable_area_logging = true
log_directory = "logs"
```

**Custom Log Directory:**
```toml
[Logging]
enable_area_logging = true
log_directory = "/var/log/athena"
```

**View Logs:**
```bash
# View today's logs for Courtroom
cat logs/Courtroom/Courtroom-2026-02-18.txt

# Search for specific user activity
grep "abc123def" logs/Courtroom/*.txt

# View all IC messages in an area
grep "| IC |" logs/Lobby/Lobby-2026-02-18.txt

# Find all moderator commands
grep "| CMD |" logs/*/$(date +%Y-%m-%d).txt
```

### Use Cases

**Moderation:**
- Review reported incidents
- Track problematic user behavior
- Provide evidence for ban appeals

**Investigation:**
- Trace user activity across sessions
- Identify patterns of rule violations
- Correlate events between areas

**Record Keeping:**
- Archive important roleplay sessions
- Maintain server history
- Comply with community guidelines

### Technical Details

**Implementation:**
- Uses `sync.Map` for per-area locks (memory efficient)
- Files opened in append mode with proper permissions (0644)
- Uses `filepath.Join()` for cross-platform compatibility
- Thread-safe with individual area locks (no global lock bottleneck)

**Performance:**
- O(1) lock acquisition per area
- No blocking between different areas
- Minimal memory footprint
- Efficient file I/O with append mode

**Cross-Platform:**
- Works on Linux, Windows, and macOS
- Proper path separators for each OS
- Safe filename generation

### Testing
Comprehensive test coverage includes:
- Area name sanitization (special characters)
- Directory creation
- Log file writing (single and multiple entries)
- Disabled logging behavior
- Thread safety
- Cross-platform path handling

### Security
- CodeQL scan: No vulnerabilities detected
- No path traversal vulnerabilities
- Safe filename sanitization
- Proper file permissions (0644 for files, 0755 for directories)

### Notes
- Log files can grow large over time - consider implementing log rotation or cleanup
- IPID and HDID are hashed for privacy
- Changes require server restart to take effect
- Logs are UTF-8 encoded for international character support
- Disk space should be monitored when enabled on high-traffic servers
