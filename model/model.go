package model

type User struct {
	// UserID аккаунта
	AccountID int `json:"account_id" pg:",pk"`
	// Имя аккаунта
	AccountName string `json:"account_name" pg:",unique"`
	// Никнеймы
	AccountNames []string `json:"account_names"`
	// Аватарка
	Avatar string `json:"avatar"`
	// Фон
	Background string `json:"background"`
	// VIP статус
	VIP string `json:"vip"`
	// Рейтинг Social Credits
	SocialCredits float64 `json:"social_credits"`
	// Количество убийств
	Kills int `json:"kills"`
	// Количество смертей
	Deaths int `json:"deaths"`
	// Рейтинг CopChase
	CopChaseRating int `json:"cop_chase_rating"`
	// Ограничения
	Punishments []string `json:"punishments"`
	// Подтверждение аккаунта
	Verification string `json:"verification"`
	// Достижение
	Achievement string `json:"achievement"`
	// Телеграм
	Telegram string `json:"telegram"`
	// Префикс
	Prefix string `json:"prefix"`
	// Звезда
	Star string `json:"star"`
	// Application verification
	ApplicationVerification string `json:"application_verification"`
}

type ApiUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Access    int    `json:"access"`
	Moder     int    `json:"moder"`
	Verify    int    `json:"verify"`
	VerifyTxt string `json:"verifyText"`
	Mute      int    `json:"mute"`
	Online    int    `json:"online"`
	PlayerID  int    `json:"playerid"`
	RegDate   string `json:"regdate"`
	LastLogin string `json:"lastlogin"`
	Warn      []Warn `json:"warn"`
}

type Warn struct {
	Reason  string `json:"reason"`
	Admin   string `json:"admin"`
	BanTime string `json:"bantime"`
}

type LongUser struct {
	ID                      int      `json:"id"`
	Login                   string   `json:"login"`
	Access                  int      `json:"access"`
	Moder                   int      `json:"moder"`
	Verify                  int      `json:"verify"`
	VerifyTxt               string   `json:"verifyText"`
	Mute                    int      `json:"mute"`
	Online                  int      `json:"online"`
	PlayerID                int      `json:"playerid"`
	RegDate                 string   `json:"regdate"`
	LastLogin               string   `json:"lastlogin"`
	Warn                    []Warn   `json:"warn"`
	Avatar                  string   `json:"avatar"`
	Background              string   `json:"background"`
	VIP                     string   `json:"vip"`
	SocialCredits           float64  `json:"social_credits"`
	Kills                   int      `json:"kills"`
	Deaths                  int      `json:"deaths"`
	CopChaseRating          int      `json:"cop_chase_rating"`
	Punishments             []string `json:"punishments"`
	Achievement             string   `json:"achievement"`
	Telegram                string   `json:"telegram"`
	Prefix                  string   `json:"prefix"`
	Star                    string   `json:"star"`
	ApplicationVerification string   `json:"application_verification"`
}
