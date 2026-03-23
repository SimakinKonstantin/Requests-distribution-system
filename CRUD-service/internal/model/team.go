package model

type ThemeSubtheme struct {
	ThemeID    int  `json:"theme_id"`
	SubthemeID int  `json:"subtheme_id"`
	ForVip     bool `json:"for_vip"`
}

type Team struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	ThemeSubtheme []ThemeSubtheme
}
