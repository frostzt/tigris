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

package v1

import (
	"context"
	"net/http"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	api "github.com/tigrisdata/tigris/api/server/v1"
	"github.com/tigrisdata/tigris/server/config"
	"github.com/tigrisdata/tigris/server/services/v1/auth"
	"google.golang.org/grpc"
)

const (
	authPattern = "/" + version + "/auth/*"
)

type authService struct {
	api.UnimplementedAuthServer
	auth.Provider
}

func newAuthService(authProvider auth.Provider) *authService {
	if authProvider == nil {
		log.Error().Str("Provider", config.DefaultConfig.Auth.OAuthProvider).Msg("Unable to configure external oauth provider")
		panic("Unable to configure external oauth provider")
	}
	return &authService{
		Provider: authProvider,
	}
}

func (a *authService) GetAccessToken(ctx context.Context, req *api.GetAccessTokenRequest) (*api.GetAccessTokenResponse, error) {
	return a.Provider.GetAccessToken(ctx, req)
}

func (a *authService) RegisterHTTP(router chi.Router, inproc *inprocgrpc.Channel) error {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &api.CustomMarshaler{JSONBuiltin: &runtime.JSONBuiltin{}}),
	)
	if err := api.RegisterAuthHandlerClient(context.TODO(), mux, api.NewAuthClient(inproc)); err != nil {
		return err
	}
	api.RegisterAuthServer(inproc, a)
	router.HandleFunc(authPattern, func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r)
	})
	return nil
}

func (a *authService) RegisterGRPC(grpc *grpc.Server) error {
	api.RegisterAuthServer(grpc, a)
	return nil
}
