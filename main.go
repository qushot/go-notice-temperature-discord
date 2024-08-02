package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func getAmedasCode(enName string) (string, error) {
	res, err := http.Get("https://www.jma.go.jp/bosai/amedas/const/amedastable.json")
	if err != nil {
		slog.Error("http get error", slog.String("error", err.Error()))
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("io read all error", slog.String("error", err.Error()))
		return "", err
	}

	var some any
	if err := json.Unmarshal(body, &some); err != nil {
		slog.Error("json unmarshal error", slog.String("error", err.Error()))
		return "", err
	}

	for k, v := range some.(map[string]interface{}) {
		if v.(map[string]interface{})["enName"] == enName {
			return k, nil
		}
	}

	err = errors.New("no result error")
	slog.Error(err.Error(), slog.String("enName", enName))
	return "", err
}

func getAmedasTemperatureData(amedasCode string, now time.Time) (float64, error) {
	param := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()-(now.Minute()%10), 0, 0, now.Location()).
		Add(-20 * time.Minute).  // データは10分ごとに作成されるが、タイミングが謎なので20分引く
		Format("20060102150400") // 2006-01-02 15:06:00

	res, err := http.Get(fmt.Sprintf("https://www.jma.go.jp/bosai/amedas/data/map/%s.json", param))
	if err != nil {
		slog.Error("http get error", slog.String("error", err.Error()))
		return 0, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("io read all error", slog.String("error", err.Error()))
		return 0, err
	}

	var some any
	if err := json.Unmarshal(body, &some); err != nil {
		slog.Error("json unmarshal error", slog.String("error", err.Error()))
		return 0, err
	}

	for k, v := range some.(map[string]interface{}) {
		if k == amedasCode {
			m, ok := v.(map[string]interface{})
			if !ok {
				err := errors.New("cannot assertion to 'map[string]interface{}'")
				slog.Error("type assertion error", slog.String("error", err.Error()), slog.Bool("isOK", ok))
				return 0, err
			}

			is, ok := m["temp"]
			if !ok {
				err := errors.New("key not found")
				slog.Error("map error", slog.String("error", err.Error()), slog.Bool("isOK", ok))
				return 0, err
			}

			ts, ok := is.([]interface{})
			if !ok || len(ts) < 1 {
				err := errors.New("cannot assertion to '[]interface{}' or length is 0")
				slog.Error("type assertion error", slog.String("error", err.Error()), slog.Int("length", len(ts)), slog.Bool("isOK", ok))
				return 0, err
			}

			t, ok := ts[0].(float64)
			if !ok {
				err := errors.New("cannot assertion to 'float64'")
				slog.Error("type assertion error", slog.String("error", err.Error()), slog.Bool("isOK", ok))
				return 0, err
			}

			return t, nil
		}
	}

	err = errors.New("no result error")
	slog.Error(err.Error(), slog.String("amedasCode", amedasCode), slog.Time("now", now))
	return 0, err
}

func sendToDiscord(temperature float64, url string) error {
	type body struct {
		UserName string `json:"username"`
		Content  string `json:"content"`
	}

	rb := &body{
		UserName: "Temperaturebot",
		Content:  fmt.Sprintf("只今の気温は %g ℃です", temperature),
	}

	b, err := json.Marshal(rb)
	if err != nil {
		slog.Error("json marshal error", slog.String("error", err.Error()))
		return err
	}

	if _, err := http.Post(url, "application/json", bytes.NewReader(b)); err != nil {
		slog.Error("http post error", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func main() {
	amedasCode, err := getAmedasCode("Yokohama")
	if err != nil {
		slog.Error("failed to get amedas code", slog.String("error", err.Error()))
		os.Exit(1)
	}

	temperature, err := getAmedasTemperatureData(amedasCode, time.Now())
	if err != nil {
		slog.Error("failed to get amedas temperature data", slog.String("error", err.Error()))
		os.Exit(1)
	}

	key := "NOTICE_TEMPERATURE_DISCORD_URL"
	url, ok := os.LookupEnv(key)
	if !ok {
		slog.Error("failed to get environment variables", slog.String("key", key))
		os.Exit(1)
	}

	if err := sendToDiscord(temperature, url); err != nil {
		slog.Error("failed to send to discord", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
