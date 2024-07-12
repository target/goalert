package nfydest

import (
	"context"
	"errors"
	"fmt"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/validation"
)

var (
	ErrUnknownType = validation.NewGenericError("unknown destination type")
	ErrUnsupported = errors.New("unsupported operation")
)

type Registry struct {
	providers map[string]Provider
	ids       []string
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Provider(id string) Provider {
	if r.providers == nil {
		return nil
	}

	return r.providers[id]
}

func (r *Registry) RegisterProvider(ctx context.Context, p Provider) {
	if r.providers == nil {
		r.providers = make(map[string]Provider)
	}
	if r.Provider(p.ID()) != nil {
		panic(fmt.Sprintf("provider with ID %s already registered", p.ID()))
	}

	id := p.ID()
	r.providers[id] = p
	r.ids = append(r.ids, id)
}

func (r *Registry) DisplayInfo(ctx context.Context, d gadb.DestV1) (*DisplayInfo, error) {
	p := r.Provider(d.Type)
	if p == nil {
		return nil, ErrUnknownType
	}

	return p.DisplayInfo(ctx, d.Args)
}

func (r *Registry) ValidateField(ctx context.Context, typeID, fieldID, value string) (bool, error) {
	p := r.Provider(typeID)
	if p == nil {
		return false, ErrUnknownType
	}

	return p.ValidateField(ctx, fieldID, value)
}

func (r *Registry) Types(ctx context.Context) ([]TypeInfo, error) {
	var out []TypeInfo
	for _, id := range r.ids {
		ti, err := r.providers[id].TypeInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("get type info for %s: %w", id, err)
		}

		out = append(out, *ti)
	}

	return out, nil
}

func (r *Registry) SearchField(ctx context.Context, typeID, fieldID string, options SearchOptions) (*SearchResult, error) {
	p := r.Provider(typeID)
	if p == nil {
		return nil, ErrUnknownType
	}

	s, ok := p.(FieldSearcher)
	if !ok {
		return nil, fmt.Errorf("provider %s does not support field searching: %w", typeID, ErrUnsupported)
	}

	return s.SearchField(ctx, fieldID, options)
}

func (r *Registry) FieldLabel(ctx context.Context, typeID, fieldID, value string) (string, error) {
	p := r.Provider(typeID)
	if p == nil {
		return "", ErrUnknownType
	}

	s, ok := p.(FieldSearcher)
	if !ok {
		return "", fmt.Errorf("provider %s does not support field searching: %w", typeID, ErrUnsupported)
	}

	return s.FieldLabel(ctx, fieldID, value)
}
