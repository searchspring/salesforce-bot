package models

type AccountInfo struct {
	Website     string
	Manager     string
	Active      string
	Type        string
	MRR         float64
	FamilyMRR   float64
	Platform    string
	Integration string
	Provider    string
	SiteId      string
	City        string
	State       string
}

type PartnerInfo struct {
	Name               string
	Type               string
	Status             string
	OwnerID            string
	PartnerType        string
	SupportedPlatforms string
	PartnerTerms       string
	PartnerTermsNotes  string
}
