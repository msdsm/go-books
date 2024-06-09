package services

import (
	"context"
	"graphql-sample/graph/model"

	"graphql-sample/graph/db"

	"fmt"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type userService struct {
	exec boil.ContextExecutor
}

// db.User -> model.User (SQLBoilerで生成された構造体からgraphQLで生成された構造体に変換)
func convertUser(user *db.User) *model.User {
	return &model.User{
		ID:   user.ID,
		Name: user.Name,
	}
}

func (u *userService) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	fmt.Println("GetUserByName")
	// 1. SQLBoilerで生成されたORMコードを呼び出す
	user, err := db.Users( // from users
		qm.Select(db.UserTableColumns.ID, db.UserTableColumns.Name), // select id, name
		db.UserWhere.Name.EQ(name),                                  // where name = 引数
	).One(ctx, u.exec) // limit 1

	// 2. エラー処理
	if err != nil {
		return nil, err
	}

	// 3. 戻り値の*model.User型を作成
	return convertUser(user), nil
}
