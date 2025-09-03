package model


type MetadataUser struct {
	ID                 string    `json:"id"`
	FullName           string    `json:"full_name"`
	AvatarURL          string    `json:"avatar_url"`
	PhoneNumber        string    `json:"phone_number"`
	BirthDate          string `json:"birth_date"`
	LastUpdated        string `json:"last_updated"`
}