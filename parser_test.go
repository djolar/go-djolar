package djolar

import (
	"net/url"
	"reflect"
	"testing"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Fatalf("exp not nil, but got nil")
	}
	if p.Metadata.QueryMapping == nil {
		t.Fatalf("exp not nil, but got nil")
	}
	if p.Metadata.DefaultSearch == nil {
		t.Fatalf("exp not nil, but got nil")
	}
	if p.Metadata.ForceSearch == nil {
		t.Fatalf("exp not nil, but got nil")
	}
	if p.Metadata.DefaultOrderBy == nil {
		t.Fatalf("exp not nil, but got nil")
	}
	if p.Metadata.ForceOrderBy == nil {
		t.Fatalf("exp not nil, but got nil")
	}
}

func TestNonExistOperator(t *testing.T) {
	p := NewParser()
	p.Metadata = MetaData{
		QueryMapping: map[string]string{
			"a": "a",
		},
	}
	res, err := p.ParseQuery("q=a__yy__b")
	if err != nil {
		t.Fatalf("exp: no err, got: %v", err)
	}
	if res.WhereClause.Where != "" {
		t.Fail()
	}
	if res.OrderByClause != "" {
		t.Fail()
	}
	if res.HavingClause.Where != "" {
		t.Fail()
	}
	if res.GroupByClause != "" {
		t.Fail()
	}
	if res.SelectClause != "" {
		t.Fail()
	}
}

func TestBuildWithCustomPlaceholder(t *testing.T) {
	p := NewParser()
	p.Metadata = MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
			"j": "j",
			"l": "l",
		},
	}
	p.GetPlaceHolder = func(_ *MetaData, fieldname string) string {
		return "$" + fieldname
	}

	res, err := p.ParseQuery("q=a__eq__1|b__gt__2|c__gte__3|d__lt__4|e__lte__5|f__ico__cde|g__in__[a,b,c]|h__ni__[a,b,c]|i__ne__h|j__sw__k|l__ew__m")
	if err != nil {
		t.Fatalf("exp: no err, got: %v", err)
	}

	exp := "a = $a AND b > $b AND c >= $c AND d < $d AND e <= $e AND LOWER(f) LIKE $f AND g IN ($g) AND h NOT IN ($h) AND i <> $i AND j LIKE $j AND l LIKE $l"
	if res.WhereClause.Where != exp {
		t.Fatalf("exp: %v, got: %v", exp, res.WhereClause.Where)
	}

	args := []interface{}{
		"1", "2", "3", "4", "5", "%cde%", []string{"a", "b", "c"}, []string{"a", "b", "c"}, "h", "k%", "%m",
	}
	if !reflect.DeepEqual(res.WhereClause.Arguments, args) {
		t.Fatalf("exp: %v, got: %v", args, res.WhereClause.Arguments)
	}

	argMaps := map[string]interface{}{
		"a": "1",
		"b": "2",
		"c": "3",
		"d": "4",
		"e": "5",
		"f": "%cde%",
		"g": []string{"a", "b", "c"},
		"h": []string{"a", "b", "c"},
		"i": "h",
		"j": "k%",
		"l": "%m",
	}
	if !reflect.DeepEqual(res.WhereClause.ArgumentMap, argMaps) {
		t.Fatalf("exp: %v, got: %v", args, res.WhereClause.ArgumentMap)
	}
}

