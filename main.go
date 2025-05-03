package main

import (
	// "compress/gzip"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/andybalholm/brotli" // Import Brotli library
)

func main() {
	// Define blocked URLs
	blockedUrls := []string{
		"/user/logout",
		"/v4/authLogout",
		"/account/my-account",
		"/account",
	}

	// Configure proxy
	proxyUrl := "http://aqbjjasr:v7zhmnx0ea22@216.173.120.34:6326"
	proxy, err := url.Parse(proxyUrl)
	if err != nil {
		log.Fatal("Error parsing proxy URL:", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}

	// Serve static content
	http.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Main proxy handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set required headers
		r.Header.Set("Cookie", "__Secure-next-auth.session-token=eyJhbGciOiJkaXIiLCJlbmMiOiJBMjU2R0NNIn0..FfVWvE0a4tQTlelV.Eqgm7QzPPuyru0suS85ehJtyJMsj8ON2QCrceTLjSb1PJacGs2NQPrnOTAcUBorbtEKHbSbwm9YAt2mYH5FNeJgp7pazqCHbwCY7PdWRjKOVNBbQt5CAYZ-QETFey3b2IW6s5Y3WRv6kOqyk4LtwaVVBx5RZqrBLzBNKxGf6Sl3HWaj9psAmUQa9trcOK9KKqbD6_te-AgkOula62D70Tn9hyTb9rj9PUBjUJLVD5DAwLf6ri034JxovzIT4rZEddIY_vCqi6iR8vsvvYW5rN3T4JNcwu2weEdkpkNmwVgEn6UzdQOok6bHLY8uVu81dmeyQbEFj969TV_-eNc7fgWJWj9VBRHnVC7adrzghc9aII9wdQhAkz9XR8BA7ChFTdSUY6-y8d-guhG12By_6JGbEjmEW9frVoqcjdnc1TirvtKcqd20W6rx-g1gUAwZySPSsmEwkBJW_Az_1K0dsKqrD7Z0hA_jPze_FsoiCzKQHG8NCre7dY8ku9smZeCMXHeG7bMKoZMEmpww7pX5mRoCMV6TFsI3loKHUXULr1xkIQMUpoiUvkdi6MiqkwVMa0ZKZKAVDRLvB49vXJd1E5PKdD2jU-BZH-nB2LCec4YV7H0IAuCVDaQgcULgwNcSCwcPYH7l_IvFNU1p4pyA4qk8WQ5MwAnXpHN1c1_553Hz2CvsgHz1yJ_EfZEBSOpkNa8gTuwZq0XFdnPALLnUId5fYx1pDNvKwLf3vfFCspqeNFWqo06jKApqKUAg5VzzHhKw_wKzA1o1CMCexRuBP_Q0GEypouMtr1IyJ15WHB6eWkQd9IEUbh4HQkcwsbXEVFC1qhGGMo5Wy7_4nWe2WTK-_ZFaWlmgLTjLhhbCsjmXD41-dxbsiuF3IuXUEhbARTFb3b9zjxKgrptmFQss65cNHTg-5cSDtSr69IlR5yTuY5mhe3wgA6qr9EbeVRvOSF_fivPBGR3AOAMPgo_4M_8Y3Q8nhgRGgKWufNy7wkKRT-5hRHcWNrUlMQEEPbH_SwbdCD2uV9XHh0drHDaZxIFeaJUQfNHnBl0qyGxntQnQWzkQElWNCB1jaueoTBo8DUur52PtOsD-kEng43d5SbOM2n1v7tqtZAQC3jF9YorHb8WPaDuDO4Y0gDo9jLt_9XXhcnT-bKcCZbMEuQsCf8LMubgF5M30ufDiXOpRDBVplu4oK4jHc-vuM1DlueYWHwwbXO7-daH1a7UzjBlBSNYi8CeKxiVaIzgtGyo_uQofqED5WR2ltd7XCvi7lUPuqeUN5dWOiYSQG8EJKm4sM8p5zFickX1_RC89pBG50B2hTODXtHkmarNJ8MzS7KIT4A03MmEYtGOOfST8NBxFoOHd8EfjvoEcNr8UD3LVbGFo76QEq9ushZ1-5pMA-BIFkSLroweRAeFPRsVTjt4eUhNAse8fX_QZ8ffzd6mKFE9Ti8VhWySQ_LOWE6hOuYD6h59T0Iw3hJP6s6STiIjPrCeFD3da7tjMcwqi7K0fYZSQyat3pYvIxEO5FiiQpC1wpGwxDx7YqunnaeYWx5FA67TBOH29X3dt_6jYewGt1ApdMHz4HF7UEn7TMYdSv4_QbQsH6nd2ByiQr0o1h_J5vT8GpY_x-xVmQTaqBcq87mV8fm31mLtsI_eAS-Y2MhA0o4VVUTyyecLhdQyp9LT8yrzoqx3m7QeC_MQdHS-X4vJoEVR324mSv_f9vp6ud8BiYSGP4YjSNVRAieskDWreLSuj4mU-UKLWJSJ7drpY_rSzj78YPcFEnGtaR_eBWKe7JcCcv12cZE6mAOhn0tIwijb9gso6xiTeFt1XUyiRn0HVg4aMltR0UY8vUKVFglkIemJo5MJmxESIJ1DQHAqazZ0JQoz2FTr3UskiwVzs3uHiT4_ErX2UWV3ey-6T_B_QTrr_IvAF13LGavmsl0D9HtsjlDc1Ts5uem4dLrZOawwNNRMHvV9abzZ6ovLj-XdBKNshPbF49qCQe0jQoX-toswPpMpfzjl1GbnixXvmyuosjjvpdin4AWlo9ZeHhU26E8CsWQw8LdYS6mMd_MJ-fpQH2NeF5COE8Ef3fVewjNMMl4y4gnFDn6VzesM51l9ymYWgyGfeKo2BB7yvJlWzSqDhiMZFnTPL5myJdC6x2s0el-anFFyyVj_sXoG-v4N1hbzawANIHPRXVbpHOWh_IloXasX3HTFqDUVQDjpeSmIviGyEL08LuO_vxDXJW5Gt2dlJLNoQLQ9Jr3aj8VUhqTcP1RVrn9a8V_pYXAG5E-qgAKLDKnDO1rpGBAnuivB1ZZLaWlVSPg6XEtPDgZkkuvYmxRh3rRY56ZWvHFNOilH5D8kOK1mJJLuplx3VnDcVPBW8cDehD2JFCrOe-WcFgmUNrDIayX811R6WwlEfkTHEVFEKZpbXo0bK47P1l0F3VmovoI2-r8e33Mpj4T1FhxTxzwkV4j4o_yHwguujxmGx_rH3Gd2SOYZMXFBxLY0vHzW4M30E-7zTgGYRPbedyvh1vcrb16l-Vp-8NWKsGW8Rwo84DN9iZbvJEShlqLbyV7RFkwoRGH6CA38NUeNJ7QEqDBgY4wVoQp89H3VIKLpADSjm1JmYJi8k3lODJ_eU7epQCpIwfDx3oHJCoP_HfQPHpm88XxhQ2XDiMSSf1OFwvY0zDM7XwIMNyMI0BsDM7OFyOlP6HfDGDIdtxYMv5h_8uzvyfzLQ1cNS6pUxO5Vyed-6lMVBTEpj_SYvX-qvc5AozofRsA8hQd2LamKdTLXlqpmjw6gN4YOPPDmzLsNkV-3VLH-Zz6oZOP_faQ-VuKaESg6uIgffZLBjuhnWadg0grn4tzPVKEcXfekZvkRcynsnYkdwDb7OOvoT9WNmBDUxnSCJFtLTaLQxuD_5v6lVs2kJdhVFuQBbY0biDdjeWX8hpSUYQrWTs8FAGA9YzmQUnYzy5rwBeL4hKloBld2QK63Wk_l49Uc-hv4XXlO8bhtfRYW68ijGbNbtrKYJqtT8.uzC2e3QGGpyaLI2y2Rzrvg;__cflb=04dTofELUVCxHqRn2XXJp6ZcQHxBRL5fpnzW6mqEvf;oai-hlib=true;_dd_s=rum=0&expire=1745391334394&logs=1&id=a73393ee-fcb6-491e-9de3-560ebf13cd2e&created=1745390254390;__cf_bm=9zyo5tGAlOnZF7JPWc.VEm3mKCtD1X8I6TQIMTkTj2U-1745390251-1.0.1.1-bVQg8B.1xZ1rGV8IMlo64BlpVs1cgzXSgtfK5BAueoRkvJo2iTTQBIJ98mphKkIJ3HT8GMaClSUNRjoIrdaPeLPlK0hNjhSNuIvToBl9qlw;oai-sc=0gAAAAABnkSABesZwDC0GkEzQD_uZ_aoyrw3YfxQUhz2MTYojGo6hMJywqWlUlUjVIIAZZsb7SwTledsfzeVMxCXAph3Y9hnaLUPcJnq0Do9ZAjvBi0KPU9MKGhIF2rZCC-8YjGLt1lZ65aVfiihtHJr5Mwg8J4MCgtdCS0hJoT1iLNJEv6WlaKvA1glNrR5G-3rW1XKt5EgAoTwXeOd2uKE1MZDeKzTqpZybAD9ILh8j9kcjcviR5KI;_cfuvid=WokvDxfH7fIfaYy_HpMHdEj4JaQkmxFZ8kOVQ4Fmskk-1745390251584-0.0.1.1-604800000;oai-sc=0gAAAAABoCIqvl53hJmSrMiTIEa2anSmhu5ChkRwJaDjyVSYERq2Tbg75g2I2xkcXVBd0DZ_k78Aa0qEgu4Cp9LD_F4JI7VICt5K5c3RHjjrSXGK1-EmQM_dHxi_xtZ8E_WbDpwUmkyAxikx6u1ksHBfnJFhvvABh5DjCQ_m8kTZ4vpAQINOgFV-lCTxp7ghfPJSpXPF9Hq7N8OrJKN0_AgOf32Q3Qpx2nl3zQcm0iRQYYxoWCiCW_Rs;__Host-next-auth.csrf-token=028ef98132b32c0731b04f556d0cccb1f51f00440f27c29b239e1dbbd68d97fd%7C1a68e90d1adeedeefc456eb3250e87783a7ba5690cb955ca027b13f5aad4319e;__Secure-next-auth.callback-url=https%3A%2F%2Fchatgpt.com;oai-did=b13184ed-7859-4caa-9fad-5a971857a53a;oai-nav-state=1")
		r.Header.Set("origin", "https://chatgpt.com")
		r.Header.Set("referer", "https://chatgpt.com/")
		r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")

		// Block specific paths
		for _, blocked := range blockedUrls {
			if strings.HasPrefix(r.URL.Path, blocked) {
				http.ServeFile(w, r, "index.html")
				return
			}
		}

		// Determine target URL and strip prefix
		var targetBase, prefix string
		if strings.HasPrefix(r.URL.Path, "/assetsx/") {
			targetBase = "https://cdn.oaistatic.com"
			prefix = "/assetsx"
		} else if strings.HasPrefix(r.URL.Path, "/cdnx/") {
			targetBase = "https://cdn.semrush.com"
			prefix = "/cdnx"
		} else if strings.HasPrefix(r.URL.Path, "/abx/") {
			targetBase = "https://ab.chatgpt.com"
			prefix = "/abx"
		} else {
			targetBase = "https://chatgpt.com"
		}

		// Reconstruct path without the prefix
		path := strings.TrimPrefix(r.URL.Path, prefix)

		fullUrl := targetBase + path
		fmt.Printf("Proxying: %s\n", fullUrl)

		// Create proxy request
		proxyReq, err := http.NewRequest(r.Method, fullUrl, r.Body)
		if err != nil {
			http.Error(w, "Error creating request", http.StatusInternalServerError)
			return
		}
		proxyReq.Header = r.Header

		// Send proxy request
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, "Error proxying", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// … after client.Do(proxyReq) and defer resp.Body.Close()

		// 1. Copy all upstream headers...
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		// 2. Remove these two so the browser doesn’t hang or re‐decode:
		//    - Original Content-Encoding (we’ve already decompressed)
		//    - Original Content-Length (no longer matches modified body)
		w.Header().Del("Content-Encoding")
		w.Header().Del("Content-Length")
		w.Header().Del("content-security-policy")

		// 3. Decompress if needed, read & replace
		contentType := resp.Header.Get("Content-Type")
		var reader io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(resp.Body)
			if err != nil {
				http.Error(w, "Error decompressing gzip", http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			reader = gz
		} else if resp.Header.Get("Content-Encoding") == "br" {
			reader = brotli.NewReader(resp.Body)
		}
		bodyBytes, err := ioutil.ReadAll(reader)
		if err != nil {
			http.Error(w, "Error reading body", http.StatusInternalServerError)
			return
		}

		// 4. Perform your URL replacements
		bodyStr := string(bodyBytes)
		if strings.Contains(contentType, "text/") || strings.Contains(contentType, "javascript") || strings.Contains(contentType, "json") || strings.Contains(contentType, "event-stream") {
			bodyStr = strings.ReplaceAll(bodyStr, "cdn.oaistatic.com", r.Host+"/assetsx")
			bodyStr = strings.ReplaceAll(bodyStr, "ab.chatgpt.com", r.Host+"/abx")
			bodyStr = strings.ReplaceAll(bodyStr, "chatgpt.com", r.Host)
		}

		// 5. Now write status + headers + body
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("access-control-allow-origin", "*")
		w.Header().Set("access-control-allow-credentials", "true")
		w.WriteHeader(resp.StatusCode)
		w.Write([]byte(bodyStr))
	})

	fmt.Println("Server running on port 3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
