# go-djolar

An simple library to connect backend with frontend for CURD application (with [gorm](https://github.com/jinzhu/gorm))

## Frontend-backend Query Mapping

For example, we have a model name `User`

``` go
type User struct {
    Name string
    Age int
}
```

To map frontend field to backend model field:

```go
md := MetaData{
    QueryMapping:   map[string]string{
        "n": "name", // match n to Name
        "a": "age",
    },
    DefaultSearch:  map[string]interface{}{
        "age > ?": 12 // if no criteria specify, filter user age above 12
    },
    ForceSearch:    map[string]interface{}{
        "age > ?": 18 // force to filter user age above 18
    },
    DefaultOrderBy: []string{
        "age ASC"
    },
}

p := &Parser{
    Metadata: md,
}
```

Backend Filter logic

```go
var users []User
res := parser.Parse(r.URL.Query())
if len(res.WhereClause.Where) > 0 {
    db = db.Where(res.WhereClause.Where, res.WhereClause.Arguments...)
}
if len(res.OrderByClause) > 0 {
    db = db.Order(res.OrderByClause)
}
db.Find(&users)
```

### Criteria 1

Filter user with `name` contains `enix` and `age` above `18` years

Frontend:

```
http://www.example.com/v1/users?q=n__co__enix
```

### Criteria 2

Filter user with `name` contain equal `enix` and `age` above `18` years

Frontend:

```
http://www.example.com/v1/users?q=n__eq__enix
```


## Benchmark

```
cpu: Intel(R) Core(TM) i5-5250U CPU @ 1.60GHz
BenchmarkParser-4   	   40946	     35032 ns/op	    4982 B/op	     105 allocs/op
```