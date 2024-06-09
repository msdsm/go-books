package services

import (
	"context"
	"graphql-sample/graph/model"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

func New(exec boil.ContextExecutor) Services {
	return &services{
		userService: &userService{exec: exec},
	}
}

// すべてのテーブルに対応するサービス構造体をまとめたもの
type Services interface {
	UserService
	// 今後ここに追加
}

type services struct {
	*userService
}

type UserService interface {
	GetUserByName(ctx context.Context, name string) (*model.User, error)
}
