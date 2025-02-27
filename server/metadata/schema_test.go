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
	"github.com/tigrisdata/tigris/store/kv"
)

func initSchemaTest(t *testing.T) (*SchemaSubspace, transaction.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s := NewSchemaStore(&NameRegistry{
		SchemaSB: "test_schema",
	})

	_ = kvStore.DropTable(ctx, s.SubspaceName)

	tm := transaction.NewManager(kvStore)
	tx, err := tm.StartTx(ctx)
	require.NoError(t, err)

	return s, tx
}

func TestSchemaSubspace(t *testing.T) {
	t.Run("put_error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		s, tx := initSchemaTest(t)
		defer func() { assert.NoError(t, tx.Rollback(ctx)) }()

		require.Equal(t, errors.InvalidArgument("invalid schema version %d", 0), s.Put(ctx, tx, 1, 2, 3, nil, 0))
		require.Equal(t, errors.InvalidArgument("empty schema"), s.Put(ctx, tx, 1, 2, 3, nil, 1))
	})

	t.Run("put_duplicate_error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		schema := []byte(`{"title": "test schema1"}`)

		s, tx := initSchemaTest(t)
		defer func() { assert.NoError(t, tx.Rollback(ctx)) }()

		require.NoError(t, s.Put(ctx, tx, 1, 2, 3, schema, 1))
		require.Equal(t, kv.ErrDuplicateKey, s.Put(ctx, tx, 1, 2, 3, schema, 1))
	})

	t.Run("put_get", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		s, tx := initSchemaTest(t)
		defer func() { assert.NoError(t, tx.Rollback(ctx)) }()

		schema := []byte(`{
		"title": "collection1",
		"description": "this schema is for client integration tests",
		"properties": {
			"K1": {
				"type": "string"
			},
			"K2": {
				"type": "integer"
			},
			"D1": {
				"type": "string",
				"max_length": 128
			}
		},
		"primary_key": ["K1", "K2"]
	}`)

		require.NoError(t, s.Put(ctx, tx, 1, 2, 3, schema, 1))
		sch, err := s.GetLatest(ctx, tx, 1, 2, 3)
		require.NoError(t, err)
		require.Equal(t, schema, sch.Schema)
		require.Equal(t, 1, sch.Version)

		schemas, err := s.Get(ctx, tx, 1, 2, 3)
		require.NoError(t, err)
		require.Equal(t, schema, schemas[0].Schema)
		require.Equal(t, 1, schemas[0].Version)

		_ = kvStore.DropTable(ctx, s.SubspaceName)
	})

	t.Run("put_get_multiple", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		s, tx := initSchemaTest(t)
		defer func() { assert.NoError(t, tx.Rollback(ctx)) }()

		schema1 := []byte(`{
		"title": "collection1",
		"description": "this schema is for client integration tests",
		"properties": {
			"K1": {
				"type": "string"
			},
			"K2": {
				"type": "integer"
			},
			"D1": {
				"type": "string",
				"max_length": 128
			}
		},
		"primary_key": ["K1", "K2"]
	}`)
		schema2 := []byte(`{
		"title": "collection1",
		"description": "this schema is for client integration tests",
		"properties": {
			"K1": {
				"type": "string"
			},
			"K2": {
				"type": "integer"
			}
		},
		"primary_key": ["K1"]
	}`)

		require.NoError(t, s.Put(ctx, tx, 1, 2, 3, schema1, 1))
		require.NoError(t, s.Put(ctx, tx, 1, 2, 3, schema2, 2))
		sch, err := s.GetLatest(ctx, tx, 1, 2, 3)
		require.NoError(t, err)
		require.Equal(t, schema2, sch.Schema)
		require.Equal(t, 2, sch.Version)

		schemas, err := s.Get(ctx, tx, 1, 2, 3)
		require.NoError(t, err)
		require.Equal(t, schema1, schemas[0].Schema)
		require.Equal(t, schema2, schemas[1].Schema)
		require.Equal(t, 1, schemas[0].Version)
		require.Equal(t, 2, schemas[1].Version)
	})

	t.Run("put_delete_get", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		s, tx := initSchemaTest(t)
		defer func() { assert.NoError(t, tx.Rollback(ctx)) }()

		schema1 := []byte(`{
		"title": "collection1",
		"description": "this schema is for client integration tests",
		"properties": {
			"K1": {
				"type": "string"
			},
			"K2": {
				"type": "integer"
			},
			"D1": {
				"type": "string",
				"max_length": 128
			}
		},
		"primary_key": ["K1", "K2"]
	}`)
		schema2 := []byte(`{
		"title": "collection1",
		"description": "this schema is for client integration tests",
		"properties": {
			"K1": {
				"type": "string"
			},
			"K2": {
				"type": "integer"
			}
		},
		"primary_key": ["K1"]
	}`)

		require.NoError(t, s.Put(ctx, tx, 1, 2, 3, schema1, 1))
		require.NoError(t, s.Put(ctx, tx, 1, 2, 3, schema2, 2))

		schemas, err := s.Get(ctx, tx, 1, 2, 3)
		require.NoError(t, err)
		require.Equal(t, schema1, schemas[0].Schema)
		require.Equal(t, schema2, schemas[1].Schema)
		require.Len(t, schemas, 2)
		require.NoError(t, tx.Commit(ctx))

		tm := transaction.NewManager(kvStore)

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		require.NoError(t, s.Delete(ctx, tx, 1, 2, 3))
		require.NoError(t, tx.Commit(ctx))

		tx, err = tm.StartTx(ctx)
		require.NoError(t, err)
		schemas, err = s.Get(ctx, tx, 1, 2, 3)
		require.NoError(t, err)
		require.Len(t, schemas, 0)
		require.NoError(t, tx.Commit(ctx))
	})
}
