package main

import (
	"EndPoints"
	"PoW"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"sort"
	"strconv"
	"time"
)

const (
	DebugError int = 0x1
	Critical   int = 0x2
)

var ErrorHandleMode = DebugError

func main() {

	Login("LOGIN HERE", "PASSWORD HERE")

}

func Login(login, password string) {

	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: cookieJar}

	var Pow PoW.PowChallenge

	Challenge, err := EndPoints.GetChallenge(client)
	if err != nil {
		ErrorHandle(err)
		return
	}

	if err := json.Unmarshal(Challenge, &Pow); err != nil {
		return
	}

	PowResult := Pow.ResolveChallenge()

	var UserData = EndPoints.UserCredentials{
		Login:    login,
		Password: password,
	}

	Bearer, err := UserData.OpenToken(client, PowResult)
	if err != nil {
		ErrorHandle(err)
		return
	}

	LoggedUser, err := Bearer.GetAccessToken(client)
	if err != nil {
		ErrorHandle(err)
		return
	}

	for _, GameSettings := range EndPoints.GamesMap {

		UserInfo, err := LoggedUser.GetPlayerInfo(client, GameSettings)
		if err != nil {
			ErrorHandle(err)
			return
		}

		Vehicles, err := LoggedUser.GetPlayerVehicles(client, GameSettings)
		if err != nil {
			ErrorHandle(err)
			return
		}

		CreateLog(UserInfo, Vehicles, LoggedUser.AccountID, UserData)
	}

}

func ErrorHandle(err error) {

	switch ErrorHandleMode {

	case DebugError:
		log.Println(err)
		break

	case Critical:
		panic(err)
	}
}

func ParseUnixTime(unixTime int64) time.Time {
	return time.Unix(unixTime, 0)
}

func CreateLog(userInfo EndPoints.PlayerData, vehicles []EndPoints.VehicleInfo, accountID string, credentials EndPoints.UserCredentials) {

	var buffer bytes.Buffer
	buffer.WriteString("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	buffer.WriteString("\nNick: " + userInfo.Data[accountID].Nickname)
	buffer.WriteString("\nCredentials: " + credentials.Login + ":" + credentials.Password)
	buffer.WriteString("\nLast Battle: " + ParseUnixTime(userInfo.Data[accountID].Last_battle_time).Format("2006-01-02"))
	buffer.WriteString("\nGold: " + strconv.Itoa(userInfo.Data[accountID].Private.Gold) + " | " + "Silver: " + strconv.Itoa(userInfo.Data[accountID].Private.Credits))
	buffer.WriteString("\nBattles: " + strconv.Itoa(userInfo.Data[accountID].Statistics.All.Battles))
	buffer.WriteString("\n~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")

	sort.SliceStable(vehicles, func(i, j int) bool {
		return vehicles[i].Tier > vehicles[j].Tier
	})

	for _, v := range vehicles {

		if len(v.Name) > 0 {
			buffer.WriteString("\n[" + strconv.Itoa(v.Tier) + "] => " + v.Name)
			if v.Is_premium {
				buffer.WriteString(" [PREMIUM]")
			}
		}
	}

	buffer.WriteString("\n~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n\n")

	fmt.Println(buffer.String())
}
