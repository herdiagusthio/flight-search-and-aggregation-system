package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockFlightProvider is a test implementation of FlightProvider.
type mockFlightProvider struct {
	name    string
	flights []Flight
	err     error
}

func (m *mockFlightProvider) Name() string {
	return m.name
}

func (m *mockFlightProvider) Search(ctx context.Context, criteria SearchCriteria) ([]Flight, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.flights, nil
}

func TestProviderRegistry(t *testing.T) {
	tests := []struct {
		name           string
		providers      []*mockFlightProvider
		wantCount      int
		wantNames      []string
		getByName      string
		wantGetResult  bool
	}{
		{
			name:           "empty registry",
			providers:      nil,
			wantCount:      0,
			wantNames:      []string{},
			getByName:      "garuda",
			wantGetResult:  false,
		},
		{
			name: "single provider",
			providers: []*mockFlightProvider{
				{name: "garuda"},
			},
			wantCount:     1,
			wantNames:     []string{"garuda"},
			getByName:     "garuda",
			wantGetResult: true,
		},
		{
			name: "multiple providers",
			providers: []*mockFlightProvider{
				{name: "garuda"},
				{name: "lionair"},
				{name: "airasia"},
			},
			wantCount:     3,
			wantNames:     []string{"garuda", "lionair", "airasia"},
			getByName:     "lionair",
			wantGetResult: true,
		},
		{
			name: "get non-existent provider",
			providers: []*mockFlightProvider{
				{name: "garuda"},
			},
			wantCount:     1,
			wantNames:     []string{"garuda"},
			getByName:     "nonexistent",
			wantGetResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewProviderRegistry()

			// Register providers
			for _, p := range tt.providers {
				registry.Register(p)
			}

			// Verify count
			all := registry.GetAll()
			assert.Len(t, all, tt.wantCount)

			// Verify names
			names := registry.Names()
			assert.Len(t, names, tt.wantCount)
			for _, wantName := range tt.wantNames {
				assert.Contains(t, names, wantName)
			}

			// Verify Get
			provider := registry.Get(tt.getByName)
			if tt.wantGetResult {
				assert.NotNil(t, provider)
				assert.Equal(t, tt.getByName, provider.Name())
			} else {
				assert.Nil(t, provider)
			}
		})
	}
}

func TestProviderRegistry_RegisterNil(t *testing.T) {
	registry := NewProviderRegistry()
	registry.Register(nil) // Should not panic

	all := registry.GetAll()
	assert.Len(t, all, 0)
}

func TestProviderRegistry_RegisterDuplicate(t *testing.T) {
	registry := NewProviderRegistry()

	provider1 := &mockFlightProvider{name: "garuda", flights: []Flight{{ID: "1"}}}
	provider2 := &mockFlightProvider{name: "garuda", flights: []Flight{{ID: "2"}}}

	registry.Register(provider1)
	registry.Register(provider2) // Should replace

	all := registry.GetAll()
	assert.Len(t, all, 1)

	provider := registry.Get("garuda")
	// The second registration should have replaced the first
	result, _ := provider.Search(context.Background(), SearchCriteria{})
	assert.Len(t, result, 1)
	assert.Equal(t, "2", result[0].ID)
}

func TestFlightProvider_Interface(t *testing.T) {
	// This test verifies that mockFlightProvider implements FlightProvider
	var _ FlightProvider = (*mockFlightProvider)(nil)
}

func TestMockFlightProvider_Search(t *testing.T) {
	tests := []struct {
		name         string
		provider     *mockFlightProvider
		criteria     SearchCriteria
		wantFlights  int
		wantErr      bool
	}{
		{
			name: "returns flights successfully",
			provider: &mockFlightProvider{
				name: "garuda",
				flights: []Flight{
					{ID: "1", FlightNumber: "GA-123"},
					{ID: "2", FlightNumber: "GA-456"},
				},
			},
			criteria:    SearchCriteria{Origin: "CGK", Destination: "DPS"},
			wantFlights: 2,
			wantErr:     false,
		},
		{
			name: "returns empty slice when no flights",
			provider: &mockFlightProvider{
				name:    "lionair",
				flights: []Flight{},
			},
			criteria:    SearchCriteria{Origin: "CGK", Destination: "DPS"},
			wantFlights: 0,
			wantErr:     false,
		},
		{
			name: "returns error when provider fails",
			provider: &mockFlightProvider{
				name: "airasia",
				err:  ErrProviderTimeout,
			},
			criteria:    SearchCriteria{Origin: "CGK", Destination: "DPS"},
			wantFlights: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			flights, err := tt.provider.Search(ctx, tt.criteria)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, flights, tt.wantFlights)
			}
		})
	}
}
