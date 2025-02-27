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

	"github.com/stretchr/testify/require"
	"github.com/tigrisdata/tigris/errors"
	"github.com/tigrisdata/tigris/server/transaction"
	"github.com/tigrisdata/tigris/store/kv"
)

func TestDictionaryEncoding(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	k := NewMetadataDictionary(&NameRegistry{
		ReserveSB:  "test_reserved",
		EncodingSB: "test_encoding",
	})
	_ = kvStore.DropTable(ctx, k.EncodingSubspaceName())
	_ = kvStore.DropTable(ctx, k.ReservedSubspaceName())

	tm := transaction.NewManager(kvStore)

	tx, err := tm.StartTx(ctx)
	require.NoError(t, err)
	require.NoError(t, k.ReserveNamespace(ctx, tx, "proj1-org-1", NewNamespaceMetadata(1234, "proj1-org-1", "proj1-org-1-display_name")))
	require.NoError(t, tx.Commit(ctx))

	tx, err = tm.StartTx(ctx)
	require.NoError(t, err)
	dbId, err := k.CreateDatabase(ctx, tx, "db-1", 1234)
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	tx, err = tm.StartTx(ctx)
	require.NoError(t, err)
	collId, err := k.CreateCollection(ctx, tx, "coll-1", 1234, 1)
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	tx, err = tm.StartTx(ctx)
	require.NoError(t, err)
	indexId, err := k.CreateIndex(ctx, tx, "pkey", 1234, 1, 2)
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	tx, err = tm.StartTx(ctx)
	require.NoError(t, err)
	v, err := k.GetDatabaseId(ctx, tx, "db-1", 1234)
	require.NoError(t, err)
	require.Equal(t, v, dbId)

	v, err = k.GetCollectionId(ctx, tx, "coll-1", 1234, dbId)
	require.NoError(t, err)
	require.Equal(t, v, collId)

	v, err = k.GetIndexId(ctx, tx, "pkey", 1234, dbId, collId)
	require.NoError(t, err)
	require.Equal(t, v, indexId)
	require.NoError(t, tx.Commit(ctx))

	// try assigning the same namespace id to some other namespace
	tx, err = tm.StartTx(ctx)
	require.NoError(t, err)
	require.Error(t, k.ReserveNamespace(ctx, tx, "proj2-org-1", NewNamespaceMetadata(1234, "proj2-org-1", "proj2-org-1-display_name")))
	require.NoError(t, tx.Rollback(ctx))
}

