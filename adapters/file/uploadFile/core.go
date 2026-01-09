package uploadFile

import (
	"errors"
	"sync"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
)

/*
========================================

	Validation Errors

========================================
*/
var (
	ErrFileTooLarge     = errors.New("file exceeds allowed size")
	ErrInvalidMimeType  = errors.New("mime type not allowed")
	ErrInvalidExtension = errors.New("file extension not allowed")
)

type FileRule struct {
	MaxSizeBytes int64
	AllowedMIMEs map[string]struct{}
	AllowedExts  map[string]struct{}
}

type UploadProfile string

const (
	ProfileImage    UploadProfile = "image"
	ProfileVideo    UploadProfile = "video"
	ProfileAudio    UploadProfile = "audio"
	ProfileDocument UploadProfile = "document"
	ProfileMixed    UploadProfile = "mixed"
	ProfileCustom   UploadProfile = "custom"
)

const (
	KB = 1 << 10
	MB = 1 << 20
	GB = 1 << 30
)

var (
	ruleRegistry = map[UploadProfile]*FileRule{}
	registryMu   sync.RWMutex
)

func set(values ...string) map[string]struct{} {
	s := make(map[string]struct{}, len(values))
	for _, v := range values {
		s[v] = struct{}{}
	}
	return s
}

func init() {
	RegisterUploadProfile(ProfileImage, &FileRule{
		MaxSizeBytes: 20 * MB,
		AllowedMIMEs: set(
			"image/jpeg",
			"image/png",
			"image/webp",
		),
		AllowedExts: set(".jpg", ".jpeg", ".png", ".webp"),
	})

	RegisterUploadProfile(ProfileVideo, &FileRule{
		MaxSizeBytes: 5 * GB,
		AllowedMIMEs: set(
			"video/mp4",
			"video/webm",
			"video/x-matroska",
		),
		AllowedExts: set(".mp4", ".webm", ".mkv"),
	})

	RegisterUploadProfile(ProfileAudio, &FileRule{
		MaxSizeBytes: 500 * MB,
		AllowedMIMEs: set(
			"audio/mpeg",
			"audio/wav",
		),
		AllowedExts: set(".mp3", ".wav"),
	})

	RegisterUploadProfile(ProfileDocument, &FileRule{
		MaxSizeBytes: 50 * MB,
		AllowedMIMEs: set(
			"application/pdf",
			"text/plain",
			"text/csv",
			"application/vnd.ms-excel",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		),
		AllowedExts: set(".pdf", ".txt", ".csv", ".xls", ".xlsx"),
	})

	// Mixed = union of all above
	RegisterUploadProfile(ProfileMixed, MergeUploadProfiles(
		ProfileImage,
		ProfileVideo,
		ProfileAudio,
		ProfileDocument,
	))
}

func RegisterUploadProfile(name UploadProfile, rule *FileRule) {
	if rule == nil {
		helpers.Println(constant.ERROR, "RegisterUploadProfile: rule is nil")
		return
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	ruleRegistry[name] = rule
}

func GetUploadProfile(name UploadProfile) (*FileRule, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	rule, ok := ruleRegistry[name]
	return rule, ok
}

func MergeUploadProfiles(profiles ...UploadProfile) *FileRule {
	merged := &FileRule{
		MaxSizeBytes: 0,
		AllowedMIMEs: map[string]struct{}{},
		AllowedExts:  map[string]struct{}{},
	}

	foundAny := false

	for _, p := range profiles {
		rule, ok := GetUploadProfile(p)
		if !ok {
			continue
		}
		foundAny = true

		if rule.MaxSizeBytes > merged.MaxSizeBytes {
			merged.MaxSizeBytes = rule.MaxSizeBytes
		}

		for k := range rule.AllowedMIMEs {
			merged.AllowedMIMEs[k] = struct{}{}
		}
		for k := range rule.AllowedExts {
			merged.AllowedExts[k] = struct{}{}
		}
	}
	if !foundAny {
		helpers.Println(constant.ERROR, "MergeUploadProfiles: no profiles found")
		return nil
	}

	return merged
}

/*
========================================
 Custom Rule Builder
========================================
*/

func NewCustomRule(maxSize int64, mimes []string, exts []string) *FileRule {
	if maxSize <= 0 {
		maxSize = 10 * MB
	}
	return &FileRule{
		MaxSizeBytes: maxSize,
		AllowedMIMEs: set(mimes...),
		AllowedExts:  set(exts...),
	}
}
