package people

// Person represents a person.
type Person struct {
	FullName string  `json:"full_name"`
	JobTitle string  `json:"job_title"`
	Contact  Contact `json:"contact"`
}

// Contact represents a person's contact information.
type Contact struct {
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	LinkedinURL string `json:"linkedin_url"`
}

// Repository provides access to a Person store.
type Repository interface {
}
