package main

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"hrig/shared/model"
)

type App struct {
	mu      sync.Mutex
	ctx     context.Context
	client  *apiClient
	initErr error
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewApp() *App {
	client, err := newAPIClient(os.Getenv("HRIG_API_URL"), nil)
	return &App{
		client:  client,
		initErr: err,
	}
}

func (a *App) SetContext(ctx context.Context) {
	a.mu.Lock()
	a.ctx = ctx
	a.mu.Unlock()
}

func (a *App) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	a.mu.Lock()
	a.ctx = ctx
	if a.client == nil || a.initErr != nil {
		a.mu.Unlock()
		return nil
	}
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
	a.client.SetEmitFn(defaultEmit)
	wsCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel
	a.wg.Add(1)
	client := a.client
	a.mu.Unlock()

	go func() {
		defer a.wg.Done()
		client.RunRealtimeLoop(wsCtx)
	}()

	return nil
}

func (a *App) ServiceShutdown() error {
	a.mu.Lock()
	cancel := a.cancel
	a.cancel = nil
	a.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	a.wg.Wait()
	return nil
}

func (a *App) Ping() string {
	return "pong"
}

func (a *App) GetConnectionStatus() ConnectionState {
	if a.client == nil {
		errMsg := ""
		if a.initErr != nil {
			errMsg = a.initErr.Error()
		}
		return ConnectionState{
			Connected:    false,
			LastError:    errMsg,
			UpdatedAtUTC: time.Now().UTC(),
		}
	}
	return a.client.ConnectionStatus()
}

func (a *App) GetHealth() (model.HealthStatus, error) {
	client, err := a.ensureClient()
	if err != nil {
		return model.HealthStatus{}, err
	}
	ctx, cancel := a.rpcContext()
	defer cancel()
	return client.GetHealth(ctx)
}

func (a *App) GetAstronauts() ([]model.Astronaut, error) {
	client, err := a.ensureClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := a.rpcContext()
	defer cancel()
	return client.GetAstronauts(ctx)
}

func (a *App) GetMetrics() ([]model.Metric, error) {
	client, err := a.ensureClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := a.rpcContext()
	defer cancel()
	return client.GetMetrics(ctx)
}

func (a *App) GetSnapshot() (model.Snapshot, error) {
	client, err := a.ensureClient()
	if err != nil {
		return model.Snapshot{}, err
	}
	ctx, cancel := a.rpcContext()
	defer cancel()
	return client.GetSnapshot(ctx)
}

func (a *App) GetTrends(query model.TrendQuery) ([]model.TrendPoint, error) {
	client, err := a.ensureClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := a.rpcContext()
	defer cancel()
	return client.GetTrends(ctx, query)
}

func (a *App) GetHistory(query model.HistoryQuery) ([]model.HistoryEntry, error) {
	client, err := a.ensureClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := a.rpcContext()
	defer cancel()
	return client.GetHistory(ctx, query)
}

func (a *App) AckAlert(alertID string, acknowledgedBy string) (model.AlertEvent, error) {
	client, err := a.ensureClient()
	if err != nil {
		return model.AlertEvent{}, err
	}
	ctx, cancel := a.rpcContext()
	defer cancel()
	return client.AckAlert(ctx, alertID, acknowledgedBy)
}

func (a *App) GetHazards(q string) ([]model.SpaceHazard, error) {
	client, err := a.ensureClient()
	if err != nil {
		return []model.SpaceHazard{}, err
	}
	ctx, cancel := a.rpcContext()
	defer cancel()
	hazards, err := client.GetHazards(ctx, q)
	if err != nil {
		return []model.SpaceHazard{}, err
	}
	if hazards == nil {
		return []model.SpaceHazard{}, nil
	}
	return hazards, nil
}

func (a *App) GetProcedures(q string, category string, metricCode string) ([]model.MedicalProcedure, error) {
	client, err := a.ensureClient()
	if err != nil {
		return []model.MedicalProcedure{}, err
	}
	ctx, cancel := a.rpcContext()
	defer cancel()
	procedures, err := client.GetProcedures(ctx, q, category, metricCode)
	if err != nil {
		return []model.MedicalProcedure{}, err
	}
	if procedures == nil {
		return []model.MedicalProcedure{}, nil
	}
	return procedures, nil
}

func (a *App) Search(q string, tab string) (model.SearchResults, error) {
	client, err := a.ensureClient()
	if err != nil {
		var empty model.SearchResults
		empty.Normalize()
		return empty, err
	}
	ctx, cancel := a.rpcContext()
	defer cancel()
	results, err := client.Search(ctx, q, tab)
	if err != nil {
		var empty model.SearchResults
		empty.Normalize()
		return empty, err
	}
	results.Normalize()
	return results, nil
}

func (a *App) ensureClient() (*apiClient, error) {
	if a.initErr != nil {
		return nil, a.initErr
	}
	if a.client == nil {
		return nil, errors.New("api client is not initialized")
	}
	return a.client, nil
}

func (a *App) rpcContext() (context.Context, context.CancelFunc) {
	a.mu.Lock()
	base := a.ctx
	a.mu.Unlock()
	if base == nil {
		base = context.Background()
	}
	return context.WithTimeout(base, defaultHTTPTimeout)
}
