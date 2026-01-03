package enums

type BookStatus string

const (
	BookDraft      BookStatus = "draft"
	BookProcessing BookStatus = "processing"
	BookCompleted  BookStatus = "completed"
	BookError      BookStatus = "error"
)

type VolumeStatus string

const (
	VolumeCreated   VolumeStatus = "created"
	VolumeUploaded  VolumeStatus = "uploaded"
	VolumeParsing   VolumeStatus = "parsing"   // Extracting text
	VolumeParsed    VolumeStatus = "parsed"    // Text in DB
	VolumeEnhancing VolumeStatus = "enhancing" // AI analyzing
	VolumeCompleted VolumeStatus = "completed"
	VolumeError     VolumeStatus = "error"
)

func (vs VolumeStatus) CanTransitionTo(next VolumeStatus) bool {
	validTransitions := map[VolumeStatus][]VolumeStatus{
		VolumeCreated:   {VolumeUploaded, VolumeError},
		VolumeUploaded:  {VolumeParsing, VolumeError},
		VolumeParsing:   {VolumeParsed, VolumeError},
		VolumeParsed:    {VolumeEnhancing, VolumeCompleted, VolumeError},
		VolumeEnhancing: {VolumeCompleted, VolumeError},
		VolumeCompleted: {VolumeParsing, VolumeEnhancing}, // Allow re-processing
		VolumeError:     {VolumeParsing, VolumeEnhancing}, // Allow retry
	}

	allowed, exists := validTransitions[vs]
	if !exists {
		return false
	}

	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}
