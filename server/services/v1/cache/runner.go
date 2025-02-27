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

package cache

import (
	"context"

	"github.com/rs/zerolog/log"
	api "github.com/tigrisdata/tigris/api/server/v1"
	"github.com/tigrisdata/tigris/errors"
	"github.com/tigrisdata/tigris/internal"
	"github.com/tigrisdata/tigris/server/config"
	"github.com/tigrisdata/tigris/server/metadata"
	"github.com/tigrisdata/tigris/server/request"
	"github.com/tigrisdata/tigris/server/services/v1/database"
	"github.com/tigrisdata/tigris/server/transaction"
	"github.com/tigrisdata/tigris/server/types"
	"github.com/tigrisdata/tigris/store/cache"
	ulog "github.com/tigrisdata/tigris/util/log"
)

// Runner is responsible for executing the current query and return the response.
type Runner interface {
	Run(ctx context.Context, tenant *metadata.Tenant) (Response, error)
}

// TxRunner is responsible for executing the current query and return the response.
type TxRunner interface {
	Run(ctx context.Context, tx transaction.Tx, tenant *metadata.Tenant) (Response, context.Context, error)
}

type BaseRunner struct {
	encoder     metadata.CacheEncoder
	cacheStore  cache.Cache
	accessToken *types.AccessToken
}

func NewBaseRunner(encoder metadata.CacheEncoder, accessToken *types.AccessToken, cacheStore cache.Cache) *BaseRunner {
	return &BaseRunner{
		encoder:     encoder,
		accessToken: accessToken,
		cacheStore:  cacheStore,
	}
}

type CreateCacheRunner struct {
	*BaseRunner

	req *api.CreateCacheRequest
}

type DeleteCacheRunner struct {
	*BaseRunner

	req *api.DeleteCacheRequest
}

type ListCachesRunner struct {
	*BaseRunner

	req *api.ListCachesRequest
}

type SetRunner struct {
	*BaseRunner

	req *api.SetRequest
}

type GetSetRunner struct {
	*BaseRunner

	req *api.GetSetRequest
}

type GetRunner struct {
	*BaseRunner

	req *api.GetRequest
}

type DelRunner struct {
	*BaseRunner

	req *api.DelRequest
}

type KeysRunner struct {
	*BaseRunner
	req       *api.KeysRequest
	streaming StreamingKeys
}

type RunnerFactory struct {
	encoder    metadata.CacheEncoder
	cacheStore cache.Cache
}

// NewRunnerFactory returns RunnerFactory object.
func NewRunnerFactory(encoder metadata.CacheEncoder, cacheStore cache.Cache) *RunnerFactory {
	return &RunnerFactory{
		encoder:    encoder,
		cacheStore: cacheStore,
	}
}

func (f *RunnerFactory) GetCreateCacheRunner(r *api.CreateCacheRequest, accessToken *types.AccessToken) *CreateCacheRunner {
	return &CreateCacheRunner{
		BaseRunner: NewBaseRunner(f.encoder, accessToken, f.cacheStore),
		req:        r,
	}
}

func (f *RunnerFactory) GetDeleteCacheRunner(r *api.DeleteCacheRequest, accessToken *types.AccessToken) *DeleteCacheRunner {
	return &DeleteCacheRunner{
		BaseRunner: NewBaseRunner(f.encoder, accessToken, f.cacheStore),
		req:        r,
	}
}

func (f *RunnerFactory) GetListCachesRunner(r *api.ListCachesRequest, accessToken *types.AccessToken) *ListCachesRunner {
	return &ListCachesRunner{
		BaseRunner: NewBaseRunner(f.encoder, accessToken, f.cacheStore),
		req:        r,
	}
}

func (f *RunnerFactory) GetSetRunner(r *api.SetRequest, accessToken *types.AccessToken) *SetRunner {
	return &SetRunner{
		BaseRunner: NewBaseRunner(f.encoder, accessToken, f.cacheStore),
		req:        r,
	}
}

func (f *RunnerFactory) GetGetSetRunner(r *api.GetSetRequest, accessToken *types.AccessToken) *GetSetRunner {
	return &GetSetRunner{
		BaseRunner: NewBaseRunner(f.encoder, accessToken, f.cacheStore),
		req:        r,
	}
}

