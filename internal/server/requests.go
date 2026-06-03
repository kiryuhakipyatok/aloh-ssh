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

func (s *Server) passwordRequest(timeout time.Duration) ssh.RequestHandler {
	op := "server.passwordRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new password request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
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

func (s *Server) setNewKeyRequest(timeout time.Duration) ssh.RequestHandler {
	op := "server.setNewKeyRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new login request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := s.userService.SetNewKey(appCtx, nickname, req.Payload); err != nil {
			log.Error("failed to set user's new keys", logger.Err(err), logUserNickname)
			return false, castErr(err)
		}
		log.Info("user's key updated successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) newFriendRequest(timeout time.Duration) ssh.RequestHandler {
	op := "server.newFriendRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new friend request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
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
		newFriendEvent := models.NewFriendEvent(nickname)
		select {
		case friendSession.EventsChan <- newFriendEvent:
		default:
		}

		log.Info("new friend request added successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) acceptFriendshipRequest(timeout time.Duration) ssh.RequestHandler {
	op := "server.acceptFriendshipRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new accept friendship request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
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

		friendOnlineEvent := models.FriendOnlineEvent(friendNickname)
		select {
		case userSession.EventsChan <- friendOnlineEvent:
		default:
		}

		friendSession, err := s.sessionService.GetSession(appCtx, friendId)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFoundBase) {
				log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
			}

			return
		}
		newFriendEvent := models.AcceptFriendEvent(nickname)
		select {
		case friendSession.EventsChan <- newFriendEvent:
		default:
		}

		friendOnlineEvent = models.FriendOnlineEvent(nickname)
		select {
		case friendSession.EventsChan <- friendOnlineEvent:
		default:
		}

		log.Info("accept friendship request added successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) denyFriendshipRequest(timeout time.Duration) ssh.RequestHandler {
	op := "server.denyFriendshipRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new deny friendship request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
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

func (s *Server) deleteFromFriendsRequest(timeout time.Duration) ssh.RequestHandler {
	op := "server.deleteFromFriendsRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new delete from friends request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
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
		friendOfflineEvent := models.FriendOfflineEvent(friendNickname)
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
		deleteFriendEvent := models.DeleteFriendEvent(nickname)
		select {
		case friendSession.EventsChan <- deleteFriendEvent:
		default:
		}

		friendOfflineEvent = models.FriendOfflineEvent(nickname)
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

func (s *Server) fetchPersonalRequest(timeout time.Duration) ssh.RequestHandler {
	op := "server.fetchPersonalRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new fetch personal data request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
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

		for _, friend := range personalData.Friends {
			go func(friend models.Friend) {
				frCtx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				friendSession, err := s.sessionService.GetSession(frCtx, friend.ID)
				if err != nil {
					if errors.Is(err, errs.ErrNotFoundBase) {
						log.Info("friend is offline", logUserNickname)
					} else {
						log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
					}
					return
				}
				friendsOnlineEvent := models.FriendOnlineEvent(nickname)
				select {
				case friendSession.EventsChan <- friendsOnlineEvent:
					log.Info("online event sended successfully", logUserNickname)
				default:
				}
			}(friend)
		}

		log.Info("user's personal data fetched successfully", logUserNickname)
		return true, personalDataBytes
	}
}

func (s *Server) blockUserRequest(timeout time.Duration) ssh.RequestHandler {
	op := "server.blockUserRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new block user request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
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
			friendOfflineEvent := models.FriendOfflineEvent(userNickname)
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
			blockFriendEvent := models.BlockUserEvent(nickname)
			select {
			case friendSession.EventsChan <- blockFriendEvent:
			default:
			}

			friendOfflineEvent = models.FriendOfflineEvent(nickname)
			select {
			case friendSession.EventsChan <- friendOfflineEvent:
			default:
			}

		}

		log.Info("user blocked successfully", logUserNickname)
		return true, nil
	}
}

func (s *Server) unblockUserRequest(timeout time.Duration) ssh.RequestHandler {
	op := "server.unblockUserRequest"
	log := s.log.AddOp(op)
	return func(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (ok bool, payload []byte) {
		nickname := ctx.User()
		logUserNickname := logger.Attr("nickname", nickname)
		log.Info("new unblock user request", logUserNickname)
		appCtx, cancel := context.WithTimeout(context.Background(), timeout)
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
	if err := s.sessionService.NewSession(appCtx, userID); err != nil {
		log.Error("failed to create session", logger.Err(err), logUserNickname)
		return
	}

	defer func() {
		usersFriends, err := s.userService.GetUsersFriends(context.Background(), userID)
		if err != nil {
			log.Error("failed to get user", logger.Err(err), logUserNickname)
		} else {
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
					friendOfflineEvent := models.FriendOfflineEvent(nickname)
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

	userSession, err := s.sessionService.GetSession(ctx, userID)
	if err != nil {
		log.Error("failed to get user's session", logger.Err(err), logUserNickname)
		return
	}

	encoder := json.NewEncoder(channel)

	usersFriends, err := s.userService.GetUsersFriends(ctx, userID)
	if err != nil {
		log.Error("failed to get user's friends", logger.Err(err), logUserNickname)
	} else {
		for _, friend := range usersFriends {

			go func(friend models.Friend) {
				_, err := s.sessionService.GetSession(ctx, friend.ID)
				if err != nil {
					if !errors.Is(err, errs.ErrNotFoundBase) {
						log.Error("failed to get friend's session", logger.Err(err), logUserNickname)
					}
					return
				}
				friendsOnlineEvent := models.FriendOnlineEvent(friend.Nickname)
				select {
				case userSession.EventsChan <- friendsOnlineEvent:
					log.Info("online event sended successfully", logUserNickname)
				default:
				}
			}(friend)

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
