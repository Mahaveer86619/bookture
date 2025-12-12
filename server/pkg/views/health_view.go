package views

type HealthView struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewHealthView(status, message string) HealthView {
	return HealthView{
		Status:  status,
		Message: message,
	}
}
