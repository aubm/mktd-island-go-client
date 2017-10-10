package game

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"mktd5/mktd-island/client/game/web"
	"mktd5/mktd-island/client/log"
	"mktd5/mktd-island/client/mediator"
	"mktd5/mktd-island/client/utils"

	"github.com/Pallinder/go-randomdata"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const defaultPort = 9000

type Agent struct {
	AppConfig      *utils.AppConfig    `inject:""`
	MediatorClient *mediator.Client    `inject:""`
	MoveStrategy   moveStrategy        `inject:""`
	Logger         log.LoggerInterface `inject:""`
	StopCallback   func()
	playerIP       string
	playerID       int
	teamName       string
	playerPort     int
}

func (a *Agent) StartClientServer(ctx context.Context) error {
	a.Logger.Info("starting game", nil)

	if err := a.detectMyLocalIp(); err != nil {
		return errors.Wrap(err, "failed to get local network ip")
	}
	a.findApplicationPort()
	a.findTeamName()

	a.configureWebServerHandlers()

	serverStoppedListening := make(chan error, 1)
	server := &http.Server{Addr: fmt.Sprintf(":%v", a.playerPort)}

	go func(done chan error) {
		a.Logger.Info("starting player web server", log.Fields{"port": a.playerPort})
		err := server.ListenAndServe()
		switch err {
		case http.ErrServerClosed, nil:
			done <- nil
		default:
			done <- errors.Wrap(err, "player web server stopped")
		}
	}(serverStoppedListening)

	a.Logger.Info("will register to the mediator in 2 seconds", nil)
	willRegister := time.After(2 * time.Second)
	for {
		select {
		case <-ctx.Done():
			a.Logger.Info("shutting down player web server", nil)
			return server.Shutdown(context.TODO())
		case err := <-serverStoppedListening:
			return err
		case <-willRegister: // this is done to ensure that the server is correctly started before registering
			if err := a.register(); err != nil {
				if errors.Cause(err) != mediator.ErrGameFull {
					return errors.Wrap(err, "failed to register to the mediator")
				}
				a.Logger.Warn("could not register to the mediator, the game is full", nil)
			}
		}
	}
}

func (a *Agent) detectMyLocalIp() (err error) {
	a.playerIP, err = utils.GetLocalNetworkIp()
	if err != nil {
		// TODO: fallback to flag if error
		return err
	}
	a.Logger.Info("found my local ip", log.Fields{"ip": a.playerIP})
	return nil
}

func (a *Agent) findApplicationPort() {
	if a.AppConfig.Port > 0 {
		a.playerPort = a.AppConfig.Port
		return
	}
	var err error
	a.playerPort, err = a.getFreePort()
	if err != nil {
		a.playerPort = defaultPort
	}
}

func (a *Agent) getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

func (a *Agent) findTeamName() {
	if a.AppConfig.TeamName != "" {
		a.teamName = a.AppConfig.TeamName
		return
	}
	a.teamName = randomdata.SillyName()
}

func (a *Agent) register() error {
	endpoint := fmt.Sprintf("%v:%v", a.playerIP, a.playerPort)

	a.Logger.Info("about to register to the mediator", log.Fields{"team": a.teamName, "endpoint": endpoint})
	registerResult, err := a.MediatorClient.Register(mediator.RegisterOptions{
		Name:     a.teamName,
		Endpoint: endpoint,
	})
	if err != nil {
		return err
	}

	a.playerID = registerResult.PlayerID
	a.Logger.Info("registration done", log.Fields{"playerId": a.playerID})

	return nil
}

func (a *Agent) configureWebServerHandlers() {
	router := mux.NewRouter()
	router.Handle("/", http.RedirectHandler("/ui", http.StatusMovedPermanently))
	router.HandleFunc("/ui", a.ui)
	router.HandleFunc("/map/{id}", a.handleMediatorMoveRequest)
	router.HandleFunc("/map", a.handleGetGameContext).Methods(http.MethodGet)
	router.HandleFunc("/map", a.handleGameStartSignal).Methods(http.MethodPost)
	router.HandleFunc("/map", a.handleGameEndSignal).Methods(http.MethodDelete)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		a.Logger.Info("new request received", log.Fields{"method": r.Method, "request uri": r.RequestURI})
		router.ServeHTTP(w, r)
	})
}

func (a *Agent) ui(w http.ResponseWriter, r *http.Request) {
	state, err := a.MediatorClient.GameState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := web.MapTemplate.ExecuteTemplate(w, "map", state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Agent) handleMediatorMoveRequest(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// TODO: use the initial POST /map request + last moves data to avoid sending this request
	state, err := a.MediatorClient.GameState()
	if err != nil {
		a.Logger.Warn("failed to get game state", log.Fields{"error": err.Error()})
	}

	direction, err := a.MoveStrategy.DecideWhereToGo(strategyHelper{state: state, playerID: a.playerID})
	if err != nil {
		direction = mediator.None
		a.Logger.Warn("could not decide where to go", log.Fields{"error": err.Error()})
	}
	directionLogFields := log.Fields{"direction": direction}
	a.Logger.Info("about to send move request to the mediator", directionLogFields)

	moveRes, err := a.MediatorClient.Move(mediator.MoveOptions{
		ID:   id,
		Move: direction,
	})
	if err != nil {
		a.Logger.Warn("failed to send direction to the mediator", log.Fields{"error": err.Error()})
		return
	}
	if moveRes.Accepted {
		a.Logger.Info("move was accepted by the mediator", directionLogFields)
	} else {
		a.Logger.Warn("move was refused by the mediator", directionLogFields)
	}
}

func (a *Agent) handleGetGameContext(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	state, err := a.MediatorClient.GameState()
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(state)
}

func (a *Agent) handleGameStartSignal(w http.ResponseWriter, r *http.Request) {
	a.Logger.Info("received board information, game is about to start, fasten seat belts!", nil)
}

func (a *Agent) handleGameEndSignal(w http.ResponseWriter, r *http.Request) {
	a.Logger.Info("received game end signal from mediator", nil)
	if len(a.AppConfig.ManualExit) > 0 {
		a.Logger.Debug("client started with the manual exit option, will not automatically shutdown", nil)
		return
	}
	if a.StopCallback != nil {
		a.StopCallback()
		return
	}
	a.Logger.Warn("no stop callback configured, player web server won't automatically shutdown", nil)
}
