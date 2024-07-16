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
	ErrNotEnabled  = validation.NewGenericError("destination type is not enabled")
)

type Registry struct {
	providers map[string]Provider
	ids       []string
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

func (r *Registry) Provider(id string) Provider { return r.providers[id] }

func (r *Registry) LookupTypeName(ctx context.Context, typeID string) (string, error) {
	p := r.Provider(typeID)
	if p == nil {
		return "", ErrUnknownType
	}

	info, err := p.TypeInfo(ctx)
	if err != nil {
		return "", err
	}

	return info.Name, nil
}

func (r *Registry) RegisterProvider(ctx context.Context, p Provider) {
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

func (r *Registry) ValidateField(ctx context.Context, typeID, fieldID, value string) error {
	p := r.Provider(typeID)
	if p == nil {
		return ErrUnknownType
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
		ti.Type = id // ensure ID is set

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
