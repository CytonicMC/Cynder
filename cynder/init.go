package cynder

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/CytonicMC/Cynder/cynder/messaging"
	"github.com/CytonicMC/Cynder/cynder/natsMsgr"
	"github.com/CytonicMC/Cynder/cynder/natsMsgr/players"
	"github.com/CytonicMC/Cynder/cynder/natsMsgr/servers"
	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"
	"github.com/robinbraemer/event"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

type Dependencies struct {
	NatsConn *nats.Conn
	Logger   logr.Logger
	//Redis    redis.Client
}

var (
	NatsService messaging.NatsService
	Instance    *Cynder
	Context     context.Context
)

// Plugin initialization using dependencies
var Plugin = proxy.Plugin{
	Name: "CynderInit",
	Init: func(ctx context.Context, p *proxy.Proxy) error {
		deps, err := InitializeDependencies(ctx, p)
		if err != nil {
			return err
		}

		ns := &messaging.NatsServiceImpl{
			Conn: deps.NatsConn,
		}

		Instance = &Cynder{
			proxy:        p,
			dependencies: deps,
		}
		Context = ctx

		// player stuff
		players.HandlePlayerSend(ns, p, ctx)
		players.HandlePlayerKick(ns, p, ctx)

		// servers
		servers.FetchServers(ns, p)
		servers.ListenForServerRegistrations(ns, p)
		servers.ListenForServerShutdowns(ns, p)

		registerEvents(p, ns, deps.Logger)

		return nil
	},
}

func InitializeDependencies(ctx context.Context, p *proxy.Proxy) (*Dependencies, error) {
	log := logr.FromContextOrDiscard(ctx).WithName("Cynder")
	natsConn := natsMsgr.ConnectToNats()
	//redisClient := redis.PubsubClient() // Assuming NewClient initializes Redis

	return &Dependencies{
		NatsConn: natsConn,
		Logger:   log,
		//Redis:    redisClient,
	}, nil
}

type Cynder struct {
	proxy        *proxy.Proxy
	dependencies *Dependencies
}

func registerEvents(p *proxy.Proxy, nc messaging.NatsService, logger logr.Logger) {
	event.Subscribe(p.Event(), 0, func(e *proxy.PlayerChooseInitialServerEvent) {
		//todo: grouping fallbacks, etc
		logger.Info("<< - CHOOSE INITIAL SERVER - >>")
		server := servers.GetLeastLoadedServer("GROUP_HERE")
		fmt.Printf("Server: %v", server)
		e.SetInitialServer(server)
	})

	event.Subscribe(p.Event(), 0, func(e *proxy.PreLoginEvent) {
		id, _ := e.ID()
		players.BroadcastPlayerJoin(nc, e.Username(), id, logger)
	})

	event.Subscribe(p.Event(), 0, func(e *proxy.DisconnectEvent) {
		id := e.Player().ID()
		players.BroadcastPlayerLeave(nc, e.Player().Username(), id, logger)
	})

	event.Subscribe(p.Event(), 0, func(e *proxy.ServerPostConnectEvent) {

		var oldServer string
		if e.PreviousServer() != nil {
			oldServer = e.PreviousServer().ServerInfo().Name()
		}

		container := &messaging.PlayerChangeServerContainer{
			Player:    e.Player().ID(),
			OldServer: oldServer,
			NewServer: e.Player().CurrentServer().Server().ServerInfo().Name(),
		}

		data, err := json.Marshal(container)
		if err != nil {
			logger.Error(err, "Failed to marshal PlayerChangeServerContainer")
			return
		}

		er1 := nc.Publish("players.server_change.notify", data)
		if er1 != nil {
			logger.Error(err, "Failed broadcast player server change!")
			return
		}
	})
}
