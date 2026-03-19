//go:build wireinject

package cmd

import (
	"context"

	"github.com/google/wire"
	"github.com/ikelvingo/navidrome/adapters/lastfm"
	"github.com/ikelvingo/navidrome/adapters/listenbrainz"
	"github.com/ikelvingo/navidrome/core"
	"github.com/ikelvingo/navidrome/core/agents"
	"github.com/ikelvingo/navidrome/core/artwork"
	"github.com/ikelvingo/navidrome/core/lyrics"
	"github.com/ikelvingo/navidrome/core/metrics"
	"github.com/ikelvingo/navidrome/core/playback"
	"github.com/ikelvingo/navidrome/core/scrobbler"
	"github.com/ikelvingo/navidrome/db"
	"github.com/ikelvingo/navidrome/model"
	"github.com/ikelvingo/navidrome/persistence"
	"github.com/ikelvingo/navidrome/plugins"
	"github.com/ikelvingo/navidrome/scanner"
	"github.com/ikelvingo/navidrome/server"
	"github.com/ikelvingo/navidrome/server/events"
	"github.com/ikelvingo/navidrome/server/nativeapi"
	"github.com/ikelvingo/navidrome/server/public"
	"github.com/ikelvingo/navidrome/server/subsonic"
)

var allProviders = wire.NewSet(
	core.Set,
	artwork.Set,
	server.New,
	subsonic.New,
	nativeapi.New,
	public.New,
	persistence.New,
	lastfm.NewRouter,
	listenbrainz.NewRouter,
	events.GetBroker,
	scanner.New,
	scanner.GetWatcher,
	metrics.GetPrometheusInstance,
	db.Db,
	plugins.GetManager,
	wire.Bind(new(agents.PluginLoader), new(*plugins.Manager)),
	wire.Bind(new(scrobbler.PluginLoader), new(*plugins.Manager)),
	wire.Bind(new(lyrics.PluginLoader), new(*plugins.Manager)),
	wire.Bind(new(nativeapi.PluginManager), new(*plugins.Manager)),
	wire.Bind(new(core.PluginUnloader), new(*plugins.Manager)),
	wire.Bind(new(plugins.PluginMetricsRecorder), new(metrics.Metrics)),
	wire.Bind(new(core.Watcher), new(scanner.Watcher)),
)

func CreateDataStore() model.DataStore {
	panic(wire.Build(
		allProviders,
	))
}

func CreateServer() *server.Server {
	panic(wire.Build(
		allProviders,
	))
}

func CreateNativeAPIRouter(ctx context.Context) *nativeapi.Router {
	panic(wire.Build(
		allProviders,
	))
}

func CreateSubsonicAPIRouter(ctx context.Context) *subsonic.Router {
	panic(wire.Build(
		allProviders,
	))
}

func CreatePublicRouter() *public.Router {
	panic(wire.Build(
		allProviders,
	))
}

func CreateLastFMRouter() *lastfm.Router {
	panic(wire.Build(
		allProviders,
	))
}

func CreateListenBrainzRouter() *listenbrainz.Router {
	panic(wire.Build(
		allProviders,
	))
}

func CreateInsights() metrics.Insights {
	panic(wire.Build(
		allProviders,
	))
}

func CreatePrometheus() metrics.Metrics {
	panic(wire.Build(
		allProviders,
	))
}

func CreateScanner(ctx context.Context) model.Scanner {
	panic(wire.Build(
		allProviders,
	))
}

func CreateScanWatcher(ctx context.Context) scanner.Watcher {
	panic(wire.Build(
		allProviders,
	))
}

func GetPlaybackServer() playback.PlaybackServer {
	panic(wire.Build(
		allProviders,
	))
}

func getPluginManager() *plugins.Manager {
	panic(wire.Build(
		allProviders,
	))
}

func GetPluginManager(ctx context.Context) *plugins.Manager {
	manager := getPluginManager()
	manager.SetSubsonicRouter(CreateSubsonicAPIRouter(ctx))
	return manager
}
