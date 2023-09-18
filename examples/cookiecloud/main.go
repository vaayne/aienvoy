package main

import (
	"fmt"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/cookiecloud"
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
