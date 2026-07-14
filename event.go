package alohssh

import "github.com/kiryuhakipyatok/aloh-ssh/internal/domain/models"

type (
	Event                = models.Event
	FriendConnsData      = models.FriendConnsData
	UsersHardDenoiseData = models.UsersHardDenoiseData
	UsersSoftDenoiseData = models.UsersSoftDenoiseData
	TaglineData          = models.TaglineData
	Identity             = models.Identity
)

const (
	NEW_FRIEND_REQ      = models.NEW_FRIEND_REQ
	ACCEPT_FRIEND       = models.ACCEPT_FRIEND
	DENY_FRIEND         = models.DENY_FRIEND
	DELETE_FRIEND       = models.DELETE_FRIEND
	BLOCK_USER          = models.BLOCK_USER
	UNBLOCK_USER        = models.UNBLOCK_USER
	FRIEND_ONLINE       = models.FRIEND_ONLINE
	FRIEND_OFFLINE      = models.FRIEND_OFFLINE
	FRIEND_CONNECTIONS  = models.FRIEND_CONNECTIONS
	UPDATE_HARD_DENOISE = models.UPDATE_HARD_DENOISE
	UPDATE_SOFT_DENOISE = models.UPDATE_SOFT_DENOISE
	UPDATE_TAGLINE      = models.UPDATE_TAGLINE
)
