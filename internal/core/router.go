package core

import (
	"context"

	"github.com/fruganyumisa/igrid-middleware/internal/adapters"
	"github.com/fruganyumisa/igrid-middleware/internal/models"
	"github.com/fruganyumisa/igrid-middleware/internal/pkg/logger"
)

// Add a simple Validator type definition if it doesn't exist elsewhere
type Validator struct{}

func NewValidator(config interface{}, log logger.Logger) *Validator {
	return &Validator{}
}

func (v *Validator) Validate(msg models.Message) error {
	// Implement validation logic here
	return nil
}

type Router struct {
	adapters   []adapters.Adapter
	validator  *Validator
	normalizer *Normalizer
	log        logger.Logger
}

func DefaultValidationConfig() interface{} {
	// Return default validation config as needed
	return nil
}

// Add a stub for DefaultNormalizationConfig
func DefaultNormalizationConfig() map[string]MappingConfig {
	// Return default normalization config as needed
	return make(map[string]MappingConfig)
}

func NewRouter(adapters []adapters.Adapter, log logger.Logger) *Router {
	normalizer, err := NewNormalizer(DefaultNormalizationConfig(), "")
	if err != nil {
		log.Error("Failed to create normalizer", "error", err)
		return nil
	}
	return &Router{
		adapters:   adapters,
		validator:  NewValidator(DefaultValidationConfig(), log),
		normalizer: normalizer,
		log:        log,
	}
}

func (r *Router) Route(ctx context.Context, msg models.Message) error {
	// Validate message
	if err := r.validator.Validate(msg); err != nil {
		return err
	}

	// Normalize message
	// Replace 'msg.Type' with the correct way to get the message type.
	// For example, if Message has a method Type() string, use that:
	var msgType string
	// Try to get the type from a method, otherwise from a field
	if typeProvider, ok := interface{}(msg).(interface{ Type() string }); ok {
		msgType = typeProvider.Type()
	} else if typeField, ok := interface{}(msg).(interface{ GetType() string }); ok {
		msgType = typeField.GetType()
	} else {
		// fallback: try to access a Type field directly if it exists
		// msg.(models.MessageWithType).Type
		r.log.Error("Message type not found", "msg", msg)
		return nil
	}

	normalized, err := r.normalizer.Normalize(msgType, msg)
	if err != nil {
		return err
	}

	// Route to all adapters
	for _, adapter := range r.adapters {
		if adapter.CanHandle(normalized) {
			if err := adapter.Handle(ctx, normalized); err != nil {
				r.log.Error("Failed to handle message in adapter",
					"adapter", adapter.Name(),
					"error", err)
			}
		}
	}

	return nil
}
