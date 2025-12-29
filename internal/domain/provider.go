package domain

import "context"

// FlightProvider defines the contract for airline flight data providers.
// Each provider adapter must implement this interface to be registered
// with the flight search use case.
//
// This interface follows the Dependency Inversion Principle - the use case layer
// depends on this abstraction rather than concrete implementations. Provider adapters
// implement this interface, allowing new providers to be added without modifying
// existing code.
//
// Implementations should:
//   - Return normalized Flight entities regardless of source format
//   - Respect context cancellation and timeout
//   - Return provider-specific errors wrapped appropriately
//   - Return an empty slice if no flights match (not an error)
//   - Only return errors for connection/parsing failures
type FlightProvider interface {
	// Name returns the unique identifier for this provider.
	// This identifier is used for logging, metrics, and result attribution.
	//
	// The name should be a lowercase, underscore-separated string.
	// Examples: "garuda_indonesia", "lion_air", "batik_air", "airasia"
	Name() string

	// Search queries the provider for available flights matching the criteria.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout. Implementations must respect
	//     context cancellation and return promptly when ctx.Done() is signaled.
	//   - criteria: Search parameters including origin, destination, date, etc.
	//     The criteria will already be validated before Search is called.
	//
	// Returns:
	//   - []Flight: Slice of normalized flight entities. Returns an empty slice
	//     if no flights match the criteria (this is not an error).
	//   - error: Only returned for operational failures such as:
	//     - Network/connection errors
	//     - Response parsing errors
	//     - Provider API errors
	//     - Context cancellation/timeout
	//
	// The returned flights should have their Provider field set to this provider's Name().
	Search(ctx context.Context, criteria SearchCriteria) ([]Flight, error)
}

// ProviderRegistry manages the collection of available flight providers.
// This is used by the use case layer to discover and query all registered providers.
type ProviderRegistry interface {
	// Register adds a new provider to the registry.
	// If a provider with the same name already exists, it will be replaced.
	Register(provider FlightProvider)

	// GetAll returns all registered providers.
	GetAll() []FlightProvider

	// Get returns a specific provider by name, or nil if not found.
	Get(name string) FlightProvider

	// Names returns the names of all registered providers.
	Names() []string
}

// providerRegistry is the default implementation of ProviderRegistry.
type providerRegistry struct {
	providers map[string]FlightProvider
}

// NewProviderRegistry creates a new provider registry.
func NewProviderRegistry() ProviderRegistry {
	return &providerRegistry{
		providers: make(map[string]FlightProvider),
	}
}

// Register adds a new provider to the registry.
func (r *providerRegistry) Register(provider FlightProvider) {
	if provider != nil {
		r.providers[provider.Name()] = provider
	}
}

// GetAll returns all registered providers.
func (r *providerRegistry) GetAll() []FlightProvider {
	result := make([]FlightProvider, 0, len(r.providers))
	for _, p := range r.providers {
		result = append(result, p)
	}
	return result
}

// Get returns a specific provider by name, or nil if not found.
func (r *providerRegistry) Get(name string) FlightProvider {
	return r.providers[name]
}

// Names returns the names of all registered providers.
func (r *providerRegistry) Names() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
