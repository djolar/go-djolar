package djolar

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

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
}

// PlaceHolderFunc get place holder for given field name
type PlaceHolderFunc func(fieldname string) string

// ArgMapKeyFunc get argment map key
type ArgMapKeyFunc PlaceHolderFunc

var defaultPlaceHolderFunc = func(_ string) string {
	return "?"
}
var defaultArgMapFunc = func(fieldname string) string {
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
		WhereClause: &WhereClause{},
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
	if paramQ, ok := query["q"]; ok && len(paramQ) >= 1 {
		qVal := paramQ[0]
		for _, field := range strings.Split(qVal, "|") {
			// Case-insensitive Contain
			pattern := regexp.MustCompile("(\\w+)__ico__(\\S+)")
			matches := pattern.FindStringSubmatch(field)
			if len(matches) == 3 {
				fn := p.getFieldName(matches[1])
				ph := p.GetPlaceHolder(fn)
				val := fmt.Sprintf("%%%s%%", strings.ToLower(matches[2]))
				where = append(where, fmt.Sprintf("LOWER(%s) LIKE %s", fn, ph))
				args = append(args, val)
				argMap[p.GetArgMapKey(fn)] = val
				continue
			}

			// Contain
			pattern = regexp.MustCompile("(\\w+)__co__(\\S+)")
			matches = pattern.FindStringSubmatch(field)
			if len(matches) == 3 {
				fn := p.getFieldName(matches[1])
				ph := p.GetPlaceHolder(fn)
				val := fmt.Sprintf("%%%s%%", matches[2])
				where = append(where, fmt.Sprintf("%s LIKE %s", fn, ph))
				args = append(args, val)
				argMap[p.GetArgMapKey(fn)] = val
				continue
			}

			// Equal
			pattern = regexp.MustCompile("(\\w+)__eq__(\\S+)")
			matches = pattern.FindStringSubmatch(field)
			if len(matches) == 3 {
				fn := p.getFieldName(matches[1])
				ph := p.GetPlaceHolder(fn)
				where = append(where, fmt.Sprintf("%s = %s", fn, ph))
				args = append(args, matches[2])
				argMap[p.GetArgMapKey(fn)] = matches[2]
				continue
			}

			// Not Equal
			pattern = regexp.MustCompile("(\\w+)__ne__(\\S+)")
			matches = pattern.FindStringSubmatch(field)
			if len(matches) == 3 {
				fn := p.getFieldName(matches[1])
				ph := p.GetPlaceHolder(fn)
				where = append(where, fmt.Sprintf("%s <> %s", fn, ph))
				args = append(args, matches[2])
				argMap[p.GetArgMapKey(fn)] = matches[2]
				continue
			}

			// Less than
			pattern = regexp.MustCompile("(\\w+)__lt__(\\S+)")
			matches = pattern.FindStringSubmatch(field)
			if len(matches) == 3 {
				fn := p.getFieldName(matches[1])
				ph := p.GetPlaceHolder(fn)
				where = append(where, fmt.Sprintf("%s < %s", fn, ph))
				args = append(args, matches[2])
				argMap[p.GetArgMapKey(fn)] = matches[2]
				continue
			}

			// Less than or equal
			pattern = regexp.MustCompile("(\\w+)__lte__(\\S+)")
			matches = pattern.FindStringSubmatch(field)
			if len(matches) == 3 {
				fn := p.getFieldName(matches[1])
				ph := p.GetPlaceHolder(fn)
				where = append(where, fmt.Sprintf("%s <= %s", fn, ph))
				args = append(args, matches[2])
				argMap[p.GetArgMapKey(fn)] = matches[2]
				continue
			}

			// Greater than
			pattern = regexp.MustCompile("(\\w+)__gt__(\\S+)")
			matches = pattern.FindStringSubmatch(field)
			if len(matches) == 3 {
				fn := p.getFieldName(matches[1])
				ph := p.GetPlaceHolder(fn)
				where = append(where, fmt.Sprintf("%s > %s", fn, ph))
				args = append(args, matches[2])
				argMap[p.GetArgMapKey(fn)] = matches[2]
				continue
			}

			// Greater than or equal
			pattern = regexp.MustCompile("(\\w+)__gte__(\\S+)")
			matches = pattern.FindStringSubmatch(field)
			if len(matches) == 3 {
				fn := p.getFieldName(matches[1])
				ph := p.GetPlaceHolder(fn)
				where = append(where, fmt.Sprintf("%s >= %s", fn, ph))
				args = append(args, matches[2])
				argMap[p.GetArgMapKey(fn)] = matches[2]
				continue
			}

			// IN operator
			pattern = regexp.MustCompile("(\\w+)__in__\\[(\\S+)\\]")
			matches = pattern.FindStringSubmatch(field)
			if len(matches) == 3 {
				fn := p.getFieldName(matches[1])
				ph := p.GetPlaceHolder(fn)
				val := strings.Split(matches[2], ",")
				where = append(where, fmt.Sprintf("%s IN (%s)", fn, ph))
				args = append(args, val)
				argMap[p.GetArgMapKey(fn)] = val
				continue
			}

			// NOT IN operator
			pattern = regexp.MustCompile("(\\w+)__ni__\\[(\\S+)\\]")
			matches = pattern.FindStringSubmatch(field)
			if len(matches) == 3 {
				fn := p.getFieldName(matches[1])
				ph := p.GetPlaceHolder(fn)
				val := strings.Split(matches[2], ",")
				where = append(where, fmt.Sprintf("%s NOT IN (%s)", fn, ph))
				args = append(args, val)
				argMap[p.GetArgMapKey(fn)] = val
				continue
			}
		}
	} else {
		// apply default search if defined
		for fieldName, value := range p.Metadata.DefaultSearch {
			where = append(where, fieldName)
			args = append(args, value)
			argMap[p.GetArgMapKey(fieldName)] = value
		}
	}

	// Order by

	// Apply force orderby
	orderby = append(orderby, p.Metadata.ForceOrderBy...)

	if paramOrderby, ok := query["s"]; ok && len(paramOrderby) >= 1 {
		// s query param is provided
		orderbyVal := paramOrderby[0]
		orderby = p.buildOrderby(orderbyVal, orderby)
	} else if len(p.Metadata.DefaultOrderBy) != 0 {
		// Apply default order by
		orderby = append(orderby, p.Metadata.DefaultOrderBy...)
	}

	result.WhereClause.Where = strings.Join(where, " AND ")
	result.WhereClause.Arguments = args
	result.WhereClause.ArgumentMap = argMap
	result.OrderByClause = strings.Join(orderby, ",")

	return result
}

func (p *Parser) buildOrderby(param string, orderby []string) []string {
	for _, order := range strings.Split(param, ",") {
		pattern := regexp.MustCompile("(-)(\\S+)")
		matches := pattern.FindStringSubmatch(order)
		if len(matches) == 3 {
			// DESC
			orderby = append(orderby, fmt.Sprintf("%s DESC", p.getFieldName(matches[2])))
		} else {
			// ASC
			orderby = append(orderby, fmt.Sprintf("%s ASC", p.getFieldName(order)))
		}
	}

	return orderby
}

// getFieldName get internal field name from QueryMapping,
// if not found, then return field directly
func (p *Parser) getFieldName(field string) string {
	f, ok := p.Metadata.QueryMapping[field]
	if ok {
		return f
	}
	return field
}
