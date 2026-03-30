package model

type ThemeSubtheme struct {
	ThemeID    int  `json:"themeId"`
	SubthemeID int  `json:"subthemeId"`
	ForVip     bool `json:"forVip"`
}

type Team struct {
	ID            int             `json:"id"`
	Name          string          `json:"name"`
	ThemeSubtheme []ThemeSubtheme `json:"themeSubtheme"`
}
