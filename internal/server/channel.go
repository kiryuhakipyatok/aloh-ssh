package server

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/kiryuhakipyatok/aloh-ssh/internal/domain/models"
	"github.com/kiryuhakipyatok/aloh-ssh/pkg/errs"

	"github.com/charmbracelet/ssh"
	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
)

func (s *Server) proccessEventChannel(srv *ssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx ssh.Context) {
	//op := "server.proccessEventChannel"
	//log := s.log.AddOp(op)
	nickname := ctx.User()
	channel, requests, err := newChan.Accept()
	if err != nil {
		//log.Error("failed to accept channel")
		return
	}
	defer channel.Close()

	go gossh.DiscardRequests(requests)
	id, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		return
	}
	//logUserId := logger.Attr("id", id)

	//log.Info("event channel accepted successfully", logUserId)

	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		//log.Error("failed to get user id", logUserId)
		return
	}
	appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
	defer cancel()
	userSession, err := s.sessionService.NewSession(appCtx, userID)
	if err != nil {
		//log.Error("failed to create session", logger.Err(err), logUserId)
		return
	}

	defer func() {
		usersFriends, err := s.userService.GetUsersFriends(context.Background(), userID)
		if err != nil {
			//log.Error("failed to get user's friends", logger.Err(err), logUserId)
		} else {
			dataId, err := models.MarshID(id)
			if err != nil {
				//log.Error("failed to marshal id", logger.Err(err), logUserId)
				return
			}
			var wg sync.WaitGroup
			for _, friend := range usersFriends {
				wg.Go(func() {
					friendSession, err := s.sessionService.GetSession(context.Background(), friend.ID)
					if err != nil {
						//if errors.Is(err, errs.ErrNotFoundBase) {
						//log.Error("failed to get friend's session", logger.Err(err), logUserId)
						//}
						return
					}
					friendOfflineEvent := models.FriendOfflineEvent(dataId)
					select {
					case friendSession.EventsChan <- friendOfflineEvent:
					default:
					}

				})
			}

			wg.Wait()
		}
		if err := s.sessionService.DeleteSession(context.Background(), userID); err != nil {
			//s.log.Error("failed to delete session", logger.Err(err), logUserId)
		}
		//s.log.Info("session deleted successfully", logUserId)
	}()

	encoder := json.NewEncoder(channel)

	userIdentity := models.Identity{
		ID:       id,
		Nickname: nickname,
	}

	usersFriends, err := s.userService.GetUsersFriends(ctx, userID)
	if err != nil {
		//log.Error("failed to get user's friends", logger.Err(err), logUserId)
	} else {
		userFcd := models.FriendConnsData{
			Identity: userIdentity,
			Connects: userSession.CurrentConnects,
		}
		userFcdBytes, err := json.Marshal(userFcd)
		if err != nil {
			//	log.Error("failed to marshal user's connects data", logger.Err(err), logUserId)

		} else {
			userFriendsOnlineEvent := models.FriendOnlineEvent(userFcdBytes)
			for _, friend := range usersFriends {

				go func(friend models.Friend) {
					friendSession, err := s.sessionService.GetSession(ctx, friend.ID)
					if err != nil {
						if !errors.Is(err, errs.ErrNotFoundBase) {
							//log.Error("failed to get friend's session", logger.Err(err), logUserId)
						}
						return
					}

					friendIdentity := models.Identity{
						ID:       friend.ID,
						Nickname: friend.Nickname,
					}

					friendFcd := models.FriendConnsData{
						Identity: friendIdentity,
						Connects: friendSession.CurrentConnects,
					}
					friendFcdBytes, err := json.Marshal(friendFcd)
					if err != nil {
						//log.Error("failed to marshal friend's connects data", logger.Err(err), logUserId)
						return
					}

					friendsOnlineEvent := models.FriendOnlineEvent(friendFcdBytes)
					select {
					case userSession.EventsChan <- friendsOnlineEvent:
						//log.Info("friend online event sended successfully", logUserId)
					default:
					}

					select {
					case friendSession.EventsChan <- userFriendsOnlineEvent:
						//log.Info("user online event sended successfully", logUserId)
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
					//log.Error("failed to send keepalive request", logger.Err(err))
				}
			}

		}

	}()

	for {
		select {
		case <-ctx.Done():
			//log.Info("context done", logUserId)
			return
		case event := <-userSession.EventsChan:
			if err := encoder.Encode(event); err != nil {
				//log.Error("failed to encode evennt", logger.Err(err), logUserId)
				return
			}
		}
	}

}
