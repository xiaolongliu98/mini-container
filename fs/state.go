package fs

var (
	StateName = "state.json"
)

type InstanceState struct {
	Name         string `json:"name"`
	UnionMounted bool   `json:"unionMounted"`
	ImageDir     string `json:"imageDir"`
}
