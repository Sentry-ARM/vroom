package occurrence

import (
	"crypto/md5"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/getsentry/vroom/internal/frame"
	"github.com/getsentry/vroom/internal/platform"
	"github.com/getsentry/vroom/internal/profile"
	"github.com/google/uuid"
)

type (
	EvidenceNameType string
	IssueTitleType   string

	OccurrenceType int

	Evidence struct {
		Name      EvidenceNameType `json:"name"`
		Value     string           `json:"value"`
		Important bool             `json:"important"`
	}

	// Event holds the metadata related to a profile
	Event struct {
		Environment string            `json:"environment"`
		ID          string            `json:"event_id"`
		Platform    platform.Platform `json:"platform"`
		ProjectID   uint64            `json:"project_id"`
		Received    time.Time         `json:"received"`
		Release     string            `json:"release,omitempty"`
		StackTrace  StackTrace        `json:"stacktrace"`
		Tags        map[string]string `json:"tags"`
		Timestamp   time.Time         `json:"timestamp"`
		Transaction string            `json:"transaction,omitempty"`
	}

	// Occurrence represents a potential issue detected
	Occurrence struct {
		DetectionTime   time.Time              `json:"detection_time"`
		Event           Event                  `json:"event"`
		EvidenceData    map[string]interface{} `json:"evidence_data,omitempty"`
		EvidenceDisplay []Evidence             `json:"evidence_display,omitempty"`
		Fingerprint     string                 `json:"fingerprint"`
		ID              string                 `json:"id"`
		IssueTitle      IssueTitleType         `json:"issue_title"`
		Level           string                 `json:"level,omitempty"`
		ResourceID      string                 `json:"resource_id,omitempty"`
		Subtitle        string                 `json:"subtitle"`
		Type            OccurrenceType         `json:"type"`
	}

	StackTrace struct {
		Frames []frame.Frame `json:"frames"`
	}
)

const (
	ProfileBlockedThreadType OccurrenceType = 2000

	EvidenceNamePackage  EvidenceNameType = "Package"
	EvidenceNameFunction EvidenceNameType = "Suspect function"

	IssueTitleBlockingFunctionOnMainThread IssueTitleType = "Blocking function called on the main thread"
)

func NewOccurrence(p profile.Profile, title IssueTitleType, ni nodeInfo) Occurrence {
	t := p.Transaction()
	h := md5.New()
	_, _ = io.WriteString(h, strconv.FormatUint(p.ProjectID(), 10))
	_, _ = io.WriteString(h, string(title))
	_, _ = io.WriteString(h, t.Name)
	_, _ = io.WriteString(h, strconv.Itoa(int(ProfileBlockedThreadType)))
	_, _ = io.WriteString(h, ni.Node.Package)
	_, _ = io.WriteString(h, ni.Node.Name)
	fingerprint := fmt.Sprintf("%x", h.Sum(nil))
	tags := buildOccurrenceTags(p)
	return Occurrence{
		DetectionTime: time.Now().UTC(),
		Event: Event{
			Environment: p.Environment(),
			ID:          p.ID(),
			Platform:    p.Platform(),
			ProjectID:   p.ProjectID(),
			Received:    p.Received(),
			Release:     p.Release(),
			StackTrace:  StackTrace{Frames: ni.StackTrace},
			Tags:        tags,
			Timestamp:   p.Timestamp(),
			Transaction: t.ID,
		},
		EvidenceData: map[string]interface{}{},
		EvidenceDisplay: []Evidence{
			Evidence{
				Name:      EvidenceNameFunction,
				Value:     ni.Node.Name,
				Important: true,
			},
			Evidence{
				Name:  EvidenceNamePackage,
				Value: ni.Node.Package,
			},
		},
		Fingerprint: fingerprint,
		ID:          uuid.New().String(),
		IssueTitle:  title,
		Subtitle:    t.Name,
		Type:        ProfileBlockedThreadType,
	}
}

func buildOccurrenceTags(p profile.Profile) map[string]string {
	pm := p.Metadata()
	tags := map[string]string{
		"device_classification": pm.DeviceClassification,
		"device_locale":         pm.DeviceLocale,
		"device_manufacturer":   pm.DeviceManufacturer,
		"device_model":          pm.DeviceModel,
		"device_os_name":        pm.DeviceOsName,
		"device_os_version":     pm.DeviceOsVersion,
	}

	if pm.DeviceOsBuildNumber != "" {
		tags["device_os_build_number"] = pm.DeviceOsBuildNumber
	}

	return tags
}