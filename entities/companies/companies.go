package companies

// Company represents a company.
type Company struct {
	ID          uint64   `json:"id"`
	Name        string   `json:"name"`
	FoundedYear int      `json:"founded_year"`
	Industry    string   `json:"industry"`
	Revenue     int      `json:"revenue"`
	Employees   int      `json:"employees"`
	Locations   []string `json:"locations"`
	TechStack   []string `json:"tech_stack"`
}
