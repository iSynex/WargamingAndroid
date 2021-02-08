package EndPoints

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const ApplicationID string = "829a80f05f72847056c17df4c7df3caf"

type UserCredentials struct {
	Login    string
	Password string
}

type BearerToken struct {
	Access_token string `json:"access_token"`
}

type Account struct {
	Access_token string
	AccountID    string
}

type PlayerData struct {
	Data map[string]struct {
		Last_battle_time int64  `json:"last_battle_time"`
		Nickname         string `json:"nickname"`
		Logout_at        int    `json:"logout_at"`
		Created_at       int    `json:"created_at"`
		Clan_id          int    `json:"clan_id"`
		Updated_at       int    `json:"updated_at"`

		Statistics struct {
			All struct {
				Battles int `json:"battles"`
			} `json:"all"`
		} `json:"statistics"`

		Private struct {
			Credits            int  `json:"credits"`
			Free_xp            int  `json:"free_xp"`
			Gold               int  `json:"gold"`
			Bonds              int  `json:"bonds"`
			Is_bound_to_phone  bool `json:"is_bound_to_phone"`
			Is_premium         bool `json:"is_premium"`
			Premium_expires_at int  `json:"premium_expires_at"`
		} `json:"private"`
	} `json:"data"`
}

type TanksID struct {
	Data map[string][]struct {
		TankId int `json:"tank_id"`
	} `json:"data"`
}

type VehicleInfo struct {
	Is_premium bool   `json:"is_premium"`
	Name       string `json:"name"`
	Tier       int    `json:"tier"`
}

type VehicleBasic struct {
	Data map[int]VehicleInfo `json:"data"`
}

type GameSettings struct {
	Region struct {
		DomainURL string
		ApiDomain string
		ClientID  string
	}
	GameName string
}

var GamesMap = map[string]GameSettings{}

func (Credentials *UserCredentials) OpenToken(client *http.Client, Challenge int) (BearerToken, error) {

	PostData := url.Values{
		"grant_type": {"urn:wargaming:params:oauth:grant-type:basic"},
		"scope":      {"account.credentials.oauth_long_lived_token.create account.credentials.token1.create account.credentials.oauth_long_lived_token.create papi"},
		"username":   {Credentials.Login},
		"password":   {Credentials.Password},
		"client_id":  {"AYBV9XBPaQcJSXSfqko9yutYcU5RqK1bLzpZz2ef"},
		"pow":        {strconv.Itoa(Challenge)},
	}

	resp, err := client.PostForm("https://ru.wargaming.net/id/api/v2/account/credentials/create/oauth/token/?type=pow", PostData)
	if err != nil {
		return BearerToken{}, err
	}

	resp.Body.Close()

	resp, err = http.Get(resp.Header.Get("Location"))
	if err != nil {
		return BearerToken{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {

		buff, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return BearerToken{}, err
		}

		var BearerTokenResult BearerToken

		err = json.Unmarshal(buff, &BearerTokenResult)
		if err != nil {
			return BearerToken{}, err
		}

		return BearerTokenResult, nil
	} else {
		return BearerToken{}, fmt.Errorf("%s", "Invalid Credentials")
	}
}

func (token *BearerToken) GetAccessToken(client *http.Client) (Account, error) {

	PostData := url.Values{
		"display":        {"popup"},
		"application_id": {ApplicationID},
		"token":          {token.Access_token},
	}

	resp, err := client.PostForm("https://api.worldoftanks.ru/wot/auth/login/", PostData)
	if err != nil {
		return Account{}, err
	}

	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Account{}, err
	}

	resp.Body.Close()

	regex := regexp.MustCompile("<input[^>]*value=\"([^\"]*)\"[^>]*name=\"csrfmiddlewaretoken\"")
	CsrfToken := regex.FindStringSubmatch(string(buff))[1]

	PostData = url.Values{
		"csrfmiddlewaretoken": {CsrfToken},
		"allow":               {"%D0%9F%D0%BE%D0%B4%D1%82%D0%B2%D0%B5%D1%80%D0%B4%D0%B8%D1%82%D1%8C"},
	}

	req, err := http.NewRequest("POST", resp.Request.URL.String(), strings.NewReader(PostData.Encode()))
	if err != nil {
		return Account{}, err
	}
	req.Header.Set("Referer", "https://api.worldoftanks.ru/wot/auth/login/")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = client.Do(req)
	if err != nil {
		return Account{}, err
	}

	return Account{
		Access_token: resp.Request.URL.Query()["access_token"][0],
		AccountID:    resp.Request.URL.Query()["account_id"][0],
	}, nil
}

func (account *Account) GetPlayerInfo(client *http.Client, settings GameSettings) (PlayerData, error) {

	resp, err := client.Get(settings.Region.ApiDomain + "account/info/?account_id=" + account.AccountID + "&application_id=" + ApplicationID + "&access_token=" + account.Access_token)
	if err != nil {
		return PlayerData{}, err
	}

	defer resp.Body.Close()

	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PlayerData{}, err
	}

	var UserInfo PlayerData

	err = json.Unmarshal(buff, &UserInfo)
	if err != nil {
		return PlayerData{}, err
	}

	return UserInfo, nil
}

func (account *Account) GetPlayerVehicles(client *http.Client, settings GameSettings) ([]VehicleInfo, error) {

	resp, err := client.Get(settings.Region.ApiDomain + "tanks/stats/?account_id=" + account.AccountID + "&in_garage=1&application_id=" + ApplicationID + "&language=ru&fields=tank_id&access_token=" + account.Access_token)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var TanksArr TanksID

	err = json.Unmarshal(buff, &TanksArr)
	if err != nil {
		return nil, err
	}

	var Limit = 100

	tanksID := make([]string, 0)
	for _, v := range TanksArr.Data[account.AccountID] {
		tanksID = append(tanksID, strconv.Itoa(v.TankId))
	}

	TanksData := make([]VehicleInfo, 0)

	for i, j := 0, 0; i < len(tanksID); {

		j = i
		if i+Limit <= len(tanksID) {
			i += Limit
		} else {
			i += len(tanksID) - j
		}

		resp, err := client.Get(settings.Region.ApiDomain + "encyclopedia/vehicles/?application_id=" + ApplicationID + "&fields=name,tier,is_premium&language=ru&tank_id=" + strings.Join(tanksID[j:i], ","))
		if err != nil {
			return nil, err
		}

		buff, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		resp.Body.Close()

		var vehicleData VehicleBasic
		if err := json.Unmarshal(buff, &vehicleData); err != nil {
			return nil, err
		}

		for _, v := range vehicleData.Data {
			TanksData = append(TanksData, v)
		}
	}

	return TanksData, nil
}

func GetChallenge(client *http.Client) ([]byte, error) {

	req, err := http.NewRequest("GET", "https://ru.wargaming.net/id/api/v2/account/credentials/create/oauth/token/challenge/?type=pow", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return buff, nil
}
