package server

import (
	"aloh-ssh/internal/domain/models"
	"aloh-ssh/pkg/errs"
	"aloh-ssh/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
)

func (s *Server) passwordRequest() ssh.RequestHandler {
	op := "server.passwordRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new password request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()

		if err := s.userService.AddPassword(appCtx, nickname, req.Payload); err != nil {
			log.Error("failed to add user's password", logger.Err(err), logUserNickname)
			if err := s.userService.DeleteUser(ctx, nickname); err != nil {
				log.Error("failed to delete user", logUserNickname, logger.Err(err))
			}
			return false, castErr(err)
		}
		log.Info("user's password added successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) setNewKeyRequest() ssh.RequestHandler {
	op := "server.setNewKeyRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new login request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		if err := s.userService.SetNewKey(appCtx, nickname, req.Payload); err != nil {
			log.Error("failed to set user's new keys", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}
		log.Info("user's key updated successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) newFriendRequest() ssh.RequestHandler {
	op := "server.newFriendRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new friend request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		userID, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			log.Error("failed to get user id", logUserNickname)
			return false, castErr(errs.ErrInvalidType(op))
		}
		friendNickname := string(req.Payload)
		friendId, err := s.friendshipSerive.NewFriend(appCtx, userID, friendNickname)
		if err != nil {
			log.Error("failed to add new friend request", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}
		friendSession, err := s.sessionService.GetSession(appCtx, friendId)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFoundBase) {
				log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
			}

			return
		}
		dataNick, err := models.MarshNick(nickname)
		if err != nil {
			log.Error("failed to marshak nickname", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}
		newFriendEvent := models.NewFriendEvent(dataNick)
		select {
		case friendSession.EventsChan <- newFriendEvent:
		default:
		}

		log.Info("new friend request added successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) acceptFriendshipRequest() ssh.RequestHandler {
	op := "server.acceptFriendshipRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new accept friendship request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		userID, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			log.Error("failed to get user id", logUserNickname)
			return false, castErr(errs.ErrInvalidType(op))
		}
		userSession, err := s.sessionService.GetSession(appCtx, userID)
		if err != nil {
			log.Error("failed to get user's session", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}
		friendNickname := string(req.Payload)
		friendId, err := s.friendshipSerive.AcceptFriendship(appCtx, userID, friendNickname)
		if err != nil {
			log.Error("failed to accept friendship", logger.Err(err), logUserNickname)
		}

		friendSession, err := s.sessionService.GetSession(appCtx, friendId)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFoundBase) {
				log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
			}

			return
		}

		userFcd := models.FriendConnsData{
			Nickname: nickname,
			Connects: userSession.CurrentConnects,
		}
		userFcdBytes, err := json.Marshal(userFcd)
		if err != nil {
			log.Error("failed to marshal user connects data", logger.Err(err), logUserNickname)
			return
		}

		friendFcd := models.FriendConnsData{
			Nickname: friendNickname,
			Connects: friendSession.CurrentConnects,
		}
		friendFcdBytes, err := json.Marshal(friendFcd)
		if err != nil {
			log.Error("failed to marshal friend connects data", logger.Err(err), logUserNickname)
			return
		}

		dataNick, err := models.MarshNick(nickname)
		if err != nil {
			log.Error("failed to marshak nickname", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}

		friendOnlineEvent := models.FriendOnlineEvent(friendFcdBytes)
		select {
		case userSession.EventsChan <- friendOnlineEvent:
		default:
		}

		newFriendEvent := models.AcceptFriendEvent(dataNick)
		select {
		case friendSession.EventsChan <- newFriendEvent:
		default:
		}

		friendOnlineEvent = models.FriendOnlineEvent(userFcdBytes)
		select {
		case friendSession.EventsChan <- friendOnlineEvent:
		default:
		}

		log.Info("accept friendship request added successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) denyFriendshipRequest() ssh.RequestHandler {
	op := "server.denyFriendshipRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new deny friendship request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		userID, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			log.Error("failed to get user id", logUserNickname)
			return false, castErr(errs.ErrInvalidType(op))
		}
		friendNickname := string(req.Payload)
		if err := s.friendshipSerive.DenyFriendship(appCtx, userID, friendNickname); err != nil {
			log.Error("failed to deny friendship", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}

		log.Info("deny friendship request added successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) deleteFromFriendsRequest() ssh.RequestHandler {
	op := "server.deleteFromFriendsRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new delete from friends request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		userID, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			log.Error("failed to get user id", logUserNickname)
			return false, castErr(errs.ErrInvalidType(op))
		}
		friendNickname := string(req.Payload)
		friendId, err := s.friendshipSerive.DeleteFromFriends(appCtx, userID, friendNickname)
		if err != nil {
			log.Error("failed to delete from friends", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}
		userSession, err := s.sessionService.GetSession(appCtx, userID)
		if err != nil {
			log.Error("failed to get user's session", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}
		dataNick, err := models.MarshNick(nickname)
		if err != nil {
			log.Error("failed to marshal nickname", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}

		dataFriendNick, err := models.MarshNick(friendNickname)
		if err != nil {
			log.Error("failed to marshak friend nickname", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}

		friendOfflineEvent := models.FriendOfflineEvent(dataFriendNick)
		select {
		case userSession.EventsChan <- friendOfflineEvent:
		default:
		}
		friendSession, err := s.sessionService.GetSession(appCtx, friendId)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFoundBase) {
				log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
			}
			return
		}
		deleteFriendEvent := models.DeleteFriendEvent(dataNick)
		select {
		case friendSession.EventsChan <- deleteFriendEvent:
		default:
		}

		friendOfflineEvent = models.FriendOfflineEvent(dataNick)
		select {
		case friendSession.EventsChan <- friendOfflineEvent:
		default:
		}

		log.Info("delete from friends request added successfully", logUserNickname)
		return true, nil
	}
}

type PersonalToGet struct {
	models.PersonalData
	Friends      []string `json:"friends"`
	RegisterTime string   `json:"registerTime"`
}

func (s *Server) fetchPersonalRequest() ssh.RequestHandler {
	op := "server.fetchPersonalRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new fetch personal data request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		userID, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			log.Error("failed to get user id", logUserNickname)
			return false, castErr(errs.ErrInvalidType(op))
		}
		personalData, err := s.userService.GetPersonalData(appCtx, userID)
		if err != nil {
			log.Error("failed to fetch personal data", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}
		friendsNicknames := make([]string, 0, len(personalData.Friends))

		for _, f := range personalData.Friends {
			friendsNicknames = append(friendsNicknames, f.Nickname)
		}
		pdg := PersonalToGet{
			PersonalData: models.PersonalData{
				Nickname:     personalData.Nickname,
				FriendsReqs:  personalData.FriendsReqs,
				BlockedUsers: personalData.BlockedUsers,
			},
			RegisterTime: personalData.RegisterTime.Format("2006-01-02"),
			Friends:      friendsNicknames,
		}

		personalDataBytes, err := json.Marshal(pdg)
		if err != nil {
			log.Error("failed to marshal user's personal data", logUserNickname, logger.Err(err))
			return false, castErr(err)
		}

		// for _, friend := range personalData.Friends {
		// 	go func(friend models.Friend) {
		// 		frCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		// 		defer cancel()
		// 		friendSession, err := s.sessionService.GetSession(frCtx, friend.ID)
		// 		if err != nil {
		// 			if errors.Is(err, errs.ErrNotFoundBase) {
		// 				log.Info("friend is offline", logUserNickname)
		// 			} else {
		// 				log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
		// 			}
		// 			return
		// 		}
		// 		friendsOnlineEvent := models.FriendOnlineEvent([]byte(nickname))
		// 		select {
		// 		case friendSession.EventsChan <- friendsOnlineEvent:
		// 			log.Info("online event sended successfully", logUserNickname)
		// 		default:
		// 		}
		// 	}(friend)
		// }

		log.Info("user's personal data fetched successfully", logUserNickname)
		return true, personalDataBytes
	}
}

func (s *Server) blockUserRequest() ssh.RequestHandler {
	op := "server.blockUserRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new block user request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		userID, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			log.Error("failed to get user id", logUserNickname)
			return false, castErr(errs.ErrInvalidType(op))
		}
		userNickname := string(req.Payload)
		friendId, err := s.blockedService.BlockUser(appCtx, userID, userNickname)
		if err != nil {
			log.Error("failed to block user", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}

		if friendId != uuid.Nil {
			userSession, err := s.sessionService.GetSession(appCtx, userID)
			if err != nil {
				log.Error("failed to get user's session", logger.Err(err), logUserNickname)
				return false, castErr(err)
			}
			dataUserNick, err := models.MarshNick(userNickname)
			if err != nil {
				log.Error("failed to marshal nickname", logger.Err(err), logUserNickname)
				return false, castErr(err)
			}
			friendOfflineEvent := models.FriendOfflineEvent(dataUserNick)
			select {
			case userSession.EventsChan <- friendOfflineEvent:
			default:
			}
			friendSession, err := s.sessionService.GetSession(appCtx, friendId)
			if err != nil {
				if !errors.Is(err, errs.ErrNotFoundBase) {
					log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
				}
				return
			}
			dataNick, err := models.MarshNick(nickname)
			if err != nil {
				log.Error("failed to marshal nickname", logger.Err(err), logUserNickname)
				return false, castErr(err)
			}
			blockFriendEvent := models.BlockUserEvent(dataNick)
			select {
			case friendSession.EventsChan <- blockFriendEvent:
			default:
			}

			friendOfflineEvent = models.FriendOfflineEvent(dataNick)
			select {
			case friendSession.EventsChan <- friendOfflineEvent:
			default:
			}

		}

		log.Info("user blocked successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) unblockUserRequest() ssh.RequestHandler {
	op := "server.unblockUserRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new unblock user request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		userID, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			log.Error("failed to get user id", logUserNickname)
			return false, castErr(errs.ErrInvalidType(op))
		}
		userNickname := string(req.Payload)
		if err := s.blockedService.UnblockUser(appCtx, userID, userNickname); err != nil {
			log.Error("failed to unblock user", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}

		log.Info("user unblocked successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) updateCurOnlineRequest() ssh.RequestHandler {
	op := "server.updateCurOnlineRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new update current online request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
		defer cancel()
		userID, ok := ctx.Value("userID").(uuid.UUID)
		if !ok {
			log.Error("failed to get user id", logUserNickname)
			return false, castErr(errs.ErrInvalidType(op))
		}

		userSession, err := s.sessionService.GetSession(appCtx, userID)
		if err != nil {
			log.Error("failed to get user's session", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}

		if err := json.Unmarshal(req.Payload, &userSession.CurrentConnects); err != nil {
			log.Error("failed to unmarshal request payload", logUserNickname)
			return false, castErr(err)
		}

		usersFriends, err := s.userService.GetUsersFriends(context.Background(), userID)
		if err != nil {
			log.Error("failed to get user's friends", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}
		fcd := models.FriendConnsData{
			Nickname: nickname,
			Connects: userSession.CurrentConnects,
		}
		fcdBytes, err := json.Marshal(fcd)
		if err != nil {
			log.Error("failed to marshal friend connects data", logger.Err(err), logUserNickname)
			return
		}
		var wg sync.WaitGroup
		for _, friend := range usersFriends {
			wg.Go(func() {
				friendSession, err := s.sessionService.GetSession(context.Background(), friend.ID)
				if err != nil {
					if !errors.Is(err, errs.ErrNotFoundBase) {
						log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
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

		log.Info("user's current online updates successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) proccessEventChannel(srv *ssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx ssh.Context) {
	op := "server.proccessEventChannel"
	log := s.log.AddOp(op)

	channel, requests, err := newChan.Accept()
	if err != nil {
		log.Error("failed to accept channel")
		return
	}

	go gossh.DiscardRequests(requests)
	nickname := ctx.User()
	logUserNickname := logger.Attr("nickname", nickname)

	log.Info("event channel accepted successfully", logUserNickname)

	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		log.Error("failed to get user id", logUserNickname)
		return
	}
	appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
	defer cancel()
	userSession, err := s.sessionService.NewSession(appCtx, userID)
	if err != nil {
		log.Error("failed to create session", logger.Err(err), logUserNickname)
		return
	}

	defer func() {
		usersFriends, err := s.userService.GetUsersFriends(context.Background(), userID)
		if err != nil {
			log.Error("failed to get user's friends", logger.Err(err), logUserNickname)
		} else {
			dataNick, err := models.MarshNick(nickname)
			if err != nil {
				log.Error("failed to marshal nickname", logger.Err(err), logUserNickname)
				return
			}
			var wg sync.WaitGroup
			for _, friend := range usersFriends {
				wg.Go(func() {
					friendSession, err := s.sessionService.GetSession(context.Background(), friend.ID)
					if err != nil {
						if !errors.Is(err, errs.ErrNotFoundBase) {
							log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
						}
						return
					}
					friendOfflineEvent := models.FriendOfflineEvent(dataNick)
					select {
					case friendSession.EventsChan <- friendOfflineEvent:
					default:
					}

				})
			}

			wg.Wait()
		}
		channel.Close()
		if err := s.sessionService.DeleteSession(context.Background(), userID); err != nil {
			s.log.Error("failed to delete session", logger.Err(err), logUserNickname)
		}
		s.log.Info("session deleted successfully", logUserNickname)
	}()

	encoder := json.NewEncoder(channel)

	usersFriends, err := s.userService.GetUsersFriends(ctx, userID)
	if err != nil {
		log.Error("failed to get user's friends", logger.Err(err), logUserNickname)
	} else {
		userFcd := models.FriendConnsData{
			Nickname: nickname,
			Connects: userSession.CurrentConnects,
		}
		userFcdBytes, err := json.Marshal(userFcd)
		if err != nil {
			log.Error("failed to marshal user's connects data", logger.Err(err), logUserNickname)

		} else {
			userFriendsOnlineEvent := models.FriendOnlineEvent(userFcdBytes)
			for _, friend := range usersFriends {

				go func(friend models.Friend) {
					friendSession, err := s.sessionService.GetSession(ctx, friend.ID)
					if err != nil {
						if !errors.Is(err, errs.ErrNotFoundBase) {
							log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
						}
						return
					}

					friendFcd := models.FriendConnsData{
						Nickname: friend.Nickname,
						Connects: friendSession.CurrentConnects,
					}
					friendFcdBytes, err := json.Marshal(friendFcd)
					if err != nil {
						log.Error("failed to marshal friend's connects data", logger.Err(err), logUserNickname)
						return
					}

					friendsOnlineEvent := models.FriendOnlineEvent(friendFcdBytes)
					select {
					case userSession.EventsChan <- friendsOnlineEvent:
						log.Info("friend online event sended successfully", logUserNickname)
					default:
					}

					select {
					case friendSession.EventsChan <- userFriendsOnlineEvent:
						log.Info("user online event sended successfully", logUserNickname)
					default:
					}

				}(friend)

			}
		}

	}

	go func() {
		ticker := time.NewTicker(s.cfg.KeepAliveTimeout)

		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := channel.SendRequest("keepalive", false, nil); err != nil {
					log.Error("failed to send keepalive request", logger.Err(err))
				}
			}

		}

	}()

	for {
		select {
		case <-ctx.Done():
			log.Info("context done", logUserNickname)
			return
		case event := <-userSession.EventsChan:
			if err := encoder.Encode(event); err != nil {
				log.Error("failed to encode evennt", logger.Err(err), logUserNickname)
				return
			}
		}
	}

}
