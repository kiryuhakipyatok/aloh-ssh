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
	nickname := ctx.User()
	channel, requests, err := newChan.Accept()
	if err != nil {
		return
	}
	defer channel.Close()

	go gossh.DiscardRequests(requests)
	id, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		return
	}

	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		return
	}
	appCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
	defer cancel()
	userSession, err := s.sessionService.NewSession(appCtx, userID)
	if err != nil {
		return
	}

	defer func() {
		usersFriends, err := s.userService.GetUsersFriends(context.Background(), userID)
		if err != nil {
		} else {
			dataId, err := models.MarshID(id)
			if err != nil {
				return
			}
			var wg sync.WaitGroup
			for _, friend := range usersFriends {
				wg.Go(func() {
					friendSession, err := s.sessionService.GetSession(context.Background(), friend.ID)
					if err != nil {
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
		}
	}()

	encoder := json.NewEncoder(channel)

	userIdentity := models.Identity{
		ID:       id,
		Nickname: nickname,
	}

	usersFriends, err := s.userService.GetUsersFriends(ctx, userID)
	if err != nil {

	} else {
		userFcd := models.FriendConnsData{
			Identity: userIdentity,
			Connects: userSession.CurrentConnects,
		}
		userFcdBytes, err := json.Marshal(userFcd)
		if err != nil {


		} else {
			userFriendsOnlineEvent := models.FriendOnlineEvent(userFcdBytes)
			for _, friend := range usersFriends {

				go func(friend models.Friend) {
					friendSession, err := s.sessionService.GetSession(ctx, friend.ID)
					if err != nil {
						if !errors.Is(err, errs.ErrNotFoundBase) {

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
			
						return
					}

					friendsOnlineEvent := models.FriendOnlineEvent(friendFcdBytes)
					select {
					case userSession.EventsChan <- friendsOnlineEvent:
				
					default:
					}

					select {
					case friendSession.EventsChan <- userFriendsOnlineEvent:
			
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
		
				}
			}

		}

	}()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-userSession.EventsChan:
			if err := encoder.Encode(event); err != nil {
				return
			}
		}
	}

}
