package brokerapi

// Service represents a service (of which there may be many variants-- "plans")
// offered by a service broker
type Service struct {
	Name           string   `json:"name"`
	ID             string   `json:"id"`
	Description    string   `json:"description"`
	Bindable       bool     `json:"bindable"`
	PlanUpdateable bool     `json:"plan_updateable, omitempty"`
	Tags           []string `json:"tags, omitempty"`
	Requires       []string `json:"requires, omitempty"`

	Metadata        interface{}   `json:"metadata, omitempty"`
	Plans           []ServicePlan `json:"plans"`
	DashboardClient interface{}   `json:"dashboard_client"`
}
