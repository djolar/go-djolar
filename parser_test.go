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

func TestBuildWithCustomPlaceholder(t *testing.T) {
	p := NewParser()
	p.GetPlaceHolder = func(_ *MetaData, fieldname string) string {
		return "$" + fieldname
	}

	res, err := p.ParseQuery("q=a__eq__1|b__gt__2|c__gte__3|d__lt__4|e__lte__5|f__ico__cde|g__in__[a,b,c]|h__ni__[a,b,c]|i__ne__h")
	if err != nil {
		t.Fatalf("exp: no err, got: %v", err)
	}

	exp := "a = $a AND b > $b AND c >= $c AND d < $d AND e <= $e AND LOWER(f) LIKE $f AND g IN ($g) AND h NOT IN ($h) AND i <> $i"
	if res.WhereClause.Where != exp {
		t.Fatalf("exp: %v, got: %v", exp, res.WhereClause.Where)
	}

	args := []interface{}{
		"1", "2", "3", "4", "5", "%cde%", []string{"a", "b", "c"}, []string{"a", "b", "c"}, "h",
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
		QueryMapping:   map[string]string{},
		DefaultSearch:  map[string]interface{}{},
		ForceSearch:    map[string]interface{}{},
		DefaultOrderBy: []string{},
	}

	p := &Parser{
		Metadata: md,
	}

	res, err := p.ParseQuery("q=a__eq__1|b__gt__2|c__gte__3|d__lt__4|e__lte__5|d__co__abc|f__ico__cde|g__in__[a,b,c]|h__ni__[a,b,c]|i__ne__h")
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

func TestBuildQueryWithInvalidURI(t *testing.T) {
	md := MetaData{
		QueryMapping:   map[string]string{},
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
		QueryMapping:   map[string]string{},
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
		QueryMapping:   map[string]string{},
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
		QueryMapping:   map[string]string{},
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
		QueryMapping:  map[string]string{},
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
		QueryMapping: map[string]string{},
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
		QueryMapping:  map[string]string{},
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
		QueryMapping:   map[string]string{},
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

	u, _ := url.ParseRequestURI("http://abc.com?q=a__eq__1|b__ne__2")
	res := p.Parse(u.Query())

	expWhere := "a = ? AND b <> ?"
	expArgs := []interface{}{
		"1",
		"2",
	}

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