func TestDictionaryEncodingDropped(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("drop_database", func(t *testing.T) {
		k := NewMetadataDictionary(&NameRegistry{
			ReserveSB:  "test_reserved",
			EncodingSB: "test_encoding",
		})

		_ = kvStore.DropTable(ctx, k.EncodingSubspaceName())
		_ = kvStore.DropTable(ctx, k.ReservedSubspaceName())

		tm := transaction.NewManager(kvStore)

		tx, err := tm.StartTx(ctx)
		require.NoError(t, err)
		require.NoError(t, k.ReserveNamespace(ctx, tx, "proj1-org-1", NewNamespaceMetadata(1234, "proj1-org-1", "proj1-org-1-display_name")))
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		dbId, err := k.CreateDatabase(ctx, tx, "db-1", 1234)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		v, err := k.GetDatabaseId(ctx, tx, "db-1", 1234)
		require.NoError(t, err)
		require.Equal(t, v, dbId)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		err = k.DropDatabase(ctx, tx, "db-1", 1234, dbId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		v, err = k.GetDatabaseId(ctx, tx, "db-1", 1234)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))
		require.Equal(t, v, InvalidId)
	})
	t.Run("drop_collection", func(t *testing.T) {
		k := NewMetadataDictionary(&NameRegistry{
			ReserveSB:  "test_reserved",
			EncodingSB: "test_encoding",
		})

		_ = kvStore.DropTable(ctx, k.EncodingSubspaceName())
		_ = kvStore.DropTable(ctx, k.ReservedSubspaceName())

		tm := transaction.NewManager(kvStore)
		tx, err := tm.StartTx(ctx)
		require.NoError(t, err)
		require.NoError(t, k.ReserveNamespace(ctx, tx, "proj1-org-1", NewNamespaceMetadata(1234, "proj1-org-1", "proj1-org-1-display_name")))
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		dbId, err := k.CreateDatabase(ctx, tx, "db-1", 1234)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		collId, err := k.CreateCollection(ctx, tx, "coll-1", 1234, dbId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		v, err := k.GetCollectionId(ctx, tx, "coll-1", 1234, dbId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))
		require.Equal(t, v, collId)

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		err = k.DropCollection(ctx, tx, "coll-1", 1234, dbId, collId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		v, err = k.GetCollectionId(ctx, tx, "coll-1", 1234, dbId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))
		require.Equal(t, v, InvalidId)
	})
	t.Run("drop_index", func(t *testing.T) {
		k := NewMetadataDictionary(&NameRegistry{
			ReserveSB:  "test_reserved",
			EncodingSB: "test_encoding",
		})

		_ = kvStore.DropTable(ctx, k.EncodingSubspaceName())
		_ = kvStore.DropTable(ctx, k.ReservedSubspaceName())

		tm := transaction.NewManager(kvStore)
		tx, err := tm.StartTx(ctx)
		require.NoError(t, err)
		require.NoError(t, k.ReserveNamespace(ctx, tx, "proj1-org-1", NewNamespaceMetadata(1234, "proj1-org-1", "proj1-org-1-display_name")))
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		dbId, err := k.CreateDatabase(ctx, tx, "db-1", 1234)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		collId, err := k.CreateCollection(ctx, tx, "coll-1", 1234, dbId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		idxId, err := k.CreateIndex(ctx, tx, "idx-1", 1234, dbId, collId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		v, err := k.GetIndexId(ctx, tx, "idx-1", 1234, dbId, collId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))
		require.Equal(t, v, idxId)

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		err = k.DropIndex(ctx, tx, "idx-1", 1234, dbId, collId, idxId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		v, err = k.GetIndexId(ctx, tx, "idx-1", 1234, dbId, collId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))
		require.Equal(t, v, InvalidId)
	})
	t.Run("drop_collection_multiple", func(t *testing.T) {
		k := NewMetadataDictionary(&NameRegistry{
			ReserveSB:  "test_reserved",
			EncodingSB: "test_encoding",
		})

		_ = kvStore.DropTable(ctx, k.EncodingSubspaceName())
		_ = kvStore.DropTable(ctx, k.ReservedSubspaceName())

		tm := transaction.NewManager(kvStore)
		tx, err := tm.StartTx(ctx)
		require.NoError(t, err)
		require.NoError(t, k.ReserveNamespace(ctx, tx, "proj1-org-1", NewNamespaceMetadata(1234, "proj1-org-1", "proj1-org-1-display_name")))
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		dbId, err := k.CreateDatabase(ctx, tx, "db-1", 1234)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		collId, err := k.CreateCollection(ctx, tx, "coll-1", 1234, dbId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		v, err := k.GetCollectionId(ctx, tx, "coll-1", 1234, dbId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))
		require.Equal(t, v, collId)

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		err = k.DropCollection(ctx, tx, "coll-1", 1234, dbId, collId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		v, err = k.GetCollectionId(ctx, tx, "coll-1", 1234, dbId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))
		require.Equal(t, v, InvalidId)

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		newCollId, err := k.CreateCollection(ctx, tx, "coll-1", 1234, dbId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		v, err = k.GetCollectionId(ctx, tx, "coll-1", 1234, dbId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))
		require.Equal(t, v, newCollId)
	})
}

func TestDictionaryEncoding_Error(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	k := NewMetadataDictionary(&NameRegistry{
		ReserveSB:  "test_reserved",
		EncodingSB: "test_encoding",
	})

	_ = kvStore.DropTable(ctx, k.EncodingSubspaceName())
	_ = kvStore.DropTable(ctx, k.ReservedSubspaceName())

	tm := transaction.NewManager(kvStore)
	tx, err := tm.StartTx(ctx)
	require.NoError(t, err)
	dbId, err := k.CreateDatabase(ctx, tx, "db-1", 0)
	require.Error(t, errors.InvalidArgument("invalid namespace id"), err)
	require.Equal(t, InvalidId, dbId)

	collId, err := k.CreateCollection(ctx, tx, "coll-1", 1234, 0)
	require.Error(t, errors.InvalidArgument("invalid database id"), err)
	require.Equal(t, InvalidId, collId)

	indexId, err := k.CreateIndex(ctx, tx, "pkey", 1234, 1, 0)
	require.Error(t, errors.InvalidArgument("invalid collection id"), err)
	require.Equal(t, InvalidId, indexId)
	require.NoError(t, tx.Rollback(context.TODO()))
}

