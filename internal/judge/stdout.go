package judge

const StdoutLimitExceededMessage = "STDOUT LIMIT EXCEEDED"

type StdoutCase struct {
	Index  int    `json:"index"`
	Stdout string `json:"stdout"`
}

type StdoutReport struct {
	Cases []StdoutCase `json:"cases"`
}
