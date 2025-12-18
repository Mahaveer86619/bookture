package parser

import (
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/db"
	"github.com/Mahaveer86619/bookture/server/pkg/services/gen_image"
	"github.com/Mahaveer86619/bookture/server/pkg/services/llm"
	"gorm.io/gorm"
)

type ParserService struct {
	db         *gorm.DB
	llm        llm.LLMService
	imageGen   gen_image.ImageService
	maxRetries int
	retryDelay time.Duration
}

func NewParserService(llmService llm.LLMService, imageService gen_image.ImageService) *ParserService {
	return &ParserService{
		db:         db.GetBooktureDB().DB,
		llm:        llmService,
		imageGen:   imageService,
		maxRetries: 3,
		retryDelay: 5 * time.Second,
	}
}
