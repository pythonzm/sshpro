package root

import (
	"fmt"
	"strings"
)

const (
	SUCCESS = iota
	FAILED
	UNREACHABLE
)

var MsgFlags = map[int]string{
	SUCCESS:     "SUCCESS",
	FAILED:      "FAILED",
	UNREACHABLE: "UNREACHABLE",
}

type Result struct {
	*Connection
	Status string
	Msg    string
}

func (r *Result) ResultStatus(code int) string {
	return MsgFlags[code]
}

func (r *Result) ColorResult() (res string) {
	switch r.Status {
	case "SUCCESS":
		res = Green(fmt.Sprintf("%s | %s => \n%s", r.Connection.Host, r.Status, r.Msg))
	default:
		res = Red(fmt.Sprintf("%s | %s => %s\n", r.Connection.Host, r.Status, r.Msg))
	}
	return
}

func (r *Result) GenResult(e error) Result {
	if strings.Contains(e.Error(), "no route") {
		r.Status = r.ResultStatus(2)
	} else {
		r.Status = r.ResultStatus(1)
	}
	r.Msg = e.Error()
	return *r
}
