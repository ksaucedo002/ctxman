package ctxman

import "strings"

type Params struct {
	offset    int
	limit     int
	omitfiels []string
}

func (p *Params) OffSet() int       { return p.offset }
func (p *Params) Limit() int        { return p.limit }
func (p *Params) Omitfiels() string { return strings.Join(p.omitfiels, "") }
