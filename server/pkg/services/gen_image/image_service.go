package gen_image

import "github.com/Mahaveer86619/bookture/server/pkg/config"

type ImageService interface {
	Init() error
	HealthCheck() error
	GenerateImage(prompt string) (string, error)
}

func NewImageService() ImageService {
	cfg := config.AppConfig

	switch cfg.IMAGE_PROVIDER {
	case "dummy":
		return &DummyImageService{}
	case "hugging-face":
		return &HuggingFaceDiffusersService{}
	// Case "nano-banana": return &NanoBananaService{...}
	default:
		return &DummyImageService{}
	}
}
