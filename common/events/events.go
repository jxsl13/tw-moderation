package events

// this package defines all of the existing events as well as their respective structs

const (

	// FROM SERVERS

	// HEARTBEAT to find new services(?)
	TypeDiscoveryHeartBeat = "HEARTBEAT"

	// TypePlayerJoin is an event that is created by a monitor when a player joins a server
	TypePlayerJoin = "event:PLAYER_JOIN"
	// TypePlayerLeave is created when a player leaves the server
	TypePlayerLeave = "event:PLAYER_LEAVE"
	// TypeKickvoteStart is created when a kickvote is started
	TypeKickvoteStart = "event:KICKVOTE_START"
	// TypeSpecvoteStart is created when a specvote is started,
	// trying to move a player to the spectators
	TypeSpecvoteStart = "event:SPECVOTE_START"
	// TypeChatMessage is created when some chat message is written by someone
	TypeChatMessage = "event:CHAT_MESSAGE"
	// TypeWhisperMessage is created when someone whispers to someone else
	TypeWhisperMessage = "event:WHISPER_MESSAGE"
	// TypeTeamchatMessage is created when someone writes in the teamchat
	TypeTeamchatMessage = "event:TEAMCHAT_MESSAGE"

	// TO SERVERS
	// TypeExecCommand is used to forward command execution requests to the Teeworlds servers
	TypeExecCommand = "event:EXEC_COMMAND"
)
