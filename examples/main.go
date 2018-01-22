package main

import (
	"go/build"
	"path/filepath"
	"sync"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/TheWeatherCompany/meteor"
	"github.ibm.com/TheWeatherCompany/wxapi/apis/sun/v1/forecast/daily"
	"github.ibm.com/TheWeatherCompany/wxapi/apis/sun/v3/alerts"
	"github.ibm.com/TheWeatherCompany/wxgo/types/cms"
)

var (
	Env             map[string]string
	meteorService   *meteor.Meteor
	loadServiceOnce sync.Once
)

// loadClient loads the client once.
func loadService() {
	meteorService = meteor.NewMeteor(meteor.NewCredentials(Env), nil)

}

// GetService gets the meteor.Service
func GetService() *meteor.Service {
	loadServiceOnce.Do(loadService)
	return meteorService.Common
}

func init() {
	p, _ := filepath.Abs(build.Default.GOPATH + "/src/github.ibm.com/TheWeatherCompany/meteor/.env")
	Env, _ = godotenv.Read(p)
}

const (
	sunV1API = "https://api.weather.com/v1/"
	sunV3API = "https://api.weather.com/v3/"
	dsxAPI = "https://dsx.weather.com/"
)

type Params struct {
	Key    string `url:"apiKey"`
	Format string `url:"format,omitempty"`
	Language string `url:"language,omitempty"`
	Units    string `url:"units,omitempty"`
	Geocode    string `url:"geocode,omitempty"`
}

type DSXParams struct {
	Key    string `url:"api"`
}

func doDailyForecast() {
	var sResp1 dailyforecast.DailyForecastResponse
	v1Base := GetService().New().Base(sunV1API).Client(nil)

	dailyForecastMeteor := v1Base.New().Pathf("geocode/%v/%v", "34.063", "-84.217").Pathf("forecast/daily/%vday.json", 3).Get().QueryStruct(&Params{
		Key: meteorService.GetCredBy("sun"),
		Language: "en-US",
		Units: "e",
	}).Responder(meteor.JSONSuccessResponder(&sResp1))

	req, _ := dailyForecastMeteor.Request()
	fmt.Printf("%v\n", req.URL.String())

	dailyForecastMeteor.Do(req)
	fmt.Printf("%#v\n", sResp1)
}

func doAlerts() {
	var sResp2 alerts.AlertsHeadlinesResponse

	v3Base := GetService().New().Base(sunV3API).Client(nil)
	alertsMeteor := v3Base.New().Path("alerts/headlines").Get().QueryStruct(&Params{
		Key: meteorService.GetCredBy("sun"),
		Language: "en-US",
		Geocode: "34.063,-84.217",
		Format:"json",
	}).Responder(meteor.JSONSuccessResponder(&sResp2))

	req, _ := alertsMeteor.Request()
	fmt.Printf("%v\n", req.URL.String())

	alertsMeteor.Do(req)
	fmt.Printf("%#v\n", sResp2)
}

func doBreakingNow() {
	var sResp3 cmstypes.BreakingNow

	dsxBase := GetService().New().Base(dsxAPI).Client(nil)
	breakingNowMeteor := dsxBase.New().Path("cms/v4/settings/en_US/breakingnow").Get().QueryStruct(&DSXParams{
		Key: meteorService.GetCredBy("dsx"),
	}).Responder(meteor.JSONSuccessResponder(&sResp3))

	req, _ := breakingNowMeteor.Request()
	fmt.Printf("%v\n", req.URL.String())

	breakingNowMeteor.Do(req)
	fmt.Printf("%#v\n", sResp3)
}

func main() {
	doBreakingNow()
}