func TestBuildQueryWithInvalidQueryString(t *testing.T) {
	md := MetaData{
		QueryMapping:   map[string]string{},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	_, err := p.ParseQuery("hfda=#%^&*()0123")
	if err == nil {
		t.Fatalf("exp: err, got: nil")
	}
}

func TestBuildQueryStringSuccessWithQueryString(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
			"j": "j",
			"l": "l",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	res, err := p.ParseQuery("q=a__eq__1|b__gt__2|c__gte__3|d__lt__4|e__lte__5|d__co__abc|f__ico__cde|g__in__[a,b,c]|h__ni__[a,b,c]|i__ne__h|j__sw__k|l__ew__m")
	if err != nil {
		t.Fatalf("exp: no err, got: %v", err)
	}

	exp := "a = ? AND b > ? AND c >= ? AND d < ? AND e <= ? AND d LIKE ? AND LOWER(f) LIKE ? AND g IN (?) AND h NOT IN (?) AND i <> ? AND j LIKE ? AND l LIKE ?"
	if res.WhereClause.Where != exp {
		t.Fatalf("exp: %v, got: %v", exp, res.WhereClause.Where)
	}

	args := []interface{}{
		"1", "2", "3", "4", "5", "%abc%", "%cde%", []string{"a", "b", "c"}, []string{"a", "b", "c"}, "h", "k%", "%m",
	}
	if !reflect.DeepEqual(res.WhereClause.Arguments, args) {
		t.Fatalf("exp: %v, got: %v", args, res.WhereClause.Arguments)
	}
}

func TestBuildQueryWithInvalidURI(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	_, err := p.ParseURI("q=a__eq__1|b__gt__2|c__gte__3|d__lt__4|e__lte__5|d__co__abc|f__ico__cde|g__in__[a,b,c]|h__ni__[a,b,c]|i__ne__h")
	if err == nil {
		t.Fatalf("exp: err, got: nil")
	}
}

func TestBuildQueryString(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	res, err := p.ParseURI("http://abc.com?q=a__eq__1|b__gt__2|c__gte__3|d__lt__4|e__lte__5|d__co__abc|f__ico__cde|g__in__[a,b,c]|h__ni__[a,b,c]|i__ne__h")
	if err != nil {
		t.Fatalf("exp: no err, got: %v", err)
	}

	exp := "a = ? AND b > ? AND c >= ? AND d < ? AND e <= ? AND d LIKE ? AND LOWER(f) LIKE ? AND g IN (?) AND h NOT IN (?) AND i <> ?"
	if res.WhereClause.Where != exp {
		t.Fatalf("exp: %v, got: %v", exp, res.WhereClause.Where)
	}

	args := []interface{}{
		"1", "2", "3", "4", "5", "%abc%", "%cde%", []string{"a", "b", "c"}, []string{"a", "b", "c"}, "h",
	}
	if !reflect.DeepEqual(res.WhereClause.Arguments, args) {
		t.Fatalf("exp: %v, got: %v", args, res.WhereClause.Arguments)
	}
}

func TestBuildQuerySuccess(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	u, _ := url.ParseRequestURI("http://abc.com?q=a__eq__1|b__gt__2|c__gte__3|d__lt__4|e__lte__5|d__co__abc|f__ico__cde|g__in__[a,b,c]|h__ni__[a,b,c]|i__ne__h")
	res := p.Parse(u.Query())

	exp := "a = ? AND b > ? AND c >= ? AND d < ? AND e <= ? AND d LIKE ? AND LOWER(f) LIKE ? AND g IN (?) AND h NOT IN (?) AND i <> ?"
	if res.WhereClause.Where != exp {
		t.Fatalf("exp: %v, got: %v", exp, res.WhereClause.Where)
	}

	args := []interface{}{
		"1", "2", "3", "4", "5", "%abc%", "%cde%", []string{"a", "b", "c"}, []string{"a", "b", "c"}, "h",
	}
	if !reflect.DeepEqual(res.WhereClause.Arguments, args) {
		t.Fatalf("exp: %v, got: %v", args, res.WhereClause.Arguments)
	}
}

func TestBuildOrderby(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	u, _ := url.ParseRequestURI("http://abc.com?s=-a,b,-c")
	res := p.Parse(u.Query())

	exp := "a DESC,b ASC,c DESC"
	if res.OrderByClause != exp {
		t.Fatalf("exp: %v, got: %v", exp, res.OrderByClause)
	}
}

func TestQueryWithForceSearch(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
		},
		DefaultSearch: map[string]interface{}{},
		ForceSearch: map[string]interface{}{
			"b = ?": "2",
		},
		DefaultOrderBy: []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	u, _ := url.ParseRequestURI("http://abc.com?q=a__eq__1")
	res := p.Parse(u.Query())

	exp := "b = ? AND a = ?"
	if res.WhereClause.Where != exp {
		t.Fatalf("exp: %v, got: %v", exp, res.WhereClause.Where)
	}

	args := []interface{}{
		"2", "1",
	}
	if !reflect.DeepEqual(res.WhereClause.Arguments, args) {
		t.Fatalf("exp: %v, got: %v", args, res.WhereClause.Arguments)
	}
}

