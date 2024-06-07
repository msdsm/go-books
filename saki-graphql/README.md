# Goで学ぶGraphQLサーバーサイド入門
- https://zenn.dev/hsaki/books/golang-graphql

## GraphQLとは
- metaによって開発されたWeb APIを開発するためのクエリ言語
- クライアントが必要なデータの構造を定義することができて、定義したものと同じ構造のレスポンスがサーバーから返ってくる

## GraphQLサーバーを動かす
- `gqlgen init`を実行するとデフォルトでTodoアプリが作成される
```
.
├─ graph
│   ├─ generated.go # リゾルバをサーバーで稼働させるためのコアロジック部分(編集しない)
│   ├─ model
│   │   └─ models_gen.go # GraphQLのスキーマオブジェクトがGoの構造体として定義される
│   ├─ resolver.go # ルートリゾルバ構造体の定義
│   ├─ schema.graphqls # GraphQLスキーマ定義(コード自動生成するまにこれをかく)
│   └─ schema.resolvers.go # ビジネスロジックを実装するリゾルバコードが配置(ここに処理を実装していく)
├─ gqlgen.yml # gqlgenの設定ファイル
├─ server.go # サーバーエントリポイント
├─ go.mod
└─ go.sum
```
- graphqlでスキーマを以下のように定義
```graphql
type Todo {
  id: ID!
  text: String!
  done: Boolean!
  user: User!
}

type User {
  id: ID!
  name: String!
}

type Query {
  todos: [Todo!]!
}

input NewTodo {
  text: String!
  userId: String!
}

type Mutation {
  createTodo(input: NewTodo!): Todo!
}
```
- graphQLにはID型、Int型、String型、Boolean型がスカラ型として定義されている
- また、オブジェクト型もあり、上記のファイルの例では`User`がそう
- エクスクラメーションマークは非nullを意味
- graphqlファイルでは`type`を用いてテーブルを定義する
- query, mutationを定義する必要もある
  - query : DBに影響しない処理のこと(ここではtodosというクエリで全件取得する)
  - mutation : DBに影響する処理のこと(ここではcreateTodoというmutationで新しくtodoを作成する)
- このgraphqlファイルをもとに、`model/models_gen.go`に構造体が自動生成される
```go
package model

type Mutation struct {
}

type NewTodo struct {
	Text   string `json:"text"`
	UserID string `json:"userId"`
}

type Query struct {
}

type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
	User *User  `json:"user"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
```
- `resolver.go`では`Resolver`構造体が定義される
```go
package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{}
```
- `schema.resolvers.go`で、この`Resolver`型のポインタ型として`MutationResolver`,`QueryResolver`型が定義されている
- この`MutationResolver`,`QueryResolver`をレシーバとして`CreateTodo`,`Todos`メソッドが生成されている
  - graphqlファイルで定義したものの先頭を大文字にされる
- この中身を実装すればよい
```go
package graph

import (
	"context"
	"my_gql_server/graph/model"
)

// CreateTodo is the resolver for the createTodo field.
func (r *mutationResolver) CreateTodo(ctx context.Context, input model.NewTodo) (*model.Todo, error) {
	// panic(fmt.Errorf("not implemented: CreateTodo - createTodo"))
}

// Todos is the resolver for the todos field.
func (r *queryResolver) Todos(ctx context.Context) ([]*model.Todo, error) {
	// panic(fmt.Errorf("not implemented: Todos - todos"))
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
```
- `go run ./server.go`で起動できる
- http//localhost:8080/をブラウザで開くとクエリ実行のためのPlaygroundにアクセスできる
- クエリは以下のようにquery, mutationをたたける
  - query, mutationを最初に明示
  - その中のメソッド名を明示
  - 引数がある場合はinputで明示
  - 最後にレスポンスとして返してほしい構造体の型を定義
```
query {
  todos {
    id
    text
    done
    user {
      name
    }
  }
}
mutation {
  createTodo(input: {
    text: "test-create-todo"
    userId: "test-user-id"
  }){
    id
    text
    done
    user {
      id
      name
    }
  }
}
```