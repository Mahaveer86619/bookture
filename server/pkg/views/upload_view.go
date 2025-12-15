package views

type UploadVolumeResponse struct {
	VolumeID string `json:"volume_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}
