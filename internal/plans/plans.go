package plans

// PlanLimits defines the limits for each subscription plan
type PlanLimits struct {
	Name        string `json:"name"`
	MaxBackends int    `json:"max_backends"` // -1 = unlimited
	RateLimit   int    `json:"rate_limit"`   // requests per minute, -1 = unlimited
	Analytics   bool   `json:"analytics"`
	Alerts      bool   `json:"alerts"`
	Price       int    `json:"price"` // USD per month
}

// Plans contains all available subscription plans
var Plans = map[string]PlanLimits{
	"free": {
		Name:        "Free",
		MaxBackends: 2,
		RateLimit:   100,
		Analytics:   false,
		Alerts:      false,
		Price:       0,
	},
	"pro": {
		Name:        "Pro",
		MaxBackends: 10,
		RateLimit:   1000,
		Analytics:   true,
		Alerts:      true,
		Price:       19,
	},
	"enterprise": {
		Name:        "Enterprise",
		MaxBackends: -1, // Unlimited
		RateLimit:   -1, // Unlimited
		Analytics:   true,
		Alerts:      true,
		Price:       99,
	},
}

// GetPlan returns the plan limits for a given plan name
func GetPlan(planName string) PlanLimits {
	if plan, ok := Plans[planName]; ok {
		return plan
	}
	return Plans["free"] // Default to free
}

// CanAddBackend checks if a user can add more backends
func CanAddBackend(planName string, currentCount int) bool {
	plan := GetPlan(planName)
	if plan.MaxBackends == -1 {
		return true // Unlimited
	}
	return currentCount < plan.MaxBackends
}

// HasAnalytics checks if a plan has analytics access
func HasAnalytics(planName string) bool {
	return GetPlan(planName).Analytics
}

// GetAllPlans returns all available plans for the pricing page
func GetAllPlans() []PlanLimits {
	return []PlanLimits{
		Plans["free"],
		Plans["pro"],
		Plans["enterprise"],
	}
}
