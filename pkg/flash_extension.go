package pkg

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/romsar/gonertia"
	"github.com/valkey-io/valkey-go"
	"log/slog"
	"reflect"
	"strconv"
	"time"
)

const DefaultFlashErrKeyPrefix = "flash_err:"
const DefaultFlashClearHistoryKeyPrefix = "flash_clear_history:"

// FlashExtension is an extension that provides error flash functionality
type FlashExtension struct {
	client                     valkey.Client
	logger                     *slog.Logger
	sessionExtension           *SessionExtension
	flashErrKeyPrefix          string
	flashClearHistoryKeyPrefix string
}

// FlashExtensionOption is a function that modifies the FlashExtension
type FlashExtensionOption func(*FlashExtension) error

// WithFlashErrKeyPrefix sets the flash error key prefix
func WithFlashErrKeyPrefix(prefix string) FlashExtensionOption {
	return func(f *FlashExtension) error {
		f.flashErrKeyPrefix = prefix
		return nil
	}
}

// WithFlashClearHistoryKeyPrefix sets the flash clear history key prefix
func WithFlashClearHistoryKeyPrefix(prefix string) FlashExtensionOption {
	return func(f *FlashExtension) error {
		f.flashClearHistoryKeyPrefix = prefix
		return nil
	}
}

// NewFlashExtension creates a new flash extension, it expects that the context contains a sessionID
func NewFlashExtension(opts ...FlashExtensionOption) (*FlashExtension, error) {
	ext := &FlashExtension{
		flashErrKeyPrefix:          DefaultFlashErrKeyPrefix,
		flashClearHistoryKeyPrefix: DefaultFlashClearHistoryKeyPrefix,
	}

	for _, opt := range opts {
		err := opt(ext)
		if err != nil {
			return nil, err
		}
	}

	return ext, nil
}

// Register registers the flash extension
func (f *FlashExtension) Register(app *Bat) error {
	f.client = GetExtension[*SessionExtension](app).vClient
	f.sessionExtension = GetExtension[*SessionExtension](app)
	f.logger = app.Logger.With("module", "flash_extension")
	return nil
}

// Requirements returns the requirements for the flash extension
func (f *FlashExtension) Requirements() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf(ValkeyExtension{}),
		reflect.TypeOf(SessionExtension{}),
	}
}

// FlashErrors adds the errors to the flash provider
func (f *FlashExtension) FlashErrors(ctx context.Context, errors gonertia.ValidationErrors) error {
	sessionID := f.sessionExtension.GetSessionIDFromRequest(ctx)
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(errors)
	if err != nil {
		return err
	}

	err = f.client.Do(context.Background(), f.client.B().Set().Key(f.flashErrKeyPrefix+sessionID).Value(buffer.String()).Ex(time.Hour*24).Build()).Error()
	if err != nil {
		return err
	}
	f.logger.Debug("Flash errors", "sessionID", sessionID, "errors", errors)
	for key, value := range errors {
		f.logger.Debug("Flash error", "key", key, "value", value)
	}
	return nil
}

// GetErrors returns the errors from the flash provider
func (f *FlashExtension) GetErrors(ctx context.Context) (gonertia.ValidationErrors, error) {
	sessionID := f.sessionExtension.GetSessionIDFromRequest(ctx)
	var errs gonertia.ValidationErrors
	res := f.client.Do(context.Background(), f.client.B().Get().Key(f.flashErrKeyPrefix+sessionID).Build())
	if res.Error() != nil {
		if valkey.IsValkeyNil(res.Error()) {
			// No errors found
			return gonertia.ValidationErrors{}, nil
		}
		return gonertia.ValidationErrors{}, res.Error()
	}

	b, err := res.AsBytes()
	if err != nil {
		return gonertia.ValidationErrors{}, err
	}

	err = gob.NewDecoder(bytes.NewReader(b)).Decode(&errs)
	if err != nil {
		return gonertia.ValidationErrors{}, err
	}
	f.logger.Debug("Get errors", "sessionID", sessionID)
	for key, value := range errs {
		f.logger.Debug("Got error", "key", key, "value", value)
	}

	// Clear the errors
	err = f.client.Do(context.Background(), f.client.B().Del().Key(f.flashErrKeyPrefix+sessionID).Build()).Error()
	if err != nil {
		return gonertia.ValidationErrors{}, err
	}

	return errs, nil
}

// FlashClearHistory sets the flash clear history flag
func (f *FlashExtension) FlashClearHistory(ctx context.Context) error {
	sessionID := f.sessionExtension.GetSessionIDFromRequest(ctx)

	err := f.client.Do(context.Background(), f.client.B().Set().Key(f.flashClearHistoryKeyPrefix+sessionID).Value("true").Build()).Error()
	if err != nil {
		return err
	}
	f.logger.Debug("Flash clear history set", "sessionID", sessionID)
	return nil
}

// ShouldClearHistory returns whether the history should be cleared
func (f *FlashExtension) ShouldClearHistory(ctx context.Context) (bool, error) {
	sessionID := f.sessionExtension.GetSessionIDFromRequest(ctx)

	res := f.client.Do(context.Background(), f.client.B().Get().Key(f.flashClearHistoryKeyPrefix+sessionID).Build())
	if res.Error() != nil {
		if valkey.IsValkeyNil(res.Error()) {
			// No value found, return false
			return false, nil
		}
		return false, res.Error()
	}

	strValue := res.String()

	val, err := strconv.ParseBool(strValue)
	if err != nil {
		return false, err
	}
	f.logger.Debug("Should clear history", "sessionID", sessionID, "value", val)
	return val, nil
}
