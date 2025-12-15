package enums

type BookStatus string

const (
	Draft      BookStatus = "draft"      // Draft empty book
	Processing BookStatus = "processing" // Ready for parsing job
	Completed  BookStatus = "completed"  // Parsed and volumes ready
)

func (bs BookStatus) ToString() string {
	switch bs {
	case Draft:
		return "draft"
	case Processing:
		return "processing"
	case Completed:
		return "completed"
	default:
		return "unknown"
	}
}
