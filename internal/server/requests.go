package server

import (
	"aloh-ssh/internal/domain/models"
	"aloh-ssh/pkg/errs"
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/charmbracelet/ssh"
	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
)

func (s *Server) passwordRequest() ssh.RequestHandler {
	//op := "server.passwordRequest"
	//log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		//logUserId := logger.Attr("id", id)
		//log.Info("new password request", logUserId)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()

		if err := s.userService.AddPassword(appCtx, id, req.Payload); err != nil {
			//log.Error("failed to add user's password", logger.Err(err), logUserId)
			if err := s.userService.DeleteUser(ctx, id); err != nil {
				//log.Error("failed to delete user", logUserId, logger.Err(err))
			}
			return false, castErr(err)
		}
		//log.Info("user's password added successfully", logUserId)
		return true, nil
	}
}

func (s *Server) setNewKeyRequest() ssh.RequestHandler {
	//op := "server.setNewKeyRequest"
	//log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		//logUserId := logger.Attr("id", id)
		//log.Info("new set new key request", logUserId)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		if err := s.userService.SetNewKey(appCtx, id, req.Payload); err != nil {
			//log.Error("failed to set user's new keys", logger.Err(err), logUserId)
			return false, castErr(err)
		}
		//	log.Info("user's key updated successfully", logUserId)
		return true, nil
	}
}

func (s *Server) newFriendRequest() ssh.RequestHandler {
	//op := "server.newFriendRequest"
	//log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		//logUserId := logger.Attr("id", id)
		//log.Info("new friend request", logUserId)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		friendId, err := uuid.ParseBytes(req.Payload)
		if err != nil {
			//log.Error("failed to parse friend id", logger.Err(err), logUserId)
			return false, castErr(err)
		}
		if err := s.friendshipSerive.NewFriend(appCtx, id, friendId); err != nil {
			//log.Error("failed to add new friend request", logger.Err(err), logUserId)
			return false, castErr(err)
		}
		friendSession, err := s.sessionService.GetSession(appCtx, friendId)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFoundBase) {
				//log.Error("failed to get friend's session", logger.Err(err), logUserId)
			}

			return
		}
		fpData, err := models.MarshFP(id, nickname)
		if err != nil {
			//log.Error("failed to marshal nickname", logger.Err(err), logUserId)
			return false, castErr(err)
		}
		newFriendEvent := models.NewFriendEvent(fpData)
		select {
		case friendSession.EventsChan <- newFriendEvent:
		default:
		}

		//log.Info("new friend request added successfully", logUserId)
		return true, nil
	}
}

func (s *Server) acceptFriendshipRequest() ssh.RequestHandler {
	//op := "server.acceptFriendshipRequest"
	//log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		//logUserId := logger.Attr("id", id)
		//log.Info("new accept friendship request", logUserId)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		userSession, err := s.sessionService.GetSession(appCtx, id)
		if err != nil {
			//log.Error("failed to get user's session", logger.Err(err), logUserId)
			return false, castErr(err)
		}
		friendId, err := uuid.ParseBytes(req.Payload)
		if err != nil {
			//log.Error("failed to parse friend id", logger.Err(err), logUserId)
			return false, castErr(err)
		}
		if err := s.friendshipSerive.AcceptFriendship(appCtx, id, friendId); err != nil {
			return false, castErr(err)
			//log.Error("failed to accept friendship", logger.Err(err), logUserId)
		}

		friendSession, err := s.sessionService.GetSession(appCtx, friendId)
		if err != nil {
			return
		}

		userFcd := models.FriendConnsData{
			Id:       id,
			Connects: userSession.CurrentConnects,
		}
		userFcdBytes, err := json.Marshal(userFcd)
		if err != nil {
			//log.Error("failed to marshal user connects data", logger.Err(err), logUserId)
			return false, castErr(err)
		}

		friendFcd := models.FriendConnsData{
			Id:       friendId,
			Connects: friendSession.CurrentConnects,
		}
		friendFcdBytes, err := json.Marshal(friendFcd)
		if err != nil {
			//log.Error("failed to marshal friend connects data", logger.Err(err), logUserId)
			return
		}

		dataFP, err := models.MarshFP(id, nickname)
		if err != nil {
			//log.Error("failed to marshak nickname", logger.Err(err), logUserId)
			return false, castErr(err)
		}

		friendOnlineEvent := models.FriendOnlineEvent(friendFcdBytes)
		select {
		case userSession.EventsChan <- friendOnlineEvent:
		default:
		}

		newFriendEvent := models.AcceptFriendEvent(dataFP)
		select {
		case friendSession.EventsChan <- newFriendEvent:
		default:
		}

		friendOnlineEvent = models.FriendOnlineEvent(userFcdBytes)
		select {
		case friendSession.EventsChan <- friendOnlineEvent:
		default:
		}

		//log.Info("accept friendship request added successfully", logUserId)
		return true, nil
	}
}

