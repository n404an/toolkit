package env

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"os"
	"strconv"
	"time"

	"errors"
	"github.com/joho/godotenv"
)

// Load по-умолчанию загрузит .env из текущей папки
func Load(filepathes ...string) { load(filepathes...) }

func GetString(key, defaultVal string) string           { return get(key, defaultVal) }
func GetBool(key string, defaultVal bool) bool          { return getBool(key, defaultVal) }
func GetInt(key string, defaultVal int) int             { return getInt(key, defaultVal) }
func GetDoubleSlice(s string) ([][]string, error)       { return getDoubleSlice(s) }
func GetFloat64(key string, defaultVal float64) float64 { return getFloat64(key, defaultVal) }
func GetDuration(key string, defaultVal time.Duration) time.Duration {
	return getDuration(key, defaultVal)
}

func load(fp ...string) {
	if err := godotenv.Load(fp...); err != nil {
		errAndExit("[ERROR] load env", err)
	}
}

func get(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getDoubleSlice(s string) ([][]string, error) {
	pairs := get(s, "")

	if len(pairs) == 0 {
		return nil, errors.New("no data")
	}
	var slice [][]string
	if err := jsoniter.UnmarshalFromString(pairs, &slice); err != nil {
		return nil, err
	}
	return slice, nil
}

func getBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		data, err := strconv.ParseBool(value)
		if err != nil {
			return defaultVal
		}
		return data
	}
	return defaultVal
}

func getInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		data, err := strconv.Atoi(value)
		if err != nil {
			return defaultVal
		}
		return data
	}
	return defaultVal
}

func getFloat64(key string, defaultVal float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		data, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return defaultVal
		}
		return data
	}
	return defaultVal
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		data, err := time.ParseDuration(value)
		if err != nil {
			return defaultVal
		}
		return data
	}
	return defaultVal
}

func errAndExit(msg ...any) {
	_, _ = fmt.Fprintln(os.Stderr, msg...)
	os.Exit(1)
}
