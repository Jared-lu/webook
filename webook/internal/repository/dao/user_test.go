package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	MySQL "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGormUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name      string
		ctx       context.Context
		inputUser User
		wantErr   error
		sqlMock   func(t *testing.T) *sql.DB
	}{
		{
			name: "插入数据库成功",
			ctx:  context.Background(),
			inputUser: User{
				Email: sql.NullString{
					String: "123@qq.com",
					Valid:  true,
				},
			},
			wantErr: nil,
			sqlMock: func(t *testing.T) *sql.DB {
				// 生成出一个用于测试的虚假的数据库
				mockDB, mock, err := sqlmock.New()
				// 数据库返回结果: 被插入记录的主键，受影响行数
				res := sqlmock.NewResult(3, 1)
				// 这里预期是一个正则表达式，所以只需要是INSERT到user的语句就行
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnResult(res)
				require.NoError(t, err)
				return mockDB
			},
		},
		{
			name: "邮箱冲突",
			ctx:  context.Background(),
			inputUser: User{
				Email: sql.NullString{
					String: "123@qq.com",
					Valid:  true,
				},
			},
			wantErr: ErrUserDuplicate,
			sqlMock: func(t *testing.T) *sql.DB {
				// 生成出一个用于测试的虚假的数据库
				mockDB, mock, err := sqlmock.New()
				// 这里预期是一个正则表达式，所以只需要是INSERT到user的语句就行
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(&MySQL.MySQLError{
						Number: 1062,
					})
				require.NoError(t, err)
				return mockDB
			},
		},
		{
			name: "数据库错误",
			ctx:  context.Background(),
			inputUser: User{
				Email: sql.NullString{
					String: "123@qq.com",
					Valid:  true,
				},
			},
			wantErr: errors.New("数据库错误"),
			sqlMock: func(t *testing.T) *sql.DB {
				// 生成出一个用于测试的虚假的数据库
				mockDB, mock, err := sqlmock.New()
				// 这里预期是一个正则表达式，所以只需要是INSERT到user的语句就行
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(errors.New("数据库错误"))
				require.NoError(t, err)
				return mockDB
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB := tc.sqlMock(t)
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   true,
				SkipDefaultTransaction: true,
			})
			require.NoError(t, err)
			dao := NewUserDAO(db)
			err = dao.Insert(tc.ctx, tc.inputUser)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
