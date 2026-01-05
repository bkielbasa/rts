package resource

type Type int

const (
	Metal Type = iota
	Energy
	NumTypes
)

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

type Resource struct {
	Current           float64
	Capacity          float64
	Production        float64
	Consumption       float64
	ConstructionDrain float64
}

func (r *Resource) NetFlow() float64 {
	return r.Production - r.Consumption
}
func (r *Resource) Update(delta float64) {
	r.Current += r.NetFlow() * delta
	if r.Current < 0 {
		r.Current = 0
	}
	if r.Current > r.Capacity {
		r.Current = r.Capacity
	}
}
func (r *Resource) Ratio() float64 {
	if r.Capacity <= 0 {
		return 0
	}
	return r.Current / r.Capacity
}
func (r *Resource) CanAfford(cost float64) bool {
	return r.Current >= cost
}
func (r *Resource) Spend(amount float64) bool {
	if !r.CanAfford(amount) {
		return false
	}
	r.Current -= amount
	return true
}
func (r *Resource) SpendWithTracking(amount float64) bool {
	if amount <= 0 {
		return true
	}
	r.Current -= amount
	r.ConstructionDrain += amount
	return true
}
func (r *Resource) ResetDrain() {
	r.ConstructionDrain = 0
}
func (r *Resource) Add(amount float64) {
	r.Current += amount
	if r.Current > r.Capacity {
		r.Current = r.Capacity
	}
}

type Manager struct {
	resources [NumTypes]*Resource
}

func NewManager() *Manager {
	m := &Manager{}
	m.resources[Metal] = &Resource{
		Current:     500,
		Capacity:    1000,
		Production:  2,
		Consumption: 0,
	}
	m.resources[Energy] = &Resource{
		Current:     500,
		Capacity:    2000,
		Production:  10,
		Consumption: 0,
	}
	return m
}
func (m *Manager) Get(t Type) *Resource {
	if t < 0 || t >= NumTypes {
		return nil
	}
	return m.resources[t]
}
func (m *Manager) Update(delta float64) {
	for _, r := range m.resources {
		if r != nil {
			r.Update(delta)
		}
	}
}
func (m *Manager) ResetDrains() {
	for _, r := range m.resources {
		if r != nil {
			r.ResetDrain()
		}
	}
}
func (m *Manager) CanAfford(costs map[Type]float64) bool {
	for t, cost := range costs {
		if r := m.Get(t); r == nil || !r.CanAfford(cost) {
			return false
		}
	}
	return true
}
func (m *Manager) Spend(costs map[Type]float64) bool {
	if !m.CanAfford(costs) {
		return false
	}
	for t, cost := range costs {
		m.Get(t).Spend(cost)
	}
	return true
}
func (m *Manager) AddProduction(t Type, amount float64) {
	if r := m.Get(t); r != nil {
		r.Production += amount
	}
}
func (m *Manager) AddConsumption(t Type, amount float64) {
	if r := m.Get(t); r != nil {
		r.Consumption += amount
	}
}
func (m *Manager) SetProduction(t Type, amount float64) {
	if r := m.Get(t); r != nil {
		r.Production = amount
	}
}
func (m *Manager) SetConsumption(t Type, amount float64) {
	if r := m.Get(t); r != nil {
		r.Consumption = amount
	}
}
func (m *Manager) AddCapacity(t Type, amount float64) {
	if r := m.Get(t); r != nil {
		r.Capacity += amount
	}
}
func (m *Manager) SetCapacity(t Type, amount float64) {
	if r := m.Get(t); r != nil {
		r.Capacity = amount
		if r.Current > r.Capacity {
			r.Current = r.Capacity
		}
	}
}
