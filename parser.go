package djolar

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var defaultAggregateFunctions = map[string]string{
	"sum":   "SUM",
	"count": "COUNT",
	"min":   "MIN",
	"max":   "MAX",
	"avg":   "AVG",
}

// MetaData meta data for djolar search engine
type MetaData struct {
	// used to convert query field to db field
	// QueryMapping example:
	// 		map[string]string{
	// 			"a": "age",
	// 			"n": "name",
	// 			"s": "score",
	// 		}
	QueryMapping map[string]string

	// if not query param is provided, default search is applied
	// DefaultSearch example:
	// 		map[string]interface{}{
	//			"age = ?": 18,
	//			"name LIKE ?": "Peter",
	//			"gender IN (?)": []int{1, 2},
	//		}
	DefaultSearch map[string]interface{}

	// no matter query param is provied or not, force search criteria will be applied
	// ForceSearch format is as the same as DefaultSearch
	ForceSearch map[string]interface{}

	// default order by fields
	// DefaultOrderBy example []string{"age ASC", "name DESC"}
	DefaultOrderBy []string

	// Apply order by to query no matter s query param is provided or not
	// ForceOrderBy example []string{"age ASC", "name DESC"}
	ForceOrderBy []string

	// Aggregate functions
	// eg.,
	// map[string]string{
	//   "sum": "SUM",
	//   "count": "COUNT",
	//   "min": "MIN",
	//   "max": "MAX",
	//   "avg": "AVG",
	// }
	AggregateFunctions map[string]string
}

// PlaceHolderFunc get place holder for given field name
type PlaceHolderFunc func(md *MetaData, fieldname string) string

// ArgMapKeyFunc get argment map key
type ArgMapKeyFunc func(md *MetaData, fieldname string) string

var defaultPlaceHolderFunc = func(_ *MetaData, _ string) string {
	return "?"
}
var defaultArgMapFunc = func(md *MetaData, fieldname string) string {
	if v, ok := md.QueryMapping[fieldname]; ok {
		return v
	}
	return fieldname
}

// Parser djolar search engine parser
type Parser struct {
	Metadata       MetaData
	GetPlaceHolder PlaceHolderFunc
	GetArgMapKey   ArgMapKeyFunc
}

// WhereClause where clause
type WhereClause struct {
	Where       string
	Arguments   []interface{}
	ArgumentMap map[string]interface{}
}

// ParseResult parse query result
type ParseResult struct {
	WhereClause   *WhereClause
	SelectClause  string
	GroupByClause string
	HavingClause  *WhereClause
	OrderByClause string
}

// NewParser create a new parser
func NewParser() *Parser {
	md := MetaData{
		QueryMapping:   map[string]string{},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
		ForceOrderBy:   []string{},
	}

	p := &Parser{
		Metadata:       md,
		GetPlaceHolder: defaultPlaceHolderFunc,
		GetArgMapKey:   defaultArgMapFunc,
	}
	return p
}

// ParseQuery parse with query string
func (p *Parser) ParseQuery(query string) (*ParseResult, error) {
	qv, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	res := p.Parse(qv)
	return res, nil
}

// ParseURI parse given URI string, and extrat djolar compatiable query conditions
// from the `q` parameter.
func (p *Parser) ParseURI(uri string) (*ParseResult, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, err
	}

	res := p.Parse(u.Query())
	return res, nil
}

// Parse parse url query values
func (p *Parser) Parse(query url.Values) *ParseResult {
	args := make([]interface{}, 0)
	argMap := make(map[string]interface{})
	where := make([]string, 0)
	orderby := make([]string, 0)
	result := &ParseResult{
		WhereClause:  &WhereClause{},
		HavingClause: &WhereClause{},
	}

	if p.GetPlaceHolder == nil {
		p.GetPlaceHolder = defaultPlaceHolderFunc
	}
	if p.GetArgMapKey == nil {
		p.GetArgMapKey = defaultArgMapFunc
	}

	// Apply force search if defined
	for fieldName, value := range p.Metadata.ForceSearch {
		where = append(where, fieldName)
		args = append(args, value)
	}

	// Query
	if paramQ, ok := query["q"]; ok && len(paramQ) >= 1 && len(paramQ[0]) > 0 {
		qVal := paramQ[0]
		for _, field := range strings.Split(qVal, "|") {
			col, wh, arg, ok := p.buildWhereClause(field, p.Metadata.QueryMapping)
			if !ok {
				continue
			}
			where = append(where, wh)
			args = append(args, arg)
			argMap[p.GetArgMapKey(&p.Metadata, col)] = arg
		}
	} else {
		// apply default search if defined
		for fieldName, value := range p.Metadata.DefaultSearch {
			where = append(where, fieldName)
			args = append(args, value)
			argMap[p.GetArgMapKey(&p.Metadata, fieldName)] = value
		}
	}
	result.WhereClause.Where = strings.Join(where, " AND ")
	result.WhereClause.Arguments = args
	result.WhereClause.ArgumentMap = argMap

	// Order by

	// Apply force orderby
	orderby = append(orderby, p.Metadata.ForceOrderBy...)
	if paramOrderby, ok := query["s"]; ok && len(paramOrderby) >= 1 && len(paramOrderby[0]) > 0 {
		// s query param is provided
		orderbyVal := paramOrderby[0]
		orderby = p.buildOrderby(orderbyVal, orderby)
	} else if len(p.Metadata.DefaultOrderBy) != 0 {
		// Apply default order by
		orderby = append(orderby, p.Metadata.DefaultOrderBy...)
	}
	result.OrderByClause = strings.Join(orderby, ",")

	// Group by
	// Ex. g=field1,field2
	if paramGroupBy, ok := query["g"]; ok && len(paramGroupBy) > 0 {
		groupBy := p.buildGroupBy(paramGroupBy[0])
		result.GroupByClause = strings.Join(groupBy, ",")
	}

	// Select
	var selectClause []string
	if paramSelect, ok := query["f"]; ok && len(paramSelect) > 0 {
		selectClause = p.buildSelectClause(paramSelect[0])
		result.SelectClause = strings.Join(selectClause, ",")
	}

	// Having clause
	if paramHaving, ok := query["h"]; ok && len(paramHaving) > 0 {
		result.HavingClause = p.buildHavingClause(paramHaving[0])
	}

	return result
}

