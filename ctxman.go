package ctxman

import (
	"strconv"
	"strings"

	"gorm.io/gorm"
)

const (
	default_files = 10
	max_fields    = 30
	err_message   = "query param `%s` no reconocido, por favor revise la documentación, o pruebe usar notación CamelCase"
)

type MapFuncs map[string]func(tx *gorm.DB) *gorm.DB
type Omiter interface {
	OmitFiels() ([]string, []string)
}
type QueryParamer interface {
	QueryParam(name string) string
}
type parms struct {
	OffSet int    `query:"offset"`
	Limit  int    `query:"limit"`
	Omit   string `query:"omit"`
}
type Ctxx interface {
	/**
	  WithOmiter recibe como parametro la implementacion de:

	  	type Omiter interface {
	  			//return (parametros omitibles, preloads)
	  			OmitFiels() ([]string, []string)
	  	}
	*/
	WithOmiter(o Omiter) Ctxx
	AddCustomPreloadFuns(fns MapFuncs)
	// SimpleGORM ideal para consultas de una fila, offset 0  y limit en 1
	SimpleGORM(conn *gorm.DB, preloads ...string) *gorm.DB
	// FormatGORM prepara offset y limit segun parametors
	FormatGORM(conn *gorm.DB, preloads ...string) *gorm.DB
}
type ctxx struct {
	/// list of omits
	omitfiels        []string
	fieldsForOmit    map[string]struct{}
	fieldsForPreload map[string]struct{}
	offset           int
	limit            int
	preloadfunctions MapFuncs
}

// Newctxx prepara el contexto
func Newctxx(c QueryParamer) Ctxx {
	prms := parms{}
	var err error
	prms.Omit = c.QueryParam("omit")
	prms.OffSet, err = strconv.Atoi(c.QueryParam("offset"))
	if err != nil {
		prms.OffSet = 0
	}
	prms.Limit, err = strconv.Atoi(c.QueryParam("limit"))
	if err != nil {
		prms.Limit = default_files
	}
	if prms.Limit == 0 {
		prms.Limit = default_files
	}
	if prms.Limit > max_fields {
		prms.Limit = max_fields
	}
	var omits []string
	if prms.Omit != "" {
		omits = strings.Split(prms.Omit, ",")
	}
	return &ctxx{
		omitfiels:        omits,
		offset:           prms.OffSet,
		limit:            prms.Limit,
		fieldsForOmit:    make(map[string]struct{}),
		fieldsForPreload: make(map[string]struct{}),
		preloadfunctions: make(MapFuncs),
	}
}

func (c *ctxx) WithOmiter(o Omiter) Ctxx {
	allowsOmits, allowPreloads := o.OmitFiels()
	for _, alp := range allowPreloads {
		c.fieldsForPreload[alp] = struct{}{}
	}
	for _, fil := range c.omitfiels {
		if search(allowsOmits, fil) {
			c.fieldsForOmit[fil] = struct{}{}
			continue
		}
		if search(allowPreloads, fil) {
			delete(c.fieldsForPreload, fil)
			continue
		}
	}
	return c
}

// AddCustomPreloadFuns permite añadir funciones, las cuales se ejecutaran
// antes de realizar el preload, ideal para configurar omits, selects o limits
func (c *ctxx) AddCustomPreloadFuns(fns MapFuncs) {
	for key, f := range fns {
		c.preloadfunctions[key] = f
	}
}

func (c *ctxx) formatGorm(conn *gorm.DB, simple bool, preloads []string) *gorm.DB {
	for _, p := range preloads {
		c.fieldsForOmit[p] = struct{}{}
	}
	var redypreload []string
	for val := range c.fieldsForOmit {
		redypreload = append(redypreload, val)
	}
	var tx *gorm.DB
	if simple {
		tx = conn.Limit(1).Omit(redypreload...)
	} else {
		tx = conn.Limit(c.limit).Offset(c.offset).Omit(redypreload...)
	}
	for p := range c.fieldsForPreload {
		f, ok := c.preloadfunctions[p]
		if ok {
			tx = tx.Preload(p, f)
			continue
		}
		tx = tx.Preload(p)
	}
	return tx
}

// SimpleGORM ideal para consultas de una fila, offset 0  y limit en 1
func (c *ctxx) SimpleGORM(conn *gorm.DB, preloads ...string) *gorm.DB {
	return c.formatGorm(conn, true, preloads)
}

// FormatGORM prepara offset y limit segun parametors
func (c *ctxx) FormatGORM(conn *gorm.DB, preloads ...string) *gorm.DB {
	return c.formatGorm(conn, false, preloads)
}
func search(collection []string, s string) bool {
	for _, item := range collection {
		if item == s {
			return true
		}
	}
	return false
}
