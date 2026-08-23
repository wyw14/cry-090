package bootstrap

import (
	"context"
	"time"

	"github.com/wyw14/cry-090/internal/application/ports"
	"github.com/wyw14/cry-090/internal/application/services"
	"github.com/wyw14/cry-090/internal/repository/memory"
	httptransport "github.com/wyw14/cry-090/internal/transport/http"
)

type App struct {
	Server *httptransport.Server
	Store  *memory.Store
	Clock  *memory.Clock
}

func NewApp() *App {
	store := memory.NewStore()
	clock := memory.NewClock(time.Now().UTC())
	ids := &memory.IDs{}
	repos := memory.Repositories{Store: store}
	deps := ports.Dependencies{Clock: clock, IDs: ids, Tx: memory.Tx{}, Users: repos.UsersRepo(), Events: repos.EventsRepo(), Needs: repos.NeedsRepo(), Invitations: repos.InvitationsRepo(), Registrations: repos.RegistrationsRepo(), Audits: repos.AuditsRepo(), Notifications: memory.Notifier{}}
	events := services.NewEventService(deps)
	matching := services.NewMatchingService(deps)
	invitations := services.NewInvitationService(deps)
	registrations := services.NewRegistrationService(deps)
	return &App{Server: httptransport.NewServer(events, matching, invitations, registrations), Store: store, Clock: clock}
}
func (a *App) Shutdown(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
