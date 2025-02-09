package pkg

import (
	"github.com/valkey-io/valkey-go"
	"reflect"
)

// ValkeyExtension is an extension that provides valkey functionality
type ValkeyExtension struct {
	client valkey.Client
}

// NewValkeyExtension creates a new valkey extension
func NewValkeyExtension(client valkey.Client) *ValkeyExtension {
	return &ValkeyExtension{client: client}
}

// Register registers the valkey extension
func (v *ValkeyExtension) Register(app *Bat) error {
	return nil
}

// Requirements returns the requirements of the valkey extension
func (v *ValkeyExtension) Requirements() []reflect.Type {
	return []reflect.Type{}
}

// GetClient returns the valkey client
func (v *ValkeyExtension) GetClient() valkey.Client {
	return v.client
}