func (s *Server) denyFriendshipRequest() ssh.RequestHandler {
	//op := "server.denyFriendshipRequest"
	//log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		//logUserId := logger.Attr("id", id)
		//log.Info("new deny friendship request", logUserId)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		friendId, err := uuid.ParseBytes(req.Payload)
		if err != nil {
			//log.Error("failed to parse friend id", logger.Err(err), logUserId)
			return false, castErr(err)
		}
		if err := s.friendshipSerive.DenyFriendship(appCtx, id, friendId); err != nil {
			//log.Error("failed to deny friendship", logger.Err(err), logUserId)
			return false, castErr(err)
		}

		//log.Info("deny friendship request added successfully", logUserId)
		return true, nil
	}
}

func (s *Server) deleteFromFriendsRequest() ssh.RequestHandler {
	//op := "server.deleteFromFriendsRequest"
	//log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		//logUserId := logger.Attr("id", id)
		//log.Info("new delete from friends request", logUserId)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()

		friendId, err := uuid.ParseBytes(req.Payload)
		if err != nil {
			//log.Error("failed to parse friend id", logger.Err(err), logUserId)
			return false, castErr(err)
		}
		if err := s.friendshipSerive.DeleteFromFriends(appCtx, id, friendId); err != nil {
			//log.Error("failed to delete from friends", logger.Err(err), logUserId)
			return false, castErr(err)
		}
		userSession, err := s.sessionService.GetSession(appCtx, id)
		if err != nil {
			//log.Error("failed to get user's session", logger.Err(err), logUserId)
			return false, castErr(err)
		}
		fpData, err := models.MarshFP(id, nickname)
		if err != nil {
			//log.Error("failed to marshal id", logger.Err(err), logUserId)
			return false, castErr(err)
		}

		dataFriendId, err := models.MarshID(friendId)
		if err != nil {
			//log.Error("failed to marshak friend id", logger.Err(err), logUserId)
			return false, castErr(err)
		}

		friendOfflineEvent := models.FriendOfflineEvent(dataFriendId)
		select {
		case userSession.EventsChan <- friendOfflineEvent:
		default:
		}
		friendSession, err := s.sessionService.GetSession(appCtx, friendId)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFoundBase) {
				//log.Error("failed to get friend's session", logger.Err(err), logUserId)
			}
			return
		}
		deleteFriendEvent := models.DeleteFriendEvent(fpData)
		select {
		case friendSession.EventsChan <- deleteFriendEvent:
		default:
		}

		friendOfflineEvent = models.FriendOfflineEvent(fpData)
		select {
		case friendSession.EventsChan <- friendOfflineEvent:
		default:
		}

		//log.Info("delete from friends request added successfully", logUserId)
		return true, nil
	}
}

// type PersonalToGet struct {
// 	models.PersonalData
// 	Friends      []string `json:"friends"`
// 	RegisterTime string   `json:"registerTime"`
// }

func (s *Server) fetchPersonalRequest() ssh.RequestHandler {
	//op := "server.fetchPersonalRequest"
	//log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		//logUserId := logger.Attr("id", id)
		//log.Info("new fetch personal data request", logUserId)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		personalData, err := s.userService.GetPersonalData(appCtx, id)
		if err != nil {
			//log.Error("failed to fetch personal data", logger.Err(err), logUserId)
			return false, castErr(err)
		}
		friendsNicknames := make([]string, 0, len(personalData.Friends))

		for _, f := range personalData.Friends {
			friendsNicknames = append(friendsNicknames, f.Nickname)
		}
		
		personalDataBytes, err := json.Marshal(personalData)
		if err != nil {
			//log.Error("failed to marshal user's personal data", logUserId, logger.Err(err))
			return false, castErr(err)
		}

		//log.Info("user's personal data fetched successfully", logUserId)
		return true, personalDataBytes
	}
}