func (f *RunnerFactory) GetGetRunner(r *api.GetRequest, accessToken *types.AccessToken) *GetRunner {
	return &GetRunner{
		BaseRunner: NewBaseRunner(f.encoder, accessToken, f.cacheStore),
		req:        r,
	}
}

func (f *RunnerFactory) GetDelRunner(r *api.DelRequest, accessToken *types.AccessToken) *DelRunner {
	return &DelRunner{
		BaseRunner: NewBaseRunner(f.encoder, accessToken, f.cacheStore),
		req:        r,
	}
}

func (f *RunnerFactory) GetKeysRunner(r *api.KeysRequest, accessToken *types.AccessToken, streaming StreamingKeys) *KeysRunner {
	return &KeysRunner{
		BaseRunner: NewBaseRunner(f.encoder, accessToken, f.cacheStore),
		streaming:  streaming,
		req:        r,
	}
}

func (runner *CreateCacheRunner) Run(ctx context.Context, tx transaction.Tx, tenant *metadata.Tenant) (Response, context.Context, error) {
	currentSub, err := request.GetCurrentSub(ctx)
	if err != nil && config.DefaultConfig.Auth.Enabled {
		return Response{}, ctx, errors.Internal("Failed to get current sub for the request")
	}

	_, err = tenant.CreateCache(ctx, tx, runner.req.GetProject(), runner.req.GetName(), currentSub)
	if err != nil {
		return Response{}, ctx, createApiError(err)
	}
	return Response{
		Status: database.CreatedStatus,
	}, ctx, nil
}

func (runner *DeleteCacheRunner) Run(ctx context.Context, tx transaction.Tx, tenant *metadata.Tenant) (Response, context.Context, error) {
	tableName, err := getEncodedCacheTableName(ctx, tenant, runner.req.GetProject(), runner.req.GetName(), runner.encoder)
	if err != nil {
		return Response{}, ctx, err
	}

	internalKeys, err := runner.cacheStore.Keys(ctx, tableName, "*")
	if err != nil {
		return Response{}, ctx, err
	}
	for _, internalKey := range internalKeys {
		// translate the key to user key
		userKey := runner.encoder.DecodeInternalCacheKeyNameToExternal(internalKey)
		_, err = runner.cacheStore.Delete(ctx, tableName, userKey)
		if err != nil {
			log.Warn().Str("cacheTableName", tableName).Str("cacheKey", userKey).Msg("Failed to delete cache key")
		}
	}

	_, err = tenant.DeleteCache(ctx, tx, runner.req.GetProject(), runner.req.GetName())
	if err != nil {
		log.Warn().
			Str("tenant", tenant.GetNamespace().StrId()).
			Str("project", runner.req.GetProject()).
			Str("cache", runner.req.GetName()).
			Msg("Failed to update project metadata entry to delete cache")
		return Response{}, ctx, createApiError(err)
	}
	return Response{
		Status: database.DeletedStatus,
	}, ctx, nil
}

func (runner *ListCachesRunner) Run(ctx context.Context, tx transaction.Tx, tenant *metadata.Tenant) (Response, context.Context, error) {
	caches, err := tenant.ListCaches(ctx, tx, runner.req.GetProject())
	if err != nil {
		return Response{}, ctx, err
	}
	cachesMetadata := make([]*api.CacheMetadata, len(caches))
	for i, cache := range caches {
		cachesMetadata[i] = &api.CacheMetadata{
			Name: cache,
		}
	}
	return Response{
		Caches: cachesMetadata,
	}, ctx, nil
}

func (runner *SetRunner) Run(ctx context.Context, tenant *metadata.Tenant) (Response, error) {
	tableName, err := getEncodedCacheTableName(ctx, tenant, runner.req.GetProject(), runner.req.GetName(), runner.encoder)
	if err != nil {
		return Response{}, err
	}

	options := &cache.SetOptions{
		NX: runner.req.GetNx(),
		XX: runner.req.GetXx(),
		EX: runner.req.GetEx(),
		PX: runner.req.GetPx(),
	}

	if err = runner.cacheStore.Set(ctx, tableName, runner.req.GetKey(), internal.NewCacheData(runner.req.GetValue()), options); err != nil {
		return Response{}, errors.Internal("Failed to invoke set, reason %s", err.Error())
	}

	return Response{
		Status: SetStatus,
	}, nil
}

