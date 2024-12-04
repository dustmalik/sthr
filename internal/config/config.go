package config

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/MunifTanjim/stremthru/core"
)

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

type StoreAuthTokenMap map[string]map[string]string

func (m StoreAuthTokenMap) GetToken(user, store string) string {
	if um, ok := m[user]; ok {
		if token, ok := um[store]; ok {
			return token
		}
	}
	if user != "*" {
		return m.GetToken("*", store)
	}
	return ""
}

func (m StoreAuthTokenMap) setToken(user, store, token string) {
	if _, ok := m[user]; !ok {
		m[user] = make(map[string]string)
	}
	m[user][store] = token
}

func (m StoreAuthTokenMap) GetPreferredStore(user string) string {
	store := m.GetToken(user, "*")
	if store == "" {
		store = m.GetToken("*", "*")
	}
	return store
}

func (m StoreAuthTokenMap) ListStores(user string) []string {
	names := []string{}
	if um, ok := m[user]; ok {
		for name := range um {
			if name != "*" {
				names = append(names, name)
			}
		}
	}
	return names
}

func (m StoreAuthTokenMap) setPreferredStore(user, store string) {
	if m.GetPreferredStore(user) == "" {
		m.setToken(user, "*", store)
	}
}

type ProxyAuthPasswordMap map[string]string

func (m ProxyAuthPasswordMap) GetPassword(userName string) string {
	if token, ok := m[userName]; ok {
		return token
	}
	return ""
}

type Config struct {
	Port              string
	StoreAuthToken    StoreAuthTokenMap
	ProxyAuthPassword ProxyAuthPasswordMap
	BuddyURL          string
	BuddyAuthToken    string
	RedisURI          string
	DatabaseURI       string
}

var config = func() Config {
	if value := getEnv("STREMTHRU_HTTP_PROXY", ""); len(value) > 0 {
		if err := os.Setenv("HTTP_PROXY", value); err != nil {
			log.Fatal("failed to set http proxy")
		}
	}

	if value := getEnv("STREMTHRU_HTTPS_PROXY", ""); len(value) > 0 {
		if err := os.Setenv("HTTPS_PROXY", value); err != nil {
			log.Fatal("failed to set https proxy")
		}
	}

	proxyAuthCredList := strings.FieldsFunc(getEnv("STREMTHRU_PROXY_AUTH", ""), func(c rune) bool {
		return c == ','
	})
	proxyAuthPasswordMap := make(ProxyAuthPasswordMap)
	for _, cred := range proxyAuthCredList {
		if strings.ContainsRune(cred, ':') {
			username, password, ok := strings.Cut(cred, ":")
			if ok {
				proxyAuthPasswordMap[username] = password
			}
		} else if decoded, err := core.Base64Decode(cred); err == nil {
			username, password, ok := strings.Cut(strings.TrimSpace(decoded), ":")
			if ok {
				proxyAuthPasswordMap[username] = password
			}
		}
	}

	storeAlldebridTokenList := strings.FieldsFunc(getEnv("STREMTHRU_STORE_AUTH", ""), func(c rune) bool {
		return c == ','
	})
	storeAuthTokenMap := make(StoreAuthTokenMap)
	for _, userStoreToken := range storeAlldebridTokenList {
		if user, storeToken, ok := strings.Cut(userStoreToken, ":"); ok {
			if store, token, ok := strings.Cut(storeToken, ":"); ok {
				storeAuthTokenMap.setPreferredStore(user, store)
				storeAuthTokenMap.setToken(user, store, token)
			}
		}
	}

	buddyUrl := ""
	buddyAuthToken := ""
	if buddyUri := getEnv("STREMTHRU_BUDDY_URI", ""); buddyUri != "" {
		u, err := url.Parse(buddyUri)
		if err != nil {
			log.Fatal("invalid buddy uri")
		}
		if password, ok := u.User.Password(); ok {
			buddyAuthToken = password
		} else {
			buddyAuthToken = u.User.Username()
		}
		u.User = nil
		buddyUrl = strings.TrimSpace(u.String())
	}

	databaseUri := getEnv("STREMTHRU_DATABASE_URI", "sqlite://./data/stremthru.db")

	return Config{
		Port:              getEnv("STREMTHRU_PORT", "8080"),
		ProxyAuthPassword: proxyAuthPasswordMap,
		StoreAuthToken:    storeAuthTokenMap,
		BuddyURL:          buddyUrl,
		BuddyAuthToken:    buddyAuthToken,
		RedisURI:          getEnv("STREMTHRU_REDIS_URI", ""),
		DatabaseURI:       databaseUri,
	}
}()

var Port = config.Port
var ProxyAuthPassword = config.ProxyAuthPassword
var StoreAuthToken = config.StoreAuthToken
var BuddyURL = config.BuddyURL
var BuddyAuthToken = config.BuddyAuthToken
var RedisURI = config.RedisURI
var DatabaseURI = config.DatabaseURI