func TestQueryWithDefaultSearch(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
		},
		DefaultSearch: map[string]interface{}{
			"b = ?": "2",
		},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	// No query provided

	u, _ := url.ParseRequestURI("http://abc.com")
	res := p.Parse(u.Query())

	exp := "b = ?"
	if res.WhereClause.Where != exp {
		t.Fatalf("exp: %v, got: %v", exp, res.WhereClause.Where)
	}

	args := []interface{}{
		"2",
	}
	if !reflect.DeepEqual(res.WhereClause.Arguments, args) {
		t.Fatalf("exp: %v, got: %v", args, res.WhereClause.Arguments)
	}

	// query provided

	u1, _ := url.ParseRequestURI("http://abc.com?q=a__eq__1")
	res1 := p.Parse(u1.Query())

	exp1 := "a = ?"
	if res1.WhereClause.Where != exp1 {
		t.Fatalf("exp: %v, got: %v", exp1, res1.WhereClause.Where)
	}

	args1 := []interface{}{
		"1",
	}
	if !reflect.DeepEqual(res1.WhereClause.Arguments, args1) {
		t.Fatalf("exp: %v, got: %v", args1, res1.WhereClause.Arguments)
	}
}

func TestQueryWithDefaultOrderby(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
		},
		DefaultSearch: map[string]interface{}{},
		ForceSearch:   map[string]interface{}{},
		DefaultOrderBy: []string{
			"a ASC",
			"b DESC",
		},
	}

	p := &Parser{
		Metadata: md,
	}

	// Without s query

	u, _ := url.ParseRequestURI("http://abc.com")
	res := p.Parse(u.Query())

	exp := "a ASC,b DESC"
	if res.OrderByClause != exp {
		t.Fatalf("exp: %v, got: %v", exp, res.OrderByClause)
	}

	// With s query

	u, _ = url.ParseRequestURI("http://abc.com?s=-c,d")
	res = p.Parse(u.Query())

	exp = "c DESC,d ASC"
	if res.OrderByClause != exp {
		t.Fatalf("exp: %v, got: %v", exp, res.OrderByClause)
	}
}

func TestQueryWithForceOrderby(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
		ForceOrderBy: []string{
			"a ASC",
			"b DESC",
		},
	}

	p := &Parser{
		Metadata: md,
	}

	// Without s query

	u, _ := url.ParseRequestURI("http://abc.com")
	res := p.Parse(u.Query())

	exp := "a ASC,b DESC"
	if res.OrderByClause != exp {
		t.Fatalf("exp: %v, got: %v", exp, res.OrderByClause)
	}

	// With s query

	u, _ = url.ParseRequestURI("http://abc.com?s=-c,d")
	res = p.Parse(u.Query())

	exp = "a ASC,b DESC,c DESC,d ASC"
	if res.OrderByClause != exp {
		t.Fatalf("exp: %v, got: %v", exp, res.OrderByClause)
	}
}

func TestParseQueryWithMapping(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "age",
			"b": "b",
			"n": "name",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
		ForceOrderBy:   []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	p.GetArgMapKey = func(_ *MetaData, fn string) string {
		return fn
	}

	u, _ := url.ParseRequestURI("http://abc.com?q=a__eq__1|b__ne__2|n__co__peter")
	res := p.Parse(u.Query())

	expWhere := "age = ? AND b <> ? AND name LIKE ?"
	expArgs := []interface{}{
		"1",
		"2",
		"%peter%",
	}
	expArgMap := map[string]interface{}{
		"a": "1",
		"b": "2",
		"n": "%peter%",
	}

	if res.WhereClause.Where != expWhere {
		t.Fatalf("exp: %v, got: %v", expWhere, res.WhereClause.Where)
	}
	if !reflect.DeepEqual(res.WhereClause.Arguments, expArgs) {
		t.Fatalf("exp: %v, got: %v", expArgs, res.WhereClause.Arguments)
	}
	if !reflect.DeepEqual(res.WhereClause.ArgumentMap, expArgMap) {
		t.Fatalf("exp: %v, got: %v", expArgMap, res.WhereClause.ArgumentMap)
	}
}

func TestParseQueryWithMappingAndDefaultGetMapKey(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "age",
			"b": "b",
			"n": "name",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
		ForceOrderBy:   []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	u, _ := url.ParseRequestURI("http://abc.com?q=a__eq__1|b__ne__2|n__co__peter")
	res := p.Parse(u.Query())

	expWhere := "age = ? AND b <> ? AND name LIKE ?"
	expArgs := []interface{}{
		"1",
		"2",
		"%peter%",
	}
	expArgMap := map[string]interface{}{
		"age":  "1",
		"b":    "2",
		"name": "%peter%",
	}

	if res.WhereClause.Where != expWhere {
		t.Fatalf("exp: %v, got: %v", expWhere, res.WhereClause.Where)
	}
	if !reflect.DeepEqual(res.WhereClause.Arguments, expArgs) {
		t.Fatalf("exp: %v, got: %v", expArgs, res.WhereClause.Arguments)
	}
	if !reflect.DeepEqual(res.WhereClause.ArgumentMap, expArgMap) {
		t.Fatalf("exp: %v, got: %v", expArgMap, res.WhereClause.ArgumentMap)
	}
}

