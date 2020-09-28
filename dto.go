package main

type resp struct {
	Data struct {
		Ranks []rank `json:"ranks"`
	} `json:"data"`
}

type rank struct {
	Rank      int  `json:"rank"`
	Score     int  `json:"score"`
	FirstGift bool `json:"first_gift"`
	User      User `json:"user"`
}

type User struct {
	ID   string `json:"id_str"`
	Name string `json:"nickname"`
}
