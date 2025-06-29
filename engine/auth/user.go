package auth

import (
	"context"
	"log/slog"
	"strings"

	"github.com/aarondl/authboss/v3"
	"github.com/gofrs/uuid/v5"
	"github.com/pkg/errors"

	"fixit/engine/ent"
	user2 "fixit/engine/ent/user"
)

type User struct {
	*ent.User
}

func (u User) GetPID() string {
	return u.Email
}

func (u User) PutPID(pid string) {
	u.Email = pid
}

func (u User) GetPassword() string {
	return u.Password
}

func (u User) PutPassword(password string) {
	u.Password = password
}

func (u User) GetEmail() string {
	return u.Email
}

func (u User) PutEmail(email string) {
	u.Email = email
}

func (u User) GetUsername() string {
	return u.Username
}

func (u User) PutUsername(username string) {
	u.Username = username
}

func (u User) GetConfirmed() bool {
	return true
}

func (u User) PutConfirmed(confirmed bool) {
}

func (u User) GetLocked() bool {
	return false
}

func (u User) PutLocked(locked bool) {
}

func (u User) GetArbitrary() map[string]string {
	return map[string]string{
		"username": u.Username,
	}
}

func (u User) PutArbitrary(arbitrary map[string]string) {
	if username, ok := arbitrary["username"]; ok {
		u.Username = username
	}
}

type Storer struct {
	client *ent.Client
}

func NewStorer(client *ent.Client) *Storer {
	return &Storer{client: client}
}

func (s *Storer) Load(ctx context.Context, key string) (authboss.User, error) {
	user, err := s.client.User.Query().Where(
		user2.EmailEQ(key),
	).Only(ctx)
	slog.Info("load result", "user", user, "err", err)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, authboss.ErrUserNotFound
		}
		return nil, err
	}
	return User{user}, nil
}

func (s *Storer) Save(ctx context.Context, user authboss.User) error {
	u := user.(User)

	if u.ID == uuid.Nil {
		// Generate username from email if not provided
		username := u.Username
		if username == "" {
			// Extract username part from email before @
			emailParts := strings.Split(u.Email, "@")
			if len(emailParts) > 0 {
				username = emailParts[0]
			} else {
				username = "user" + u.Email[:5] // Fallback
			}
		}

		_, err := s.client.User.Create().
			SetUsername(username).
			SetEmail(u.Email).
			SetPassword(u.Password).
			Save(ctx)
		return errors.WithStack(err)
	}

	_, err := s.client.User.UpdateOneID(u.ID).
		SetUsername(u.Username).
		SetEmail(u.Email).
		SetPassword(u.Password).
		Save(ctx)
	return errors.WithStack(err)
}

func (s *Storer) New(ctx context.Context) authboss.User {
	return User{&ent.User{}}
}

func (s *Storer) Create(ctx context.Context, user authboss.User) error {
	return s.Save(ctx, user)
}

func (s *Storer) LoadByConfirmSelector(ctx context.Context, selector string) (authboss.User, error) {
	return s.Load(ctx, selector)
}
