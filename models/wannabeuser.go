package models

// WannabeUser:
type WannabeUser struct {
	Address            string `json:"address"`
	City               string `json:"city"`
	Country            string `json:"country"`
	Email              string `json:"email"`
	Lang               string `json:"lang"`
	Name               string `json:"name"`
	Password           string `json:"password"`
	Phone              string `json:"phone"`
	SubscriptionSource string `json:"subscriptionSource"`
	Terms              bool   `json:"terms"`
	Zipcode            string `json:"zipcode"`
}
