package companies

// Company represents a company.
type Company struct {
	Name        string   `json:"name"`
	FoundedYear int      `json:"founded_year"`
	Industry    string   `json:"industry"`
	Revenue     int      `json:"revenue"`
	Employees   int      `json:"employees"`
	Locations   []string `json:"locations"`
	TechStack   []string `json:"tech_stack"`
}

// Repository provides access to a Company store.
type Repository interface {
}