func TestDictionaryEncoding_GetMethods(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tm := transaction.NewManager(kvStore)
	t.Run("get_databases", func(t *testing.T) {
		k := NewMetadataDictionary(&NameRegistry{
			ReserveSB:  "test_reserved",
			EncodingSB: "test_encoding",
		})

		_ = kvStore.DropTable(ctx, k.EncodingSubspaceName())
		_ = kvStore.DropTable(ctx, k.ReservedSubspaceName())

		tx, err := tm.StartTx(ctx)
		require.NoError(t, err)
		dbId1, err := k.CreateDatabase(ctx, tx, "db-1", 1)
		require.NoError(t, err)
		dbId2, err := k.CreateDatabase(ctx, tx, "db-2", 1)
		require.NoError(t, err)

		dbToId, err := k.GetDatabases(ctx, tx, 1)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))
		require.Len(t, dbToId, 2)
		require.Equal(t, dbToId["db-1"], dbId1)
		require.Equal(t, dbToId["db-2"], dbId2)
	})
	t.Run("get_collections", func(t *testing.T) {
		k := NewMetadataDictionary(&NameRegistry{
			ReserveSB:  "test_reserved",
			EncodingSB: "test_encoding",
		})

		_ = kvStore.DropTable(ctx, k.EncodingSubspaceName())
		_ = kvStore.DropTable(ctx, k.ReservedSubspaceName())

		tx, err := tm.StartTx(ctx)
		require.NoError(t, err)
		dbId, err := k.CreateDatabase(ctx, tx, "db-1", 1)
		require.NoError(t, err)

		cid1, err := k.CreateCollection(ctx, tx, "coll-1", 1, dbId)
		require.NoError(t, err)
		cid2, err := k.CreateCollection(ctx, tx, "coll-2", 1, dbId)
		require.NoError(t, err)

		collToId, err := k.GetCollections(ctx, tx, 1, dbId)
		require.NoError(t, err)
		require.NoError(t, tx.Commit(ctx))
		require.Len(t, collToId, 2)
		require.Equal(t, collToId["coll-1"], cid1)
		require.Equal(t, collToId["coll-2"], cid2)
	})
	t.Run("get_indexes", func(t *testing.T) {
		k := NewMetadataDictionary(&NameRegistry{
			ReserveSB:  "test_reserved",
			EncodingSB: "test_encoding",
		})

		_ = kvStore.DropTable(ctx, k.EncodingSubspaceName())
		_ = kvStore.DropTable(ctx, k.ReservedSubspaceName())

		tx, err := tm.StartTx(ctx)
		require.NoError(t, err)
		dbId, err := k.CreateDatabase(ctx, tx, "db-1", 1)
		require.NoError(t, err)

		cid1, err := k.CreateCollection(ctx, tx, "coll-1", 1, dbId)
		require.NoError(t, err)

		pkID, err := k.CreateIndex(ctx, tx, "pkey", 1, dbId, cid1)
		require.NoError(t, err)
		idxToId, err := k.GetIndexes(ctx, tx, 1, dbId, cid1)
		require.NoError(t, err)
		require.Len(t, idxToId, 1)
		require.Equal(t, idxToId["pkey"], pkID)
		require.NoError(t, tx.Commit(ctx))
	})
}

func TestReservedNamespace(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r := newReservedSubspace(&NameRegistry{
		ReserveSB:  "test_reserved",
		EncodingSB: "test_encoding",
	})

	_ = kvStore.DropTable(ctx, r.EncodingSubspaceName())
	_ = kvStore.DropTable(ctx, r.ReservedSubspaceName())

	tm := transaction.NewManager(kvStore)

	tx, err := tm.StartTx(ctx)
	require.NoError(t, err)
	require.NoError(t, r.reserveNamespace(ctx, tx, "p1-o1", NewNamespaceMetadata(123, "p1-o1", "p1-o1-display_name")))
	require.NoError(t, tx.Commit(ctx))

	// check in the allocated id is assigned
	tx, err = tm.StartTx(ctx)
	require.NoError(t, err)
	require.NoError(t, r.reload(ctx, tx))
	require.Equal(t, "p1-o1", r.idToNamespaceStruct[123].StrId)

	// try assigning the same namespace id to some other namespace
	tx, err = tm.StartTx(context.TODO())
	require.NoError(t, err)
	expError := errors.AlreadyExists("id is already assigned to the namespace 'p1-o1'")
	require.Equal(t, expError, r.reserveNamespace(context.TODO(), tx, "p2-o2", NewNamespaceMetadata(123, "p2-o2", "p2-o2-display_name")))
}

func TestDecode(t *testing.T) {
	k := kv.BuildKey(encVersion, UInt32ToByte(1234), dbKey, "db-1", keyEnd)
	mp, err := NewMetadataDictionary(&NameRegistry{
		ReserveSB:  "test_reserved",
		EncodingSB: "test_encoding",
	}).decode(context.TODO(), k)
	require.NoError(t, err)
	require.Equal(t, mp[dbKey], "db-1")
}