func (runner *GetSetRunner) Run(ctx context.Context, tenant *metadata.Tenant) (Response, error) {
	tableName, err := getEncodedCacheTableName(ctx, tenant, runner.req.GetProject(), runner.req.GetName(), runner.encoder)
	if err != nil {
		return Response{}, err
	}

	oldVal, err := runner.cacheStore.GetSet(ctx, tableName, runner.req.GetKey(), internal.NewCacheData(runner.req.GetValue()))
	if err != nil {
		return Response{}, errors.Internal("Failed to invoke set, reason %s", err.Error())
	}

	var result Response = Response{
		Status: SetStatus,
	}

	if oldVal != nil && oldVal.RawData != nil {
		result.OldValue = oldVal.GetRawData()
	}
	return result, nil
}

func (runner *GetRunner) Run(ctx context.Context, tenant *metadata.Tenant) (Response, error) {
	tableName, err := getEncodedCacheTableName(ctx, tenant, runner.req.GetProject(), runner.req.GetName(), runner.encoder)
	if err != nil {
		return Response{}, err
	}

	options := &cache.GetOptions{
		Expiry:    0,
		GetDelete: false,
	}
	data, err := runner.cacheStore.Get(ctx, tableName, runner.req.GetKey(), options)
	if err != nil {
		if err == cache.ErrKeyNotFound {
			return Response{}, errors.NotFound(err.Error())
		}
		return Response{}, err
	}
	return Response{
		Data: data.GetRawData(),
	}, nil
}

func (runner *DelRunner) Run(ctx context.Context, tenant *metadata.Tenant) (Response, error) {
	tableName, err := getEncodedCacheTableName(ctx, tenant, runner.req.GetProject(), runner.req.GetName(), runner.encoder)
	if err != nil {
		return Response{}, err
	}

	deletedCount, err := runner.cacheStore.Delete(ctx, tableName, runner.req.GetKey())
	if err != nil {
		return Response{}, errors.Internal("Failed to invoke del, reason %s", err.Error())
	}
	return Response{
		Status:       DeletedStatus,
		DeletedCount: deletedCount,
	}, nil
}

func (runner *KeysRunner) Run(ctx context.Context, tenant *metadata.Tenant) (Response, error) {
	tableName, err := getEncodedCacheTableName(ctx, tenant, runner.req.GetProject(), runner.req.GetName(), runner.encoder)
	if err != nil {
		return Response{}, err
	}

	pattern := runner.req.GetPattern()
	if pattern == "" {
		pattern = "*"
	}
	cursor := runner.req.GetCursor()
	var internalKeys []string
	for {
		internalKeys, cursor = runner.cacheStore.Scan(ctx, tableName, cursor, runner.req.GetCount(), pattern)

		// transform internal keys to user facing keys
		userKeys := make([]string, len(internalKeys))
		for index, internalKey := range internalKeys {
			userKeys[index] = runner.encoder.DecodeInternalCacheKeyNameToExternal(internalKey)
		}

		if len(userKeys) > 0 {
			if err := runner.streaming.Send(&api.KeysResponse{
				Keys:   userKeys,
				Cursor: &cursor,
			}); ulog.E(err) {
				return Response{}, err
			}
		}
		if cursor == 0 {
			break
		}
	}
	return Response{}, nil
}

func getEncodedCacheTableName(_ context.Context, tenant *metadata.Tenant, projectName string, cacheName string, encoder metadata.CacheEncoder) (string, error) {
	project, err := tenant.GetProject(projectName)
	if err != nil {
		return "", createApiError(err)
	}

	// Encode cache table is encoding tenant id, project id(main database id) and cache name.
	encodedCacheTableName, err := encoder.EncodeCacheTableName(tenant.GetNamespace().Id(), project.Id(), cacheName)
	if err != nil {
		return "", err
	}

	return encodedCacheTableName, nil
}
