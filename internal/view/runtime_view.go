package view

import (
	"context"

	"mockserver/internal/model/bo"
	"mockserver/internal/service"
)

type RuntimeView interface {
	MatchPublished(ctx context.Context, event bo.Event) (bo.SimulationResult, error)
	DecidePublished(ctx context.Context, event bo.Event) (bo.RuntimeDecision, error)
}

type runtimeView struct {
	service service.RuntimeService
}

func NewRuntimeView(runtimeService service.RuntimeService) RuntimeView {
	return &runtimeView{service: runtimeService}
}

func (v *runtimeView) MatchPublished(ctx context.Context, event bo.Event) (bo.SimulationResult, error) {
	return v.service.MatchPublished(ctx, event)
}

func (v *runtimeView) DecidePublished(ctx context.Context, event bo.Event) (bo.RuntimeDecision, error) {
	return v.service.DecidePublished(ctx, event)
}
