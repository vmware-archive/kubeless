package function

type Status struct {
	CurrentVersion string `json:"currentVersion"`
	TargetVersion  string `json:"targetVersion"`
}

func (s *Status) upgradeVersionTo(v string) {
	s.TargetVersion = v
}

func (s *Status) setVersion(v string) {
	s.TargetVersion = ""
	s.CurrentVersion = v
}
