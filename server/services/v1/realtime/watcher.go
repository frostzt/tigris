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

package realtime

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/tigrisdata/tigris/store/cache"
)

// Watch is called when an event is received by ChannelWatcher.
type Watch func(*cache.StreamMessages, error) ([]string, error)

// ChannelWatcher is to watch events for a single channel. It accepts a watch that will be notified when a new event
// is read from the stream. As ChannelWatcher is mapped to a consumer group on a stream therefore the state is restored
// from the cache during restart which means a watcher is only created if it doesn’t exist otherwise the existing one
// is returned.
type ChannelWatcher struct {
	ctx           context.Context
	name          string
	watch         Watch
	stream        cache.Stream
	sigStop       chan struct{}
	sigDisconnect chan struct{}
}

func CreateWatcher(ctx context.Context, name string, pos string, existingPos string, stream cache.Stream) (*ChannelWatcher, error) {
	w := newWatcher(ctx, name, stream)
	if len(pos) == 0 {
		// just use the existing id
		err := w.move(ctx, existingPos)
		return w, err
	} else {
		if pos > existingPos {
			log.Warn().Msgf("new position '%s' is greater than existing pos '%s' for watch '%s'", pos, existingPos, name)
		}

		err := w.move(ctx, pos)
		return w, err
	}
}

func CreateAndRegisterWatcher(ctx context.Context, name string, pos string, stream cache.Stream) (*ChannelWatcher, error) {
	if len(pos) == 0 {
		pos = cache.ConsumerGroupDefaultCurrentPos
	}

	if err := stream.CreateConsumerGroup(ctx, name, pos); err != nil {
		return nil, err
	}

	return newWatcher(ctx, name, stream), nil
}

func newWatcher(ctx context.Context, id string, stream cache.Stream) *ChannelWatcher {
	return &ChannelWatcher{
		ctx:           ctx,
		name:          id,
		stream:        stream,
		sigStop:       make(chan struct{}),
		sigDisconnect: make(chan struct{}),
	}
}

func (watcher *ChannelWatcher) StartWatching(watch Watch) {
	watcher.watch = watch
	go watcher.watchEvents()
}

func (watcher *ChannelWatcher) move(ctx context.Context, newPos string) error {
	if len(newPos) == 0 {
		return nil
	}

	return watcher.stream.SetID(ctx, watcher.name, newPos)
}

func (watcher *ChannelWatcher) Stop() {
	close(watcher.sigStop)
}

func (watcher *ChannelWatcher) Disconnect() {
	close(watcher.sigDisconnect)
}

func (watcher *ChannelWatcher) watchEvents() {
	for {
		select {
		case <-watcher.sigStop:
			return
		case <-watcher.sigDisconnect:
			_ = watcher.stream.RemoveConsumerGroup(watcher.ctx, watcher.name)
			return
		default:
			resp, hasData, err := watcher.stream.ReadGroup(watcher.ctx, watcher.name, cache.ReadGroupPosCurrent)
			if err != nil {
				continue
			}

			if !hasData {
				continue
			}

			if ids, err := watcher.watch(resp, nil); err == nil {
				_ = watcher.ack(watcher.ctx, ids)
			}
		}
	}
}

func (watcher *ChannelWatcher) ack(ctx context.Context, ids []string) error {
	return watcher.stream.Ack(ctx, watcher.name, ids...)
}
