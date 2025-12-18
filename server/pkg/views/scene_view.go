package views

type SceneGenerationResponse struct {
	Scenes []GeneratedScene `json:"scenes"`
}

type GeneratedScene struct {
	SectionNumber   int      `json:"section_number"`
	Summary         string   `json:"summary"`
	ImportanceScore float64  `json:"importance_score"`
	SceneType       string   `json:"scene_type"`
	ImagePrompt     string   `json:"image_prompt"`
	Characters      []string `json:"characters,omitempty"`
	Location        string   `json:"location,omitempty"`
	Mood            string   `json:"mood,omitempty"`
}
