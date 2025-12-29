package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProviderRegistry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		providerNames []string
		wantCount     int
		getByName     string
		wantGetResult bool
	}{
		{
			name:          "empty registry",
			providerNames: nil,
			wantCount:     0,
			getByName:     "garuda",
			wantGetResult: false,
		},
		{
			name:          "single provider",
			providerNames: []string{"garuda"},
			wantCount:     1,
			getByName:     "garuda",
			wantGetResult: true,
		},
		{
			name:          "multiple providers",
			providerNames: []string{"garuda", "lionair", "airasia"},
			wantCount:     3,
			getByName:     "lionair",
			wantGetResult: true,
		},
		{
			name:          "get non-existent provider",
			providerNames: []string{"garuda"},
			wantCount:     1,
			getByName:     "nonexistent",
			wantGetResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewProviderRegistry()

			// Register mock providers
			for _, name := range tt.providerNames {
				mock := NewMockFlightProvider(ctrl)
				mock.EXPECT().Name().Return(name).AnyTimes()
				registry.Register(mock)
			}

			// Verify count
			all := registry.GetAll()
			assert.Len(t, all, tt.wantCount)

			// Verify names
			names := registry.Names()
			assert.Len(t, names, tt.wantCount)
			for _, wantName := range tt.providerNames {
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	registry := NewProviderRegistry()

	// First provider with ID "1"
	provider1 := NewMockFlightProvider(ctrl)
	provider1.EXPECT().Name().Return("garuda").AnyTimes()
	provider1.EXPECT().Search(gomock.Any(), gomock.Any()).Return([]Flight{{ID: "1"}}, nil).AnyTimes()

	// Second provider with ID "2" (should replace first)
	provider2 := NewMockFlightProvider(ctrl)
	provider2.EXPECT().Name().Return("garuda").AnyTimes()
	provider2.EXPECT().Search(gomock.Any(), gomock.Any()).Return([]Flight{{ID: "2"}}, nil).AnyTimes()

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// This test verifies that MockFlightProvider implements FlightProvider
	var _ FlightProvider = NewMockFlightProvider(ctrl)
}

func TestMockFlightProvider_Search(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("returns flights successfully", func(t *testing.T) {
		mock := NewMockFlightProvider(ctrl)
		mock.EXPECT().Name().Return("garuda").AnyTimes()
		mock.EXPECT().Search(gomock.Any(), gomock.Any()).Return([]Flight{
			{ID: "1", FlightNumber: "GA-123"},
			{ID: "2", FlightNumber: "GA-456"},
		}, nil)

		ctx := context.Background()
		flights, err := mock.Search(ctx, SearchCriteria{Origin: "CGK", Destination: "DPS"})

		assert.NoError(t, err)
		assert.Len(t, flights, 2)
	})

	t.Run("returns empty slice when no flights", func(t *testing.T) {
		mock := NewMockFlightProvider(ctrl)
		mock.EXPECT().Name().Return("lionair").AnyTimes()
		mock.EXPECT().Search(gomock.Any(), gomock.Any()).Return([]Flight{}, nil)

		ctx := context.Background()
		flights, err := mock.Search(ctx, SearchCriteria{Origin: "CGK", Destination: "DPS"})

		assert.NoError(t, err)
		assert.Len(t, flights, 0)
	})

	t.Run("returns error when provider fails", func(t *testing.T) {
		mock := NewMockFlightProvider(ctrl)
		mock.EXPECT().Name().Return("airasia").AnyTimes()
		mock.EXPECT().Search(gomock.Any(), gomock.Any()).Return(nil, ErrProviderTimeout)

		ctx := context.Background()
		flights, err := mock.Search(ctx, SearchCriteria{Origin: "CGK", Destination: "DPS"})

		assert.Error(t, err)
		assert.Nil(t, flights)
	})
}