func TestParserWithoutMetadataDef(t *testing.T) {
	p := &Parser{}

	u, _ := url.ParseRequestURI("http://abc.com?q=a__eq__1|b__gt__2|c__gte__3|d__lt__4|e__lte__5|d__co__abc|f__ico__cde|g__in__[a,b,c]|h__ni__[a,b,c]|i__ne__h|j__xx_j")
	res := p.Parse(u.Query())

	expWhere := ""
	expArgs := []interface{}{}

	if res.WhereClause.Where != expWhere {
		t.Fatalf("exp: %v, got: %v", expWhere, res.WhereClause.Where)
	}
	if !reflect.DeepEqual(res.WhereClause.Arguments, expArgs) {
		t.Fatalf("exp: %v, got: %v", expArgs, res.WhereClause.Arguments)
	}

	u, _ = url.ParseRequestURI("http://abc.com")
	res = p.Parse(u.Query())

	expWhere = ""
	expArgs = []interface{}{}

	if res.WhereClause.Where != expWhere {
		t.Fatalf("exp: %v, got: %v", expWhere, res.WhereClause.Where)
	}
	if !reflect.DeepEqual(res.WhereClause.Arguments, expArgs) {
		t.Fatalf("exp: %v, got: %v", expArgs, res.WhereClause.Arguments)
	}
}

func TestParseGroupBySuccess(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	u, _ := url.ParseRequestURI("http://abc.com?g=a,b,c,d")
	res := p.Parse(u.Query())

	exp := "a,b,c"
	if !reflect.DeepEqual(res.GroupByClause, exp) {
		t.Fatalf("exp: %v, got: %v", exp, res.GroupByClause)
	}
}

func TestParseSelectClauseSuccess(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := NewParser()
	p.Metadata = md

	u, _ := url.ParseRequestURI("http://abc.com?f=a__sum,b,c__count,d__min,e__max,f__avg")
	res := p.Parse(u.Query())

	exp := "SUM(a) AS a__sum,b,COUNT(c) AS c__count,MIN(d) AS d__min,MAX(e) AS e__max,AVG(f) AS f__avg"
	if !reflect.DeepEqual(res.SelectClause, exp) {
		t.Fatalf("exp: %v, got: %v", exp, res.SelectClause)
	}
}

func TestParseSelectCustomAggregateFn(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
		AggregateFunctions: map[string]string{
			"fn1": "FN1",
			"fn2": "FN2",
		},
	}

	p := NewParser()
	p.Metadata = md

	u, _ := url.ParseRequestURI("http://abc.com?f=a__fn1,b,c__fn2")
	res := p.Parse(u.Query())

	exp := "FN1(a) AS a__fn1,b,FN2(c) AS c__fn2"
	if !reflect.DeepEqual(res.SelectClause, exp) {
		t.Fatalf("exp: %v, got: %v", exp, res.SelectClause)
	}
}

func TestParseHavingClauseSuccess(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := NewParser()
	p.Metadata = md

	u, _ := url.ParseRequestURI("http://abc.com?h=a__sum__lt__1|c__count__eq__0")
	res := p.Parse(u.Query())

	exp := "SUM(a) < ? AND COUNT(c) = ?"
	if !reflect.DeepEqual(res.HavingClause.Where, exp) {
		t.Fatalf("exp: %v, got: %v", exp, res.HavingClause.Where)
	}
	exparg := []interface{}{"1", "0"}
	if !reflect.DeepEqual(res.HavingClause.Arguments, exparg) {
		t.Fatalf("exp: %v, got: %v", exparg, res.HavingClause.Arguments)
	}
	expMap := map[string]interface{}{"a__sum": "1", "c__count": "0"}
	if !reflect.DeepEqual(res.HavingClause.ArgumentMap, expMap) {
		t.Fatalf("exp: %v, got: %v", expMap, res.HavingClause.ArgumentMap)
	}
}

