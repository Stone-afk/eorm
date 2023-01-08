// Copyright 2021 gotomicro
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package eorm

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestInserter_Values(t *testing.T) {
	type User struct {
		Id        int64 `eorm:"auto_increment,primary_key"`
		FirstName string
		Ctime     uint64
	}
	n := uint64(1000)
	u := &User{
		Id:        12,
		FirstName: "Tom",
		Ctime:     n,
	}
	u1 := &User{
		Id:        13,
		FirstName: "Jerry",
		Ctime:     n,
	}
	db := memoryDB()
	testCases := []CommonTestCase{
		{
			name:    "no examples of values",
			builder: NewInserter[User](db).Values(),
			wantErr: errors.New("插入0行"),
		},
		{
			name:     "single example of values",
			builder:  NewInserter[User](db).Values(u),
			wantSql:  "INSERT INTO `user`(`id`,`first_name`,`ctime`) VALUES(?,?,?);",
			wantArgs: []interface{}{int64(12), "Tom", n},
		},

		{
			name:     "multiple values of same type",
			builder:  NewInserter[User](db).Values(u, u1),
			wantSql:  "INSERT INTO `user`(`id`,`first_name`,`ctime`) VALUES(?,?,?),(?,?,?);",
			wantArgs: []interface{}{int64(12), "Tom", n, int64(13), "Jerry", n},
		},

		{
			name:     "no example of a whole columns",
			builder:  NewInserter[User](db).Columns("Id", "FirstName").Values(u),
			wantSql:  "INSERT INTO `user`(`id`,`first_name`) VALUES(?,?);",
			wantArgs: []interface{}{int64(12), "Tom"},
		},
		{
			name:    "an example with invalid columns",
			builder: NewInserter[User](db).Columns("id", "FirstName").Values(u),
			wantErr: errors.New("eorm: 未知字段 id"),
		},
		{
			name:     "no whole columns and multiple values of same type",
			builder:  NewInserter[User](db).Columns("Id", "FirstName").Values(u, u1),
			wantSql:  "INSERT INTO `user`(`id`,`first_name`) VALUES(?,?),(?,?);",
			wantArgs: []interface{}{int64(12), "Tom", int64(13), "Jerry"},
		},
		{
			name: "ignore pk",
			builder: func() *Inserter[User] {
				res := NewInserter[User](db)
				cols, err := res.NonPKColumns(&User{})
				require.NoError(t, err)
				res.Columns(cols...).Values(u)
				return res
			}(),
			wantSql:  "INSERT INTO `user`(`first_name`,`ctime`) VALUES(?,?);",
			wantArgs: []interface{}{"Tom", uint64(1000)},
		},
		{
			name: "ignore pk multi",
			builder: func() *Inserter[User] {
				res := NewInserter[User](db)
				cols, err := res.NonPKColumns(&User{})
				require.NoError(t, err)
				res.Columns(cols...).Values(u, u1)
				return res
			}(),
			wantSql:  "INSERT INTO `user`(`first_name`,`ctime`) VALUES(?,?),(?,?);",
			wantArgs: []interface{}{"Tom", uint64(1000), "Jerry", uint64(1000)},
		},
	}

	for _, tc := range testCases {
		c := tc
		t.Run(tc.name, func(t *testing.T) {
			q, err := c.builder.Build()
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantSql, q.SQL)
			assert.Equal(t, c.wantArgs, q.Args)
		})
	}
}

