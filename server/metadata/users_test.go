// Copyright 2022-2023 Tigris Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metadata

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tigrisdata/tigris/errors"
	"github.com/tigrisdata/tigris/server/transaction"
)

var testUserPayload = []byte(`{
	"user_name": "123",
	"last_visit": "123123"
}`)

func initUserTest(t *testing.T) (*UserSubspace, transaction.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u := NewUserStore(&NameRegistry{
		UserSB: "test_user",
	})
	_ = kvStore.DropTable(ctx, u.SubspaceName)

	tm := transaction.NewManager(kvStore)
	tx, err := tm.StartTx(ctx)
	require.NoError(t, err)

	return u, tx
}

func TestUserSubspace(t *testing.T) {
	t.Run("put_error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		u, tx := initUserTest(t)
		defer func() { assert.NoError(t, tx.Rollback(ctx)) }()

		require.Equal(t, errors.InvalidArgument("invalid empty userId"), u.InsertUserMetadata(ctx, tx, 1, User, "", "meta-key-1", testUserPayload))
		require.Equal(t, errors.InvalidArgument("invalid nil payload"), u.InsertUserMetadata(ctx, tx, 1, User, "some-valid-user-id", "meta-key-1", nil))
		require.Equal(t, errors.InvalidArgument("invalid namespace, id must be greater than 0"), u.InsertUserMetadata(ctx, tx, 0, User, "user-id-1", "meta-key-1", testUserPayload))
		_ = kvStore.DropTable(ctx, u.SubspaceName)
	})

	t.Run("put_get_1", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		u, tx := initUserTest(t)
		defer func() { assert.NoError(t, tx.Rollback(ctx)) }()

		require.NoError(t, u.InsertUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1", testUserPayload))
		user, err := u.GetUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1")
		require.NoError(t, err)
		require.Equal(t, testUserPayload, user)

		_ = kvStore.DropTable(ctx, u.SubspaceName)
	})

	t.Run("put_get_2", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		u, tx := initUserTest(t)
		defer func() { assert.NoError(t, tx.Rollback(ctx)) }()

		appPayload := []byte(`{
								"app_id": "123",
								"last_login": "123123"
								}`)

		require.NoError(t, u.InsertUserMetadata(ctx, tx, 1, Application, "app-id-123", "meta-key-1", appPayload))
		user, err := u.GetUserMetadata(ctx, tx, 1, Application, "app-id-123", "meta-key-1")
		require.NoError(t, err)
		require.Equal(t, appPayload, user)

		_ = kvStore.DropTable(ctx, u.SubspaceName)
	})

	t.Run("put_get_update_get", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		u, tx := initUserTest(t)
		defer func() { assert.NoError(t, tx.Rollback(ctx)) }()

		require.NoError(t, u.InsertUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1", testUserPayload))
		user, err := u.GetUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1")
		require.NoError(t, err)
		require.Equal(t, testUserPayload, user)

		updatedUserPayload := []byte(`{
								"user_name": "123",
								"last_visit": "456456"
								}`)
		require.NoError(t, u.UpdateUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1", updatedUserPayload))
		user, err = u.GetUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1")
		require.NoError(t, err)
		require.Equal(t, updatedUserPayload, user)

		_ = kvStore.DropTable(ctx, u.SubspaceName)
	})

	t.Run("put_get_deleteuser_get", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		u, tx := initUserTest(t)
		defer func() { assert.NoError(t, tx.Rollback(ctx)) }()

		metaVal1 := []byte(`{
								"meta-key-1": "val1",
								"meta-key-2": "val2"
								}`)

		metaVal2 := []byte(`{
								"meta-key-1": "val1",
								"meta-key-2": "val2"
								}`)

		require.NoError(t, u.InsertUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1", metaVal1))
		require.NoError(t, u.InsertUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-2", metaVal2))

		metaValRetrieved, err := u.GetUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1")
		require.NoError(t, err)
		require.Equal(t, metaVal1, metaValRetrieved)

		metaValRetrieved, err = u.GetUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-2")
		require.NoError(t, err)
		require.Equal(t, metaVal2, metaValRetrieved)

		require.NoError(t, u.DeleteUser(ctx, tx, 1, User, "user-id-1"))

		metaValRetrieved, err = u.GetUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1")
		require.NoError(t, err)
		require.Nil(t, metaValRetrieved)

		metaValRetrieved, err = u.GetUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-2")
		require.NoError(t, err)
		require.Nil(t, metaValRetrieved)

		_ = kvStore.DropTable(ctx, u.SubspaceName)
	})

	t.Run("put_get_delete_get", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		u, tx := initUserTest(t)
		defer func() { assert.NoError(t, tx.Rollback(ctx)) }()

		require.NoError(t, u.InsertUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1", testUserPayload))
		user, err := u.GetUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1")
		require.NoError(t, err)
		require.Equal(t, testUserPayload, user)

		require.NoError(t, u.DeleteUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1"))
		user, err = u.GetUserMetadata(ctx, tx, 1, User, "user-id-1", "meta-key-1")
		require.NoError(t, err)
		require.Nil(t, user)

		_ = kvStore.DropTable(ctx, u.SubspaceName)
	})
}