func TestParseHavingClauseWithCustomAggregrateFunc(t *testing.T) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
		AggregateFunctions: map[string]string{
			"fn1": "FN1",
			"fn2": "FN2",
		},
	}

	p := NewParser()
	p.Metadata = md

	u, _ := url.ParseRequestURI("http://abc.com?h=a__fn1__lt__1|c__fn2__eq__0")
	res := p.Parse(u.Query())

	exp := "FN1(a) < ? AND FN2(c) = ?"
	if !reflect.DeepEqual(res.HavingClause.Where, exp) {
		t.Fatalf("exp: %v, got: %v", exp, res.HavingClause.Where)
	}
	exparg := []interface{}{"1", "0"}
	if !reflect.DeepEqual(res.HavingClause.Arguments, exparg) {
		t.Fatalf("exp: %v, got: %v", exparg, res.HavingClause.Arguments)
	}
	expMap := map[string]interface{}{"a__fn1": "1", "c__fn2": "0"}
	if !reflect.DeepEqual(res.HavingClause.ArgumentMap, expMap) {
		t.Fatalf("exp: %v, got: %v", expMap, res.HavingClause.ArgumentMap)
	}
}

func TestParseHavingWithNonExistField(t *testing.T) {
	md := MetaData{
		QueryMapping:   map[string]string{},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := NewParser()
	p.Metadata = md

	u, _ := url.ParseRequestURI("http://abc.com?h=a__sum__lt__1|c__count__eq__0")
	res := p.Parse(u.Query())

	exp := ""
	if !reflect.DeepEqual(res.HavingClause.Where, exp) {
		t.Fatalf("exp: %v, got: %v", exp, res.HavingClause.Where)
	}
	exparg := []interface{}{}
	if !reflect.DeepEqual(res.HavingClause.Arguments, exparg) {
		t.Fatalf("exp: %v, got: %v", exparg, res.HavingClause.Arguments)
	}
	expMap := map[string]interface{}{}
	if !reflect.DeepEqual(res.HavingClause.ArgumentMap, expMap) {
		t.Fatalf("exp: %v, got: %v", expMap, res.HavingClause.ArgumentMap)
	}
}

func TestParseQueryDefaultOrderByEmpty(t *testing.T) {
	p := NewParser()
	p.Metadata.DefaultOrderBy = []string{
		"id DESC",
		"created_at ASC",
	}
	res, _ := p.ParseQuery("s=")
	if res.OrderByClause != "id DESC,created_at ASC" {
		t.Fatalf("exp: %v, got: %v(length: %d)", "id DESC,created_at ASC", res.OrderByClause, len(res.OrderByClause))
	}
}

func TestParseQueryDefaultSearchEmpty(t *testing.T) {
	p := NewParser()
	p.Metadata.DefaultSearch = map[string]interface{}{
		"id = ?": 1,
	}
	res, _ := p.ParseQuery("q=")
	if res.WhereClause.Where != "id = ?" {
		t.Fatalf("exp: %v, got: %v(length: %d)", "id = ?", res.WhereClause.Where, len(res.WhereClause.Where))
	}
}

func TestParseDatetime(t *testing.T) {
	p := NewParser()
	p.Metadata = MetaData{
		QueryMapping: map[string]string{
			"created_at_from": "created_at",
			"created_at_to":   "created_at",
		},
	}

	res, err := p.ParseURI("http://a/v1/a/orders?limit=20&offset=0&q=created_at_from__gte__2021-01-11%2000%3A00%7Ccreated_at_to__lte__2021-01-11%2023%3A59")
	if err != nil {
		t.Fatal(err)
	}

	if res.WhereClause.Where != "created_at >= ? AND created_at <= ?" {
		t.Fatal(res.WhereClause.Where)
	}

	if len(res.WhereClause.Arguments) != 2 {
		t.Fatal(res.WhereClause.Arguments...)
	}

	first := res.WhereClause.Arguments[0].(string)
	second := res.WhereClause.Arguments[1].(string)
	if first != "2021-01-11 00:00" {
		t.Fatal(res.WhereClause.Arguments...)
	}
	if second != "2021-01-11 23:59" {
		t.Fatal(res.WhereClause.Arguments...)
	}
}

func BenchmarkParser(b *testing.B) {
	md := MetaData{
		QueryMapping: map[string]string{
			"a": "a",
			"b": "b",
			"c": "c",
			"d": "d",
			"e": "e",
			"f": "f",
			"g": "g",
			"h": "h",
			"i": "i",
			"j": "j",
			"l": "l",
		},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	for i := 0; i < b.N; i++ {
		p.ParseQuery("q=a__eq__1|b__gt__2|c__gte__3|d__lt__4|e__lte__5|d__co__abc|f__ico__cde|g__in__[a,b,c]|h__ni__[a,b,c]|i__ne__h|j__sw__k|l__ew__m")
	}
}
