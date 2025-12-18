package gen_image

type DummyImageService struct {
}

func (s *DummyImageService) Init() error {
	return nil
}

func (s *DummyImageService) HealthCheck() error {
	return nil
}

func (s *DummyImageService) GenerateImage(prompt string) (string, error) {
	return "", nil
}