func (p *Parser) buildWhereClause(field string, queryMapping map[string]string) (colName, where string, arg interface{}, ok bool) {
	// Case-insensitive Contain
	pattern := regexp.MustCompile("(\\w+)__ico__(\\S+)")
	matches := pattern.FindStringSubmatch(field)
	var fn string
	if len(matches) == 3 {
		fn, ok = queryMapping[matches[1]]
		if !ok {
			return "", "", nil, false
		}
		ph := p.GetPlaceHolder(&p.Metadata, matches[1])
		arg = fmt.Sprintf("%%%s%%", strings.ToLower(matches[2]))
		where = fmt.Sprintf("LOWER(%s) LIKE %s", fn, ph)
		colName = matches[1]
		return
	}

	// Contain
	pattern = regexp.MustCompile("(\\w+)__co__(\\S+)")
	matches = pattern.FindStringSubmatch(field)
	if len(matches) == 3 {
		fn, ok = queryMapping[matches[1]]
		if !ok {
			return "", "", nil, false
		}
		ph := p.GetPlaceHolder(&p.Metadata, matches[1])
		arg = fmt.Sprintf("%%%s%%", matches[2])
		where = fmt.Sprintf("%s LIKE %s", fn, ph)
		colName = matches[1]
		return
	}

	// Equal
	pattern = regexp.MustCompile("(\\w+)__eq__(\\S+)")
	matches = pattern.FindStringSubmatch(field)
	if len(matches) == 3 {
		fn, ok = queryMapping[matches[1]]
		if !ok {
			return "", "", nil, false
		}
		ph := p.GetPlaceHolder(&p.Metadata, matches[1])
		where = fmt.Sprintf("%s = %s", fn, ph)
		arg = matches[2]
		colName = matches[1]
		return
	}

	// Not Equal
	pattern = regexp.MustCompile("(\\w+)__ne__(\\S+)")
	matches = pattern.FindStringSubmatch(field)
	if len(matches) == 3 {
		fn, ok = queryMapping[matches[1]]
		if !ok {
			return "", "", nil, false
		}
		ph := p.GetPlaceHolder(&p.Metadata, matches[1])
		where = fmt.Sprintf("%s <> %s", fn, ph)
		arg = matches[2]
		colName = matches[1]
		return
	}

	// Less than
	pattern = regexp.MustCompile("(\\w+)__lt__(\\S+)")
	matches = pattern.FindStringSubmatch(field)
	if len(matches) == 3 {
		fn, ok = queryMapping[matches[1]]
		if !ok {
			return "", "", nil, false
		}
		ph := p.GetPlaceHolder(&p.Metadata, matches[1])
		where = fmt.Sprintf("%s < %s", fn, ph)
		arg = matches[2]
		colName = matches[1]
		return
	}

	// Less than or equal
	pattern = regexp.MustCompile("(\\w+)__lte__(\\S+)")
	matches = pattern.FindStringSubmatch(field)
	if len(matches) == 3 {
		fn, ok = queryMapping[matches[1]]
		if !ok {
			return "", "", nil, false
		}
		ph := p.GetPlaceHolder(&p.Metadata, matches[1])
		where = fmt.Sprintf("%s <= %s", fn, ph)
		arg = matches[2]
		colName = matches[1]
		return
	}

	// Greater than
	pattern = regexp.MustCompile("(\\w+)__gt__(\\S+)")
	matches = pattern.FindStringSubmatch(field)
	if len(matches) == 3 {
		fn, ok = queryMapping[matches[1]]
		if !ok {
			return "", "", nil, false
		}
		ph := p.GetPlaceHolder(&p.Metadata, matches[1])
		where = fmt.Sprintf("%s > %s", fn, ph)
		arg = matches[2]
		colName = matches[1]
		return
	}

	// Greater than or equal
	pattern = regexp.MustCompile("(\\w+)__gte__(\\S+)")
	matches = pattern.FindStringSubmatch(field)
	if len(matches) == 3 {
		fn, ok = queryMapping[matches[1]]
		if !ok {
			return "", "", nil, false
		}
		ph := p.GetPlaceHolder(&p.Metadata, matches[1])
		where = fmt.Sprintf("%s >= %s", fn, ph)
		arg = matches[2]
		colName = matches[1]
		return
	}

	// IN operator
	pattern = regexp.MustCompile("(\\w+)__in__\\[(\\S+)\\]")
	matches = pattern.FindStringSubmatch(field)
	if len(matches) == 3 {
		fn, ok = queryMapping[matches[1]]
		if !ok {
			return "", "", nil, false
		}
		ph := p.GetPlaceHolder(&p.Metadata, matches[1])
		arg = strings.Split(matches[2], ",")
		where = fmt.Sprintf("%s IN (%s)", fn, ph)
		colName = matches[1]
		return
	}

	// NOT IN operator
	pattern = regexp.MustCompile("(\\w+)__ni__\\[(\\S+)\\]")
	matches = pattern.FindStringSubmatch(field)
	if len(matches) == 3 {
		fn, ok = queryMapping[matches[1]]
		if !ok {
			return "", "", nil, false
		}
		ph := p.GetPlaceHolder(&p.Metadata, matches[1])
		arg = strings.Split(matches[2], ",")
		where = fmt.Sprintf("%s NOT IN (%s)", fn, ph)
		colName = matches[1]
		return
	}

	return "", "", nil, false
}

