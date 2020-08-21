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
			Max      float64 `json:"max"`
			MaxStamp int64   `json:"max_stamp"`
			Min      float64 `json:"min"`
			MinStamp int64   `json:"min_stamp"`
			Original float64 `json:"original"`
			Current  float64 `json:"current"`
			Data     []struct {
				X int64   `json:"x"`
				Y float64 `json:"y"`
			} `json:"data"`
			Trend  int `json:"trend"`
			Period int `json:"period"`
		} `json:"series"`
	} `json:"data"`
}
