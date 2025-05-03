package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
	// Set the proxy URL with authentication (username: aqbjjasr, password: v7zhmnx0ea22)
	proxyURL, err := url.Parse("http://aqbjjasr:v7zhmnx0ea22@216.173.120.34:6326")
	if err != nil {
		fmt.Println("Error parsing proxy URL:", err)
		return
	}

	// Create a custom HTTP transport that uses the proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	// Create an HTTP client with the custom transport
	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", "https://app.ahrefs.com/user/login", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set a User-Agent header to simulate a browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36")
	// req.Header.Set("Cookie", "kd-1258d9ef2538f660=Mav2NaYncxNKL8A9h0gcfpOmgKW6Q%2F6YRlU1DJr7oE3rib4Jt9WQ4AXoLACSy322n4p5RRFuA1ENhoBLfVA; intercom-device-id-dic5omcp=a386e995-9d7d-402c-9b88-ada987400c4e; __stripe_mid=c6275cfa-17b3-41e8-9101-5b436f39054a07f24a; __cflb=02DiuFMJyRDQ1SqAwiXo5YsPbMTGqHELuJvax1nVmqLBN; __cf_bm=K0NndsSYs.IeimEZ3RD42BNyo5VRDkiYUkUL1aW7VZE-1745406715-1.0.1.1-uBr6QmtS6eNy7KZW1p96PCU_eyEFr8HTuFhUPQh1scYBTxd3aX.HTdkeg5Jysu.jNOyH8meOn2qRrUlzx_2xDmqUxorkf_8lwPshNbFMqRg; cf_clearance=ZRTbm6WrPnyONked25fyM8ePVZF5MtjG16CTwh4Yd0k-1745406718-1.2.1.1-LKhzSWtkt2rKlmQT9mFLRpYy.ZqXtzwUkytJaFNcRQKz90sW_SA858WfLRtn0ipfrhbPI6J5T5IvXGn8owjuiOIe6CmEokroYnt8L8JJKVrexwPjBwy47NDoEr7HB9L2C9Bjimn8Ecel0mfddled5WYrVcOUYEuKD6TiHNfVrg0Zv2lm7X3g6zxMMn3Wl0_QkVFffoiU4p5ckZZjAZ8.b.I2KHzMJAurcUHrxpYfV1Yp3Ahf8Mm4mD09e1oQLGv110OWORLdOaKNYtB79LSuZ8WVsyfubmkoAJaNRuS6NMmMq7bYJ0r4cyageEMO_mG1PVwjARPpa8lY7G7O2aVTa9KiPkD3LSX.aOxkQsnAsqU")
	req.Header.Set("origin", "https://app.ahrefs.com")
	req.Header.Set("referer", "https://app.ahrefs.com/")
	req.Header.Set("sec-ch-ua-platform-version", "\"15.0.0\"")
	req.Header.Set("sec-ch-ua-full-version-list", "\"Not)A;Brand\";v=\"99.0.0.0\", \"Google Chrome\";v=\"127.0.6533.122\", \"Chromium\";v=\"127.0.6533.122\"")
	req.Header.Set("sec-ch-ua-bitness", "\"64\"")
	req.Header.Set("sec-ch-ua-model", "\"\"")
	req.Header.Set("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Set("sec-ch-ua-arch", "\"x86\"")
	req.Header.Set("sec-ch-ua-full-version", "\"127.0.6533.122\"")
	req.Header.Set("sec-ch-ua", "\"Not)A;Brand\";v=\"99\", \"Google Chrome\";v=\"127\", \"Chromium\";v=\"127\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")

	// Make the HTTPS request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	// Read the body content
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading body:", err)
		return
	}

	// Print the HTML content
	fmt.Println("HTML Response:")
	fmt.Println(string(body)) // Print the body as a string
}