func (p *Parser) buildOrderby(param string, orderby []string) []string {
	for _, order := range strings.Split(param, ",") {
		pattern := regexp.MustCompile("(-)(\\S+)")
		matches := pattern.FindStringSubmatch(order)
		if len(matches) == 3 {
			// DESC
			if field, ok := p.Metadata.QueryMapping[matches[2]]; ok {
				orderby = append(orderby, fmt.Sprintf("%s DESC", field))
			}
		} else {
			// ASC
			if field, ok := p.Metadata.QueryMapping[order]; ok {
				orderby = append(orderby, fmt.Sprintf("%s ASC", field))
			}
		}
	}

	return orderby
}

func (p *Parser) buildGroupBy(param string) []string {
	groupby := make([]string, 0)
	for _, item := range strings.Split(param, ",") {
		if field, ok := p.Metadata.QueryMapping[item]; ok {
			groupby = append(groupby, field)
		}
	}

	return groupby
}

func (p *Parser) buildSelectClause(param string) []string {
	clause := make([]string, 0)

	var aggregrateFns map[string]string
	if p.Metadata.AggregateFunctions == nil {
		aggregrateFns = defaultAggregateFunctions
	} else {
		aggregrateFns = p.Metadata.AggregateFunctions
	}

	for _, item := range strings.Split(param, ",") {
		if field, ok := p.Metadata.QueryMapping[item]; ok {
			clause = append(clause, field)
		} else {
			// check if using aggregate functions
			// loop over all aggregate functions
			for k, fn := range aggregrateFns {
				pattern := regexp.MustCompile(fmt.Sprintf("(\\w+)__%s", k))
				matches := pattern.FindStringSubmatch(item)
				if len(matches) == 2 {
					clause = append(clause, fmt.Sprintf("%s(%s) AS %s", fn, matches[1], item))
				}
			}
		}
	}

	return clause
}

// Build HAVING clause
// eg., h=a__sum__lt__1|b__count__gt__0&f=a__sum
// => [a__sum, lt, 1], [COUNT(b), gt, 0]
//
// Steps:
// 1. check if the column name exist in the SELECT clause
// 2. If not exist, then build the column with aggrgrate function
// 3. go through the where clause building workflow
func (p *Parser) buildHavingClause(param string) *WhereClause {
	whereClause := &WhereClause{}
	args := make([]interface{}, 0)
	argMap := make(map[string]interface{})
	where := make([]string, 0)

	queryMapping := make(map[string]string)
	for k, v := range p.Metadata.QueryMapping {
		queryMapping[k] = v
	}

	var aggregrateFns map[string]string
	if p.Metadata.AggregateFunctions == nil {
		aggregrateFns = defaultAggregateFunctions
	} else {
		aggregrateFns = p.Metadata.AggregateFunctions
	}

	for _, field := range strings.Split(param, "|") {
		for k, fn := range aggregrateFns {
			pattern := regexp.MustCompile(fmt.Sprintf("(\\w+)__%s", k))
			matches := pattern.FindStringSubmatch(field)
			if len(matches) == 2 {
				if fieldName, ok := queryMapping[matches[1]]; ok {
					f := fmt.Sprintf("%s(%s)", fn, fieldName)
					queryMapping[matches[0]] = f
				}
				break
			}
		}

		col, wh, arg, ok := p.buildWhereClause(field, queryMapping)
		if !ok {
			continue
		}
		where = append(where, wh)
		args = append(args, arg)
		argMap[p.GetArgMapKey(&p.Metadata, col)] = arg
	}

	whereClause.ArgumentMap = argMap
	whereClause.Arguments = args
	whereClause.Where = strings.Join(where, " AND ")

	return whereClause
}
