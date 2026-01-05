package resource

// Type represents a kind of resource
type Type int

const (
	Metal Type = iota
	Energy
	NumTypes // Keep this last - represents count of resource types
)

// String returns the display name of the resource type
func (t Type) String() string {
	switch t {
	case Metal:
		return "Metal"
	case Energy:
		return "Energy"
	default:
		return "Unknown"
	}
}

// Resource represents a single resource with current amount, capacity, and flow rates
type Resource struct {
	Current    float64 // Current amount stored
	Capacity   float64 // Maximum storage capacity
	Production float64 // Units produced per second
	Consumption float64 // Units consumed per second
}

// NetFlow returns the net change per second (production - consumption)
func (r *Resource) NetFlow() float64 {
	return r.Production - r.Consumption
}

// Update advances the resource by delta seconds
func (r *Resource) Update(delta float64) {
	r.Current += r.NetFlow() * delta

	// Clamp to bounds
	if r.Current < 0 {
		r.Current = 0
	}
	if r.Current > r.Capacity {
		r.Current = r.Capacity
	}
}

// Ratio returns current/capacity as a value from 0 to 1
func (r *Resource) Ratio() float64 {
	if r.Capacity <= 0 {
		return 0
	}
	return r.Current / r.Capacity
}

// CanAfford returns true if current amount >= cost
func (r *Resource) CanAfford(cost float64) bool {
	return r.Current >= cost
}

// Spend deducts an amount if affordable, returns true if successful
func (r *Resource) Spend(amount float64) bool {
	if !r.CanAfford(amount) {
		return false
	}
	r.Current -= amount
	return true
}

// Add adds an amount to current (clamped to capacity)
func (r *Resource) Add(amount float64) {
	r.Current += amount
	if r.Current > r.Capacity {
		r.Current = r.Capacity
	}
}

// Manager handles multiple resource types
type Manager struct {
	resources [NumTypes]*Resource
}

// NewManager creates a new resource manager with default values
// Initial capacity comes from constructor unit (1k metal, 2k energy)
func NewManager() *Manager {
	m := &Manager{}

	// Initialize metal (constructor provides 1000 storage)
	m.resources[Metal] = &Resource{
		Current:     500,
		Capacity:    1000,
		Production:  2,
		Consumption: 0,
	}

	// Initialize energy (constructor provides 2000 storage)
	m.resources[Energy] = &Resource{
		Current:     500,
		Capacity:    2000,
		Production:  10,
		Consumption: 0,
	}

	return m
}

// Get returns the resource of the given type
func (m *Manager) Get(t Type) *Resource {
	if t < 0 || t >= NumTypes {
		return nil
	}
	return m.resources[t]
}

// Update advances all resources by delta seconds
func (m *Manager) Update(delta float64) {
	for _, r := range m.resources {
		if r != nil {
			r.Update(delta)
		}
	}
}

// CanAfford checks if all costs can be afforded
func (m *Manager) CanAfford(costs map[Type]float64) bool {
	for t, cost := range costs {
		if r := m.Get(t); r == nil || !r.CanAfford(cost) {
			return false
		}
	}
	return true
}

// Spend deducts multiple costs atomically (all or nothing)
func (m *Manager) Spend(costs map[Type]float64) bool {
	if !m.CanAfford(costs) {
		return false
	}
	for t, cost := range costs {
		m.Get(t).Spend(cost)
	}
	return true
}

// AddProduction adds to a resource's production rate
func (m *Manager) AddProduction(t Type, amount float64) {
	if r := m.Get(t); r != nil {
		r.Production += amount
	}
}

// AddConsumption adds to a resource's consumption rate
func (m *Manager) AddConsumption(t Type, amount float64) {
	if r := m.Get(t); r != nil {
		r.Consumption += amount
	}
}

// SetProduction sets a resource's production rate
func (m *Manager) SetProduction(t Type, amount float64) {
	if r := m.Get(t); r != nil {
		r.Production = amount
	}
}

// SetConsumption sets a resource's consumption rate
func (m *Manager) SetConsumption(t Type, amount float64) {
	if r := m.Get(t); r != nil {
		r.Consumption = amount
	}
}

// AddCapacity increases a resource's storage capacity
func (m *Manager) AddCapacity(t Type, amount float64) {
	if r := m.Get(t); r != nil {
		r.Capacity += amount
	}
}

// SetCapacity sets a resource's storage capacity
func (m *Manager) SetCapacity(t Type, amount float64) {
	if r := m.Get(t); r != nil {
		r.Capacity = amount
		// Clamp current to new capacity if needed
		if r.Current > r.Capacity {
			r.Current = r.Capacity
		}
	}
}
