package config

type AppConstants struct {
	HuggingFaceStableDifussionXLbaseV1 string
}

var AppConst AppConstants

func LoadConstants() {
	appConsts := AppConstants{
		HuggingFaceStableDifussionXLbaseV1: "https://router.huggingface.co/hf-inference/models/stabilityai/",
	}

	AppConst = appConsts
}
