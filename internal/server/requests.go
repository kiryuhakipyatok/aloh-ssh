package server

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/kiryuhakipyatok/aloh-ssh/internal/domain/models"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/errs"

	"github.com/charmbracelet/ssh"
	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
)

func (s *Server) passwordRequest() ssh.RequestHandler {
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()

		if err := s.userService.AddPassword(appCtx, id, req.Payload); err != nil {
			if err := s.userService.DeleteUser(ctx, id); err != nil {
			}
			return false, castErr(err)
		}
		return true, nil
	}
}

func (s *Server) setNewKeyRequest() ssh.RequestHandler {
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		if err := s.userService.SetNewKey(appCtx, id, req.Payload); err != nil {
			return false, castErr(err)
		}
		return true, nil
	}
}

func (s *Server) newFriendRequest() ssh.RequestHandler {
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		var friendNickname string
		if err := json.Unmarshal(req.Payload, &friendNickname); err != nil {
			return false, castErr(err)
		}

		friendId, err := s.friendshipSerive.NewFriend(appCtx, id, friendNickname)
		if err != nil {
			return false, castErr(err)
		}
		friendSession, err := s.sessionService.GetSession(appCtx, friendId)
		if err == nil {
			fpData, err := models.MarshIdentity(id, nickname)
			if err != nil {
				return false, castErr(err)
			}
			newFriendEvent := models.NewFriendEvent(fpData)
			select {
			case friendSession.EventsChan <- newFriendEvent:
			default:
			}
		}

		return true, nil
	}
}

func (s *Server) acceptFriendshipRequest() ssh.RequestHandler {
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()

		var friendIdentity models.Identity
		if err := json.Unmarshal(req.Payload, &friendIdentity); err != nil {
			return false, castErr(err)
		}
		if err := s.friendshipSerive.AcceptFriendship(appCtx, id, friendIdentity.ID); err != nil {
			return false, castErr(err)
		}

		friendSession, frSessErr := s.sessionService.GetSession(appCtx, friendIdentity.ID)
		userSession, usSessErr := s.sessionService.GetSession(appCtx, id)

		if frSessErr == nil {
			dataIden, err := models.MarshIdentity(id, nickname)
			if err != nil {
				return false, castErr(err)
			}
			newFriendEvent := models.AcceptFriendEvent(dataIden)
			select {
			case friendSession.EventsChan <- newFriendEvent:
			default:
			}
		}

		userIdentity := models.Identity{
			ID:       id,
			Nickname: nickname,
		}

		if frSessErr == nil && usSessErr == nil {
			friendFcd := models.FriendConnsData{
				Identity: friendIdentity,
				Connects: friendSession.CurrentConnects,
			}
			friendFcdBytes, err := json.Marshal(friendFcd)
			if err != nil {
				return false, castErr(err)
			}

			friendOnlineEvent := models.FriendOnlineEvent(friendFcdBytes)
			select {
			case userSession.EventsChan <- friendOnlineEvent:
			default:
			}

			userFcd := models.FriendConnsData{
				Identity: userIdentity,
				Connects: userSession.CurrentConnects,
			}
			userFcdBytes, err := json.Marshal(userFcd)
			if err != nil {
				return false, castErr(err)
			}

			userOnlineEvent := models.FriendOnlineEvent(userFcdBytes)
			select {
			case friendSession.EventsChan <- userOnlineEvent:
			default:
			}

		}
		return true, nil
	}
}

func (s *Server) denyFriendshipRequest() ssh.RequestHandler {
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		var friendId uuid.UUID
		if err := json.Unmarshal(req.Payload, &friendId); err != nil {
			return false, castErr(err)
		}

		if err := s.friendshipSerive.DenyFriendship(appCtx, id, friendId); err != nil {
			return false, castErr(err)
		}
		return true, nil
	}
}

func (s *Server) deleteFromFriendsRequest() ssh.RequestHandler {
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		var friendId uuid.UUID
		if err := json.Unmarshal(req.Payload, &friendId); err != nil {
			return false, castErr(err)
		}

		if err := s.friendshipSerive.DeleteFromFriends(appCtx, id, friendId); err != nil {
			return false, castErr(err)
		}

		userIdenData, err := models.MarshIdentity(id, nickname)
		if err != nil {
			return false, castErr(err)
		}

		friendSession, err := s.sessionService.GetSession(appCtx, friendId)
		if err == nil {
			deleteFriendEvent := models.DeleteFriendEvent(userIdenData)
			select {
			case friendSession.EventsChan <- deleteFriendEvent:
			default:
			}
		}

		return true, nil
	}
}

func (s *Server) fetchPersonalRequest() ssh.RequestHandler {
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		personalData, err := s.userService.GetPersonalData(appCtx, id)
		if err != nil {
			return false, castErr(err)
		}

		personalDataBytes, err := json.Marshal(personalData)
		if err != nil {
			return false, castErr(err)
		}

		return true, personalDataBytes
	}
}