func TestInserter_ValuesForCombination(t *testing.T) {
	type BaseEntity struct {
		Id         int64 `eorm:"auto_increment,primary_key"`
		CreateTime uint64
	}
	type User struct {
		BaseEntity
		FirstName string
	}
	n := uint64(1000)
	u := &User{
		FirstName: "Tom",
		BaseEntity: BaseEntity{
			Id:         12,
			CreateTime: n,
		},
	}
	u1 := &User{
		FirstName: "Jerry",
		BaseEntity: BaseEntity{
			Id:         13,
			CreateTime: n,
		},
	}
	db := memoryDB()
	testCases := []CommonTestCase{
		{
			name:    "no examples of values",
			builder: NewInserter[User](db).Values(),
			wantErr: errors.New("插入0行"),
		},
		{
			name:     "single example of values",
			builder:  NewInserter[User](db).Values(u),
			wantSql:  "INSERT INTO `user`(`id`,`create_time`,`first_name`) VALUES(?,?,?);",
			wantArgs: []interface{}{int64(12), n, "Tom"},
		},

		{
			name:     "multiple values of same type",
			builder:  NewInserter[User](db).Values(u, u1),
			wantSql:  "INSERT INTO `user`(`id`,`create_time`,`first_name`) VALUES(?,?,?),(?,?,?);",
			wantArgs: []interface{}{int64(12), n, "Tom", int64(13), n, "Jerry"},
		},

		{
			name:     "no example of a whole columns",
			builder:  NewInserter[User](db).Columns("Id", "FirstName").Values(u),
			wantSql:  "INSERT INTO `user`(`id`,`first_name`) VALUES(?,?);",
			wantArgs: []interface{}{int64(12), "Tom"},
		},

		{
			name:     "no example of a whole columns2",
			builder:  NewInserter[User](db).Columns("FirstName", "Id").Values(u),
			wantSql:  "INSERT INTO `user`(`first_name`,`id`) VALUES(?,?);",
			wantArgs: []interface{}{"Tom", int64(12)},
		},
		{
			name:    "an example with invalid columns",
			builder: NewInserter[User](db).Columns("id", "FirstName").Values(u),
			wantErr: errors.New("eorm: 未知字段 id"),
		},
		{
			name:     "no whole columns and multiple values of same type",
			builder:  NewInserter[User](db).Columns("Id", "FirstName").Values(u, u1),
			wantSql:  "INSERT INTO `user`(`id`,`first_name`) VALUES(?,?),(?,?);",
			wantArgs: []interface{}{int64(12), "Tom", int64(13), "Jerry"},
		},
		{
			name: "ignore pk",
			builder: func() *Inserter[User] {
				res := NewInserter[User](db)
				cols, err := res.NonPKColumns(&User{})
				require.NoError(t, err)
				res.Columns(cols...).Values(u)
				return res
			}(),
			wantSql:  "INSERT INTO `user`(`create_time`,`first_name`) VALUES(?,?);",
			wantArgs: []interface{}{uint64(1000), "Tom"},
		},
		{
			name: "ignore pk multi",
			builder: func() *Inserter[User] {
				res := NewInserter[User](db)
				cols, err := res.NonPKColumns(&User{})
				require.NoError(t, err)
				res.Columns(cols...).Values(u, u1)
				return res
			}(),
			wantSql:  "INSERT INTO `user`(`create_time`,`first_name`) VALUES(?,?),(?,?);",
			wantArgs: []interface{}{uint64(1000), "Tom", uint64(1000), "Jerry"},
		},
	}

	for _, tc := range testCases {
		c := tc
		t.Run(tc.name, func(t *testing.T) {
			q, err := c.builder.Build()
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantSql, q.SQL)
			assert.Equal(t, c.wantArgs, q.Args)
		})
	}
}

func TestInserter_Exec(t *testing.T) {
	orm := memoryDB()
	testCases := []struct {
		name         string
		i            *Inserter[TestModel]
		wantErr      string
		wantAffected int64
	}{
		{
			name:    "invalid query",
			i:       NewInserter[TestModel](orm).Values(),
			wantErr: "插入0行",
		},
		{
			// 表没创建
			name:    "table not exist",
			i:       NewInserter[TestModel](orm).Values(&TestModel{}),
			wantErr: "no such table: test_model",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.i.Exec(context.Background())
			if res.Err() != nil {
				assert.EqualError(t, res.Err(), tc.wantErr)
				return
			}
			assert.Nil(t, tc.wantErr)
			affected, err := res.RowsAffected()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.wantAffected, affected)
		})
	}
}

func ExampleInserter_Build() {
	db := memoryDB()
	query, _ := NewInserter[TestModel](db).Values(&TestModel{
		Id:  1,
		Age: 18,
	}).Build()
	fmt.Printf("case1\n%s", query.string())

	query, _ = NewInserter[TestModel](db).Values(&TestModel{}).Build()
	fmt.Printf("case2\n%s", query.string())

	// Output:
	// case1
	// SQL: INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?);
	// Args: []interface {}{1, "", 18, (*sql.NullString)(nil)}
	// case2
	// SQL: INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?);
	// Args: []interface {}{0, "", 0, (*sql.NullString)(nil)}
}

func ExampleInserter_Columns() {
	db := memoryDB()
	query, _ := NewInserter[TestModel](db).Values(&TestModel{
		Id:  1,
		Age: 18,
	}).Columns("Id", "Age").Build()
	fmt.Printf("case1\n%s", query.string())

	query, _ = NewInserter[TestModel](db).Values(&TestModel{
		Id:  1,
		Age: 18,
	}, &TestModel{}, &TestModel{FirstName: "Tom"}).Columns("Id", "Age").Build()
	fmt.Printf("case2\n%s", query.string())

	// Output:
	// case1
	// SQL: INSERT INTO `test_model`(`id`,`age`) VALUES(?,?);
	// Args: []interface {}{1, 18}
	// case2
	// SQL: INSERT INTO `test_model`(`id`,`age`) VALUES(?,?),(?,?),(?,?);
	// Args: []interface {}{1, 18, 0, 0, 0, 0}

}

func ExampleInserter_Values() {
	db := memoryDB()
	query, _ := NewInserter[TestModel](db).Values(&TestModel{
		Id:  1,
		Age: 18,
	}, &TestModel{}).Build()
	fmt.Println(query.string())
	// Output:
	// SQL: INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?),(?,?,?,?);
	// Args: []interface {}{1, "", 18, (*sql.NullString)(nil), 0, "", 0, (*sql.NullString)(nil)}
}

func ExampleNewInserter() {
	db := memoryDB()
	tm := &TestModel{}
	query, _ := NewInserter[TestModel](db).Values(tm).Build()
	fmt.Printf("SQL: %s", query.SQL)
	// Output:
	// SQL: INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?);
}
