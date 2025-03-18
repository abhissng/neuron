package blame

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
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

// BlameWrapper is a wrapper around the blame definitions.
type BlameWrapper struct {
	BlameDefinitions map[types.ErrorCode]Blame
}

// RetrieveBlameCache retrieves a blame definition from the cache.
func (bw *BlameWrapper) RetrieveBlameCache(errorCode types.ErrorCode) Blame {
	if cache, ok := bw.BlameDefinitions[errorCode]; ok {
		return cache
	}
	return NewBasicBlame(types.ErrorCode(errorCode))
}

// FetchBlameForError fetches a blame definition for the given error code.
func (bw *BlameWrapper) FetchBlameForError(errorCode types.ErrorCode, opts ...BlameOption) Blame {
	return bw.RetrieveBlameCache(errorCode).EmptyCause().Wrap(opts...)
}

// NewBlameWrapper creates a new BlameWrapper instance.
func NewBlameWrapper(localeDir string, languageTag string) (*BlameWrapper, error) {

	bundle := helpers.NewBundle(helpers.ParseLanguageTag(languageTag))

	err := InitLocalBlameWrapper(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise local blame wrapper: %w", err)
	}

	// Load error definitions from JSON file
	file, err := os.Open(filepath.Clean(localeDir))
	if err != nil {
		return nil, fmt.Errorf("failed to open error definitions file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			helpers.Println(constant.ERROR, "Error closing file: ", err)
		}
	}()

	var blameDefinitions []BlameDefinition
	if err := json.NewDecoder(file).Decode(&blameDefinitions); err != nil {
		return nil, fmt.Errorf("failed to decode error definitions: %w", err)
	}

	// Create a map of error definitions
	blameDefinitionsMap := make(map[types.ErrorCode]Blame)
	for index, def := range blameDefinitions {
		if helpers.IsEmpty(def.StatusCode) {
			def.StatusCode = helpers.GenerateStatusCode(helpers.GetServiceName(), 100+index)
		}
		blameDefinitionsMap[types.ErrorCode(def.Code)] =
			NewBlame(def.StatusCode, types.ErrorCode(def.Code), def.Message, def.Description).
				WithComponent(types.ComponentErrorType(def.Component)).
				WithResponseType(types.ResponseErrorType(def.ResponseType)).
				WithBundle(bundle)
	}

	return &BlameWrapper{
		BlameDefinitions: blameDefinitionsMap,
		// Bundle:           i18n.NewBundle(language),
	}, nil
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
