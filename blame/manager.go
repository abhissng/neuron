package blame

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// BlameDefinition represents a blame definition.
type BlameDefinition struct {
	StatusCode   string `json:"StatusCode"`
	Code         string `json:"Code"`
	Message      string `json:"Message"`
	Description  string `json:"Description"`
	Component    string `json:"Component"`
	ResponseType string `json:"ResponseType"`
}

// BlameManager is a wrapper around the blame definitions.
type BlameManager struct {
	BlameDefinitions map[types.ErrorCode]Blame
}

// RetrieveBlameCache retrieves a blame definition from the cache.
func (bw *BlameManager) RetrieveBlameCache(errorCode types.ErrorCode) Blame {
	if cache, ok := bw.BlameDefinitions[errorCode]; ok {
		return cache
	}
	return NewBasicBlame(types.ErrorCode(errorCode))
}

// FetchBlameForError fetches a blame definition for the given error code.
func (bw *BlameManager) FetchBlameForError(errorCode types.ErrorCode, opts ...BlameOption) Blame {
	return bw.RetrieveBlameCache(errorCode).EmptyCause().Wrap(opts...)
}

// NewBlameManager creates a new BlameManager instance.
func NewBlameManager(opt *BlameManagerOption) (*BlameManager, error) {
	if opt.Bundle == nil {
		opt.Bundle = helpers.NewBundle(helpers.ParseLanguageTag(opt.LanguageTag))
	}

	err := InitLocalBlameManager(opt.Bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise local blame manager: %w", err)
	}

	var blameDefinitions []BlameDefinition

	// Load error definitions from JSON file
	if !helpers.IsEmpty(opt.LocaleDir) {
		file, err := os.Open(filepath.Clean(opt.LocaleDir))
		if err != nil {
			return nil, fmt.Errorf("failed to open error definitions file: %w", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				helpers.Println(constant.ERROR, "Error closing file: ", err)
			}
		}()

		if err := json.NewDecoder(file).Decode(&blameDefinitions); err != nil {
			return nil, fmt.Errorf("failed to decode error definitions: %w", err)
		}
	}

	// Create a map of error definitions
	blameDefinitionsMap := make(map[types.ErrorCode]Blame)

	if opt.ExistingManager != nil {
		blameDefinitionsMap = opt.ExistingManager.BlameDefinitions
	}

	if len(blameDefinitions) > 0 {
		for index, def := range blameDefinitions {
			if helpers.IsEmpty(def.StatusCode) {
				def.StatusCode = helpers.GenerateStatusCode(helpers.GetServiceName(), 100+index)
			}
			blameDefinitionsMap[types.ErrorCode(def.Code)] =
				NewBlame(def.StatusCode, types.ErrorCode(def.Code), def.Message, def.Description).
					WithComponent(types.ComponentErrorType(def.Component)).
					WithResponseType(types.ResponseErrorType(def.ResponseType)).
					WithBundle(opt.Bundle)
		}

		// return &BlameManager{
		// 	BlameDefinitions: blameDefinitionsMap,
		// }, nil
	}
	if len(blameDefinitionsMap) > 0 {
		return &BlameManager{
			BlameDefinitions: blameDefinitionsMap,
		}, nil
	}
	return localBlameManager, nil

}

// BlameOption defines an option for modifying Blame creation.
type BlameOption func(*BlameOptions)

// BlameOptions holds options for creating Blame instances.
type BlameOptions struct {
	Fields map[string]any
	Causes []error
}

// NewBlameOptions creates a new BlameOptions instance.
func NewBlameOptions() *BlameOptions {
	return &BlameOptions{
		Fields: make(map[string]any),
		Causes: make([]error, 0),
	}
}

// WithField adds a single field to the Blame.
func WithField(key string, value any) BlameOption {
	return func(opts *BlameOptions) {
		if opts.Fields == nil {
			opts.Fields = make(map[string]any)
		}
		opts.Fields[key] = value
	}
}

// WithFields takes a map[string]any and applies all key-value pairs to BlameOptions.
func WithFields(fields map[string]any) BlameOption {
	return func(opts *BlameOptions) {
		if opts.Fields == nil {
			opts.Fields = make(map[string]any)
		}
		for key, value := range fields {
			opts.Fields[key] = value
		}
	}
}

// WithCauses adds causes to the Blame.
func WithCauses(causes ...error) BlameOption {
	return func(opts *BlameOptions) {
		opts.Causes = causes
	}
}

// ExtendBlameDefinitions adds new BlameDefinitions to an existing slice.
func ExtendBlameDefinitions(initialDefinitions []BlameDefinition, newDefinitions []BlameDefinition) []BlameDefinition {
	// Use the append function to efficiently add new definitions.
	return append(initialDefinitions, newDefinitions...) // ... is crucial to append elements
}

// BuildBlame constructs a Blame object in a generic way.
// If manager is nil, it defaults to getLocalBlameManager().
func BuildBlame(
	errorCode types.ErrorCode,
	fields map[string]any,
	cause error,
	manager *BlameManager,
) Blame {
	if manager == nil {
		manager = getLocalBlameManager()
	}

	options := []BlameOption{}
	if len(fields) > 0 {
		options = append(options, WithFields(fields))
	}
	if cause != nil {
		options = append(options, WithCauses(cause))
	}

	return manager.FetchBlameForError(errorCode, options...)
}

// BlameManagerOption holds configuration
type BlameManagerOption struct {
	LocaleDir       string
	LanguageTag     string
	Bundle          *i18n.Bundle
	ExistingManager *BlameManager
}

// Option defines a function that configures BlameManager
type Option func(*BlameManagerOption)

// WithLocaleDir sets the locale directory
func WithLocaleDir(dir string) Option {
	return func(bw *BlameManagerOption) {
		bw.LocaleDir = dir
	}
}

// WithLanguageTag sets the language tag
func WithLanguageTag(tag string) Option {
	return func(bw *BlameManagerOption) {
		bw.LanguageTag = tag
	}
}

// WithExistingManager sets the existing manager
func WithExistingManager(manager *BlameManager) Option {
	return func(bw *BlameManagerOption) {
		bw.ExistingManager = manager
	}
}

// WithBundle sets the bundle
func WithBundle(bundle *i18n.Bundle) Option {
	return func(bw *BlameManagerOption) {
		bw.Bundle = bundle
	}
}

func NewBlameManagerOption(opts ...Option) *BlameManagerOption {
	bw := &BlameManagerOption{
		LanguageTag: helpers.GetDefaultLanguageTag().String(),
	}
	for _, opt := range opts {
		opt(bw)
	}
	return bw
}
