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
u, _ := url.ParseRequestURI(urlString)
res := p.Parse(u.Query())
db.Where(res.WhereClause.Where, res.WhereClause.ArgumentMap).Sort(res.OrderByClause).Find(&users)
```

## Criteria 1

Filter user with `name` contains `enix` and `age` above `18` years

Frontend:

```
http://www.example.com/v1/users?q=n__co__enix
```

## Criteria 2

Filter user with `name` contain equal `enix` and `age` above `18` years

Frontend:

```
http://www.example.com/v1/users?q=n__eq__enix
```
