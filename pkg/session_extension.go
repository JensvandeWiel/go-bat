package pkg

import (
	"context"
	"github.com/JensvandeWiel/valkeystore"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/valkey-io/valkey-go"
	"log/slog"
	"reflect"
)

const DefaultSessionName = "session"
const DefaultSessionKey = "session_id"

// SessionExtension is an extension that provides session management
type SessionExtension struct {
	vClient      valkey.Client
	logger       *Logger
	sessionStore sessions.Store
	sessionName  string
	sessionKey   string
}

// SessionExtensionOption is a function that modifies the SessionExtension
type SessionExtensionOption func(*SessionExtension) error

// WithSessionName sets the session name
func WithSessionName(name string) SessionExtensionOption {
	return func(s *SessionExtension) error {
		s.sessionName = name
		return nil
	}
}

// WithSessionStore sets the session store
func WithSessionStore(store sessions.Store) SessionExtensionOption {
	return func(s *SessionExtension) error {
		s.sessionStore = store
		return nil
	}
}

// WithSessionKey sets the session key
func WithSessionKey(key string) SessionExtensionOption {
	return func(s *SessionExtension) error {
		s.sessionKey = key
		return nil
	}
}

// NewSessionExtension creates a new session extension
func NewSessionExtension(opts ...SessionExtensionOption) (*SessionExtension, error) {
	ext := &SessionExtension{
		sessionName: DefaultSessionName,
		sessionKey:  DefaultSessionKey,
	}

	for _, opt := range opts {
		err := opt(ext)
		if err != nil {
			return nil, err
		}
	}

	return ext, nil
}

// Register registers the session extension
func (s *SessionExtension) Register(app *Bat) error {
	s.logger = &Logger{app.Logger.With("module", "session_extension")}
	valkeyExtension := GetExtension[*ValkeyExtension](app)
	s.vClient = valkeyExtension.GetClient()
	var err error
	if s.sessionStore == nil {
		s.sessionStore, err = valkeystore.NewValkeyStore(s.vClient)
		if err != nil {
			return err
		}
	}
	app.Use(
		session.Middleware(s.sessionStore),
		s.EnsureSession(),
		s.AttachSessionIDToRequestContext(),
	)
	return nil
}

// Requirements returns the requirements for the session extension
func (s *SessionExtension) Requirements() []reflect.Type {
	if s.sessionStore == nil {
		return []reflect.Type{
			reflect.TypeOf(ValkeyExtension{}),
		}
	}
	return []reflect.Type{}
}

// EnsureSession ensures that the session is created
func (s *SessionExtension) EnsureSession() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get(s.sessionName, c)
			if err != nil {
				s.logger.Error("Failed to get session", slog.String("error", err.Error()))
				return err
			}

			// Save session to ensure it's created
			if err := sess.Save(c.Request(), c.Response()); err != nil {
				s.logger.Error("Failed to save session", slog.String("error", err.Error()))
				return err
			}

			// Proceed to the next handler
			return next(c)
		}
	}
}

// AttachSessionIDToRequestContext attaches the session ID to the request context
func (s *SessionExtension) AttachSessionIDToRequestContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ss, err := session.Get(s.sessionName, c)
			if err != nil {
				s.logger.Error("Failed to get session", slog.String("error", err.Error()))
				return err
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), s.sessionKey, ss.ID)))
			return next(c)
		}
	}
}

// GetSessionIDFromRequest returns the session ID from the request context
func (s *SessionExtension) GetSessionIDFromRequest(ctx context.Context) string {
	return ctx.Value(s.sessionKey).(string)
}
