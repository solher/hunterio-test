package people

// Person represents a person.
type Person struct {
	FullName string  `json:"full_name"`
	JobTitle string  `json:"job_title"`
	Contact  Contact `json:"contact"`
}

// Contact represents a person's contact information.
type Contact struct {
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	LinkedinURL  string `json:"linkedin_url"`
	XURL         string `json:"x_url"`
	InstagramURL string `json:"instagram_url"`
	FacebookURL  string `json:"facebook_url"`
}