func (s *Server) blockUserRequest() ssh.RequestHandler {
	//op := "server.blockUserRequest"
	//log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		//	logUserId := logger.Attr("id", id)
		//log.Info("new block user request", logUserId)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()

		blockedNickname := string(req.Payload)
		friendId, err := s.blockedService.BlockUser(appCtx, id, blockedNickname)
		if err != nil {
			//	log.Error("failed to block user", logger.Err(err), logUserId)
			return false, castErr(err)
		}

		if friendId != uuid.Nil {
			userSession, err := s.sessionService.GetSession(appCtx, id)
			if err != nil {
				//log.Error("failed to get user's session", logger.Err(err), logUserId)
				return false, castErr(err)
			}
			dataBlockedId, err := models.MarshID(friendId)
			if err != nil {
				//log.Error("failed to marshal nickname", logger.Err(err), logUserId)
				return false, castErr(err)
			}
			friendOfflineEvent := models.FriendOfflineEvent(dataBlockedId)
			select {
			case userSession.EventsChan <- friendOfflineEvent:
			default:
			}
			friendSession, err := s.sessionService.GetSession(appCtx, friendId)
			if err != nil {
				if !errors.Is(err, errs.ErrNotFoundBase) {
					//	log.Error("failed to get friend's session", logger.Err(err), logUserId)
				}
				return
			}
			dataId, err := models.MarshID(id)
			if err != nil {
				//log.Error("failed to marshal nickname", logger.Err(err), logUserId)
				return false, castErr(err)
			}
			dataFP, err := models.MarshFP(id, nickname)
			if err != nil {
				//log.Error("failed to marshal nickname", logger.Err(err), logUserId)
				return false, castErr(err)
			}
			blockFriendEvent := models.BlockUserEvent(dataFP)
			select {
			case friendSession.EventsChan <- blockFriendEvent:
			default:
			}

			friendOfflineEvent = models.FriendOfflineEvent(dataId)
			select {
			case friendSession.EventsChan <- friendOfflineEvent:
			default:
			}

		}

		//log.Info("user blocked successfully", logUserId)
		return true, nil
	}
}

func (s *Server) unblockUserRequest() ssh.RequestHandler {
	//op := "server.unblockUserRequest"
	//log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		//logUserId := logger.Attr("id", id)
		//log.Info("new unblock user request", logUserId)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		unblockedNick := string(req.Payload)
		if err := s.blockedService.UnblockUser(appCtx, id, unblockedNick); err != nil {
			//log.Error("failed to unblock user", logger.Err(err), logUserId)
			return false, castErr(err)
		}

		//log.Info("user unblocked successfully", logUserId)
		return true, nil
	}
}

func (s *Server) updateCurOnlineRequest() ssh.RequestHandler {
	//op := "server.updateCurOnlineRequest"
	//log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		//logUserId := logger.Attr("id", id)
		//log.Info("new update current online request", logUserId)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()

		userSession, err := s.sessionService.GetSession(appCtx, id)
		if err != nil {
			//log.Error("failed to get user's session", logger.Err(err), logUserId)
			return false, castErr(err)
		}

		if err := json.Unmarshal(req.Payload, &userSession.CurrentConnects); err != nil {
			//log.Error("failed to unmarshal request payload", logUserId)
			return false, castErr(err)
		}

		usersFriends, err := s.userService.GetUsersFriends(context.Background(), id)
		if err != nil {
			//log.Error("failed to get user's friends", logger.Err(err), logUserId)
			return false, castErr(err)
		}
		fcd := models.FriendConnsData{
			Id:       id,
			Connects: userSession.CurrentConnects,
		}
		fcdBytes, err := json.Marshal(fcd)
		if err != nil {
			//log.Error("failed to marshal friend connects data", logger.Err(err), logUserId)
			return
		}
		var wg sync.WaitGroup
		for _, friend := range usersFriends {
			wg.Go(func() {
				friendSession, err := s.sessionService.GetSession(context.Background(), friend.ID)
				if err != nil {
					if !errors.Is(err, errs.ErrNotFoundBase) {
						//log.Error("failed to get friend's session", logger.Err(err), logUserId)
					}
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

		//log.Info("user's current online updates successfully", logUserId)
		return true, nil
	}
}

func (s *Server) setTaglineRequest() ssh.RequestHandler {
	//op := "server.denyFriendshipRequest"
	//log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		id, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			return false, castErr(errs.ErrInvalidTypeBase)
		}
		//logUserId := logger.Attr("id", id)
		//log.Info("new deny friendship request", logUserId)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		tagline := string(payload)
		if err := s.userService.SetTagline(appCtx, id, tagline); err != nil {
			return false, castErr(err)
		}

		usersFriends, err := s.userService.GetUsersFriends(context.Background(), id)
		if err != nil {
			//log.Error("failed to get user's friends", logger.Err(err), logUserId)
			return false, castErr(err)
		}

		tdData, err := models.MarshTD(id, tagline)
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

		//log.Info("deny friendship request added successfully", logUserId)
		return true, nil
	}
}
