package main

import (
	"aienvoy/internal/pkg/config"
	"aienvoy/pkg/cookiecloud"
	"fmt"
)

func main() {
	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)
	get := func(domain string) {
		cookies, _ := cc.GetCookies(domain)

		fmt.Printf("\n\nDomain: %s\n", domain)
		for _, cookie := range cookies {
			fmt.Printf("%s: %s\n", cookie.Name, cookie.Value)
		}
	}

	get("claude.ai")
	get(".claude.ai")
	get(".google.com")
	get("sspai.com")
	get(".sspai.com")
}
