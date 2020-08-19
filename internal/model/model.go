package model

// Kratos hello kratos.
type Kratos struct {
	Hello string
}

type Article struct {
	ID      int64
	Content string
	Author  string
}

type PriceTrendResp struct {
	Code int `json:"code"`
	Data struct {
		Series []struct {
			Max      int `json:"max"`
			MaxStamp int `json:"max_stamp"`
			Min      int `json:"min"`
			MinStamp int `json:"min_stamp"`
			Original int `json:"original"`
			Current  int `json:"current"`
			Data     []struct {
				X int `json:"x"`
				Y int `json:"y"`
			} `json:"data"`
			Trend  int `json:"trend"`
			Period int `json:"period"`
		} `json:"series"`
	} `json:"data"`
}