func (s *Server) blockUserRequest() ssh.RequestHandler {
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		var blockedNickname string
		if err := json.Unmarshal(req.Payload, &blockedNickname); err != nil {
			return false, castErr(err)
		}

		blockedId, err := s.blockedService.BlockUser(appCtx, id, blockedNickname)
		if err != nil {
			return false, castErr(err)
		}

		blockedSession, err := s.sessionService.GetSession(appCtx, blockedId)
		if err == nil {
			dataBlockerIden, err := models.MarshIdentity(id, nickname)
			if err != nil {
				return false, castErr(err)
			}
			blockFriendEvent := models.BlockUserEvent(dataBlockerIden)
			select {
			case blockedSession.EventsChan <- blockFriendEvent:
			default:
			}
		}

		blockedIdBytes, err := models.MarshID(blockedId)
		if err != nil {
			return false, castErr(err)
		}

		return true, blockedIdBytes
	}
}

func (s *Server) unblockUserRequest() ssh.RequestHandler {
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		var unblockedNick string
		if err := json.Unmarshal(req.Payload, &unblockedNick); err != nil {
			return false, castErr(err)
		}
		unblockedId, err := s.blockedService.UnblockUser(appCtx, id, unblockedNick)
		if err != nil {
			return false, castErr(err)
		}

		unblockedIdBytes, err := json.Marshal(unblockedId)
		if err != nil {
			return false, castErr(err)
		}

		return true, unblockedIdBytes
	}
}

func (s *Server) updateCurOnlineRequest() ssh.RequestHandler {
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}

		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()

		userSession, err := s.sessionService.GetSession(appCtx, id)
		if err == nil {
			if err := json.Unmarshal(req.Payload, &userSession.CurrentConnects); err != nil {
				return false, castErr(err)
			}

			usersFriends, err := s.userService.GetUsersFriends(context.Background(), id)
			if err != nil {
				return false, castErr(err)
			}
			userIdentity := models.Identity{
				ID:       id,
				Nickname: nickname,
			}
			fcd := models.FriendConnsData{
				Identity: userIdentity,
				Connects: userSession.CurrentConnects,
			}
			fcdBytes, err := json.Marshal(fcd)
			if err != nil {
				return false, castErr(err)
			}
			var wg sync.WaitGroup
			for _, friend := range usersFriends {
				wg.Go(func() {
					friendSession, err := s.sessionService.GetSession(context.Background(), friend.ID)
					if err != nil {
						return
					}

					friendOnlineEvent := models.FriendOnlineEvent(fcdBytes)
					select {
					case friendSession.EventsChan <- friendOnlineEvent:
					default:
					}
				})
			}

			wg.Wait()
		}

		return true, nil
	}
}

func (s *Server) setTaglineRequest() ssh.RequestHandler {
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		var tagline string
		if err := json.Unmarshal(req.Payload, &tagline); err != nil {
			return false, castErr(err)
		}
		if err := s.userService.SetTagline(appCtx, id, tagline); err != nil {
			return false, castErr(err)
		}

		usersFriends, err := s.userService.GetUsersFriends(context.Background(), id)
		if err != nil {
			return false, castErr(err)
		}

		tdData, err := models.MarshTD(id, nickname, tagline)
		if err != nil {
			return false, castErr(err)
		}

		var wg sync.WaitGroup
		for _, friend := range usersFriends {
			wg.Go(func() {
				friendSession, err := s.sessionService.GetSession(context.Background(), friend.ID)
				if err != nil {
					return
				}
				updateTaglineEvent := models.UpdateTaglineEvent(tdData)
				select {
				case friendSession.EventsChan <- updateTaglineEvent:
				default:
				}
			})
		}

		wg.Wait()

		return true, nil
	}
}

func (s *Server) newNicknameRequest() ssh.RequestHandler {

	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}

		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		var newNickname string
		if err := json.Unmarshal(req.Payload, &newNickname); err != nil {
			return false, castErr(err)
		}
		if err := s.userService.NewNickname(appCtx, id, newNickname); err != nil {
			return false, castErr(err)
		}

		usersFriends, err := s.userService.GetUsersFriends(context.Background(), id)
		if err != nil {
			return false, castErr(err)
		}

		usersBlockers, err := s.blockedService.FetchUsersBlockers(context.Background(), id)
		if err != nil {
			return false, castErr(err)
		}

		ndData, err := models.MarshND(id, nickname, newNickname)
		if err != nil {
			return false, castErr(err)
		}

		updateNicknameEvent := models.UpdateNicknameEvent(ndData)

		var wg sync.WaitGroup
		wg.Go(func() {
			for _, friend := range usersFriends {
				wg.Go(func() {
					friendSession, err := s.sessionService.GetSession(context.Background(), friend.ID)
					if err != nil {
						return
					}

					select {
					case friendSession.EventsChan <- updateNicknameEvent:
					default:
					}

				})
			}
		})
		wg.Go(func() {
			for _, blocker := range usersBlockers {
				wg.Go(func() {
					blockerSession, err := s.sessionService.GetSession(context.Background(), blocker.ID)
					if err != nil {
						return
					}
					select {
					case blockerSession.EventsChan <- updateNicknameEvent:
					default:
					}
				})
			}
		})

		wg.Wait()

		return true, nil
	}
}
