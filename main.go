package main

import (
	"compress/gzip"
	// "context"
	"encoding/json"
	"fmt"
	"io"

	// "io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
		"github.com/boj/redistore"
	"github.com/gorilla/sessions"
	// "github.com/redis/go-redis/v9"
)

func main() {
	// ctx := context.Background()

	// // ✅ Redis connection (optional here, just for health check)
	// redisClient := redis.NewClient(&redis.Options{
	// 	Addr:     "yc0oskkcogs8ggoow488s4g0:6379",
	// 	Password: "123454321",
	// 	DB:       0,
	// 	Protocol: 2,  // Connection protocol
	// })
	// if err := redisClient.Ping(ctx).Err(); err != nil {
	// 	log.Fatalf("❌ Failed to connect to Redis: %v", err)
	// }
	// fmt.Println("✅ Connected to Redis")

	// ✅ Redis session store (no filesystem store at all)
	store, err := redistore.NewRediStore(
		10,                                     // pool size
		"tcp",                                  // network
		"yc0oskkcogs8ggoow488s4g0:6379",        // Redis address
		"123454321",                            // Redis password
		"543212345",                            // 🔧 key (as string)
	)
	if err != nil { 
		log.Fatalf("❌ Failed to create Redis session store: %v", err)
	}

		fmt.Println("✅ Connected to Redis db")

	defer store.Close()

	// Session cookie options
	store.Options = &sessions.Options{
		Path:     "/",
		Domain:   ".cloudasd.com", // share across subdomains
		MaxAge:   86400 * 30,       // 30 days
		HttpOnly: true,
		Secure:   false,            // true if HTTPS
	}

	// ctx := context.Background()

	// // Redis connection
	// redisClient := redis.NewClient(&redis.Options{
	// 	Addr:     "yc0oskkcogs8ggoow488s4g0:6379",
	// 	Password: "123454321",
	// 	DB:       0,
	// })
	// if err := redisClient.Ping(ctx).Err(); err != nil {
	// 	log.Fatalf("Failed to connect to Redis: %v", err)
	// }
	// fmt.Println("✅ Connected to Redis")

	// // File‐system backed sessions (cookie + FS), keyed by "_user_session"
	// store := sessions.NewFilesystemStore("", []byte("543212345"))
	// store.Options = &sessions.Options{
	// 	Path:     "/",
	// 	Domain:   ".cloudasd.com",
	// 	MaxAge:   86400 * 30, // 30 days
	// 	HttpOnly: true,
	// 	Secure:   false,
	// }

	// --- session endpoints ---------------------------------------

	// /test — bootstrap a session
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "_user_session")
		session.Values["plan_ids"] = []string{"semrush1"}
		session.Values["username"] = "test"
		session.Values["user_agent"] = r.UserAgent()
		session.Values["ip"] = r.RemoteAddr
		// store seconds since epoch
		session.Values["LAST_ACTIVITY"] = time.Now().Unix()
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	// /connect — JSON ⇒ session
	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}
		session, _ := store.Get(r, "_user_session")

		var body struct {
			PlanID    string `json:"plan_id"`
			Username  string `json:"username"`
			IP        string `json:"ip"`
			UserAgent string `json:"user_agent"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		session.Values["plan_ids"] = []string{body.PlanID}
		session.Values["username"] = body.Username
		session.Values["user_agent"] = body.UserAgent
		session.Values["ip"] = body.IP
		session.Values["LAST_ACTIVITY"] = time.Now().Unix()
		session.Save(r, w)

		w.Write([]byte(session.ID))
		log.Printf("Session set: %+v", session.Values)
	})

	// /access — aliasing by token
	http.HandleFunc("/access", func(w http.ResponseWriter, r *http.Request) {
		if token := r.URL.Query().Get("token"); token == "" {
			http.Error(w, "Token is required", http.StatusBadRequest)
			return
		}
		// you could do: session.ID = token  and Save, but here we just redirect
		http.Redirect(w, r, "/", http.StatusFound)
	})

	// /getdata — check session validity
	http.HandleFunc("/getdata", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "_user_session")
		if username, ok := session.Values["username"].(string); ok && username != "" {
			json.NewEncoder(w).Encode(map[string]string{"message": "Session is valid"})
			return
		}
		http.Error(w, "Invalid session", http.StatusForbidden)
	})

	// --- proxy with session-guard on “/” ---------------------------------------

	// list of paths to serve index.html instead
	blockedUrls := []string{
		"/user/logout",
		"/v4/authLogout",
		"/account/my-account",
		"/account",
	}

	proxyUrl := "http://aqbjjasr:v7zhmnx0ea22@216.173.120.34:6326"
	pURL, _ := url.Parse(proxyUrl)
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(pURL)}}

	http.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 1) session-guard
		session, _ := store.Get(r, "_user_session")
		username, _ := session.Values["username"].(string)
		planIDs, _ := session.Values["plan_ids"].([]string)
		lastActivity, _ := session.Values["LAST_ACTIVITY"].(int64)

		// ensure they have “semrush1”
		found := false
		for _, id := range planIDs {
			if id == "semrush1" {
				found = true
			}
		}
		// if missing or never set
		if username == "" || !found || lastActivity == 0 {
			w.Write([]byte("Session expired! Please try again"))
			return
		}
		// expire after 40 minutes
		if time.Now().Unix()-lastActivity > 40*60 {
			delete(session.Values, "username")
			delete(session.Values, "plan_ids")
			delete(session.Values, "LAST_ACTIVITY")
			session.Save(r, w)
			http.Error(w, "Session expired, please log in again", http.StatusUnauthorized)
			return
		}

		// 2) inject headers
		r.Header.Set("cookie", "__Secure-next-auth.session-token=eyJhbGciOiJkaXIiLCJlbmMiOiJBMjU2R0NNIn0..OlPK5fOBf7t8bFgS.E4qHcLYWFON6YUksfpAsZguijF9r5qXiWPCNYZcRkActYvuBPU3FoEVasVyFXz5IM0sQJwZrmj8307rT7tgQc8WN6XerFsvIuoN-TooPYlNYwwWTiCW4catn2flw_QP4qFrNJMWIRKs2JSMmbA3jzc83E0zOMyZ5kGQdshCi2Qq1NoRHK18aGe1Wm28BotxLr4rzBAOt7BHO91vBC9fV7RRGg_6yjwVuIzPh8T0zKstsRyHp1LW8rXGvq04NPI1EtKuvJfptTE3ztM_vxM7YwVWNfdYNFuBHWX-6k9GTG4jkdSaYeJgRqgj_9HwMdo04tMsJF8Rvfcx9obJptrNAehBM8G1yriac_X8HAihF_ydB01loEwNgXkdinZTfQuLvfMHGlkBhHODsmOlY6kJx1hD_QSyOSeoDIzBSl8aCbsms6sec1ngH15Lt9HocOA1pJQ4V0u9iUnonTD08vhD_sstAFIUdV1_jUZrPOBeymH7PVmnfbP6gDHTWBsHm2bnbYXfju0IKYTmMOnx2e5dTo24b4J7LQQIl3qIEZ3HXVy4DB7aSAEH4nnU4OAjbUCMv8Fh-8IHCt9544wfOLIBj87fHddcj-ZkupBCbMs4WWy8X0lNUXnh9iLHApj97KRCw6CtTMYXR4ekJJ4yIUPsGHYt44j-9_c6pnyk4BMc6SVL0KpcA98EmCgybEcGHfLj-IVDkRnG7bxzb3f84wJcUC2ZffC6h1ncc8Tj5m-knlBvE9h6WNrR4F-vIr4ayUGply9M5suLv2xTGWO-PE63PGvf0s1wN1EhTDdENB6vX5AHkt2_6qrEy0D-MlztqcT1x-M1iwoJQ90pd0UZ9nDyqIu_t9DjPQx-0y_IL0JsWFGojrWCivu6gVZilCuVDKENI204Fk4LPaLA3AHQ5kSp-Cz_TNcRKPpr1qw_uICmQV13sGZTapjDTJ4dQz3DrZ9oW03wDLxJKh1EfK4_-5ZbcWaYMiBrJ_xG6LcTva3pnQJo2eYUEjjpxbBcsO1EKZBM8yEpHyv6IXjBDKFXV6z64VNzghRjdFrwHOixv78YS15epVxHPdMWzLuwaPOde-vTEmFi2QlT83Q8UkWYmmXIkraqgVapAdK43BsD2GL4HOrC8zS8kTjZDoufV53b6nxlOb-XgsG9-yH72R8MVbXW0D2dq4CV-K2n1sbBoiWzxokc6ch2laNCUZAdcAIk0anx6eCs0ztrPW-l4CG6XM_HApWwkeSptpb6-ATwPWmMl9HbF3EqUoomblN_jOCd_PIearpWb3wn_ZGOEXqCV_rN8PuGUHeXDGC3ei4OZ-GMtTiALh8184Y3ypH2G-ewpL-XDbxR0KnjzXa-0grXyfV1rmNQlpV7MXjQ57h5zg0kpdGkDltuUhDAsqbmbt2Swp1SEnX8suIM3svJBbSlydKTVMYQoDDxuflq8zxpJ6q0wm1WI8s9mZSWMzHDAxRq8lTS7L2gLNUlHjtzI9aqq3VHqUBgtqYc9TEXy8zHoei7uJssQOysCro_IDrm5EL1CMbPtsTnY3qLU1Ax1OJZfpe8W1VUhmEFjXqd4DihgGForObzr0m-0etEnfvltsuZqPGPnSun4DNdYJwxSa47Atqmgfv3mIs0u7Lu9x7i-OE4xgTgtrVijdbtXpa63LZQjxHsvbRMAnchj00xHeNpohIfWthZnCFCYnPmhg0NFJDkK9hPwDtJ8Xr1IhtfNU2Bab3-GO6hxl4PXu38KH6U4znwTYEmwt6g7x4bikezhEawnwpWUsYPgaSF2iwPrCJzUHT1GT1uT_p3whBKl1W7uV8sUmeH98tWMt-AsGGrH_19sKEIeifF0kwtMhTBvrm45NPVK491wAtpwVI9_3dmRgG_DaTKgSuuKupzev8PEjLUaFD9GuDsQBBWuTuU2v3J6UApi8M16hL3dk1FCyFNu0h27jJke3ZEy-BMe58r4jnASMvysDjZjlurUXcaLXzEMtHNx90Z3DGIEOONOPoQIrRvSilxplVa2Aa0GQslGmooTWtdTatFyg-e93xU46UHpJMVwBHJtVw88Aowf6kEfZHgEnltiu5p_IkcttmSOm_prgj6UaBZliOnVzRgnIKsJggXSBhJ4e0VaTfKdS40J_Tqh_s4TkdDzQdOlTjdY7WdAncqX05MHB5A-pvhJ57kuRUCzS-ifQzy3XI9ybsQLDqDjQpu3k6rEiFWMkD7a97e9949K0ZkJ6Svzv40gHQj3XpawjfNOeEfyYpIiv-lU8C0m4MJVyOynnfJ1oJ6bpJVL8xSvG4bKWXfk1V29lQ3qmu2pZ3MmolR_SV_gJ5awiQhoJAoeVBDjGMG78BOoM35ZKVaqgIpOHyRCig5q4KtOJhrowyobnrMXdHsj9Spa6CyjCkgWLCm-WqlVIkqHkM57oLKVcoZU8pehQkrOUkJy8xu0HxXFQ_IAxpkg4946HeYGjoeO6VLMpriWhs1fZiLrEUCvVjg4NTPgmP3RC9sPC_erEbQH-DUf4chk9L49Pl6C2DVy_j4H5-4FBeH_Z7Gmp-nCO4pZ9cmo8jZX8lseHY8DznYalTa0eLB1cvFILMAwxUJFk9jnImuzcb35OYso2UHMUB5SMhwMoDG1-dlIXjASM8_F8UMGv4VACkme8qzRy7CQM5G7GgRAI6l39uYfTn7lrlVbwsiFG4s6Y42G3rI7yN3zIk2VVVPVeFfMHiPOsrIiLpz9BZBhiAfD6-55lmZTuYkSDlMnVJ3Cy_I7Slrg8vIKq-4lTbZpCDjjgcziAqurlmSpL5_ANKQgtem05XR3YcI5yDZhlsv_RBk8QlD37lU3wTy1pxU5eouYXp9hvCff8UzUK9LHjnAOikflMiiCUJuWCvzLa2F5rXpPIV-1zZiZ8sZIaYqT_cmeWn6aIKqi_XdBiJgn60Buvg4OtMnD-j9aECN2g8Ff_b8VqcZE_JK4iLscMHYj2ddBb1XBrTJmE_qz8DRabrQ3GceIiH4gKcCUZsnx4aNhI4AmIlr7_e0.Sq0Ziw-T3E64LS-88c3icg;__cflb=0H28vzvP5FJafnkHxigzNzx1VHUBBoXPtWY2514qws1;oai-hlib=true;_dd_s=rum=0&expire=1748841119637&logs=1&id=6d734e70-e282-4663-9173-f0e11d2bb8c5&created=1748839621281;__cf_bm=XfQubiCWntntWCqV6CdBT.kn21nTe1YneuPgGzkyYjQ-1748840050-1.0.1.1-iyc2UtVx0i57YF9qC5TQSIqLeS07eZVHX7xwhV7T8TGMBKI2pMVtlgKJwvpMO70KxwptrJU.4nFb3T3GQ3r1BGQKvuurEMkCvyjMao6EU2E;_cfuvid=6vS0N8wxFmpDxsErdLd5rJdm9HODDu5w_FJkU9M78VM-1748840137816-0.0.1.1-604800000;oai-sc=0gAAAAABoPS8c4E7qagkZRziobnZSvH9RDiTuoGqAwneroF3qumVYLTZGs_gMvjCeQ3Il8sx7MNvnsMyPI1CpO1s-SyV_HhbHK7DtJRhpLfhjeY-yRn9mo_G9nJJpgyRRNeEoeiBU7XgQfydHdZDgeSrETaQ7i3cxPENQyHCpoDFiDAglvJdCIhvA7kKmJf1L3LFQK7cYlzuyDWwZqPlRSlM-mWbkrubMw4cQwnUDbMZDsPu_JBxLSTE;oai-model-sticky-for-new-chats=false;__Host-next-auth.csrf-token=5fb984f64ef8411757dd94012cacab0a95a7e0e72101ccf189e5958d5e07bd52%7Ceceaed4bb86064ca5eb0cd756caf361af53a2261a203671453ce5feede7e2a26;__Secure-next-auth.callback-url=https%3A%2F%2Fchatgpt.com%2F;oai-did=6f59be13-6db0-4a17-b992-c94527368089")
		r.Header.Set("origin", "https://chatgpt.com")
		r.Header.Set("referer", "https://chatgpt.com/")
		r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)…")

		// 3) serve index for those blocked paths
		for _, blocked := range blockedUrls {
			if strings.HasPrefix(r.URL.Path, blocked) {
				http.ServeFile(w, r, "index.html")
				return
			}
		}

		// 4) proxy logic (determine targetBase, decompress, rewrite…)
		var targetBase, prefix string
		switch {
		case strings.HasPrefix(r.URL.Path, "/assetsx/"):
			targetBase, prefix = "https://cdn.oaistatic.com", "/assetsx"
		case strings.HasPrefix(r.URL.Path, "/cdnx/"):
			targetBase, prefix = "https://cdn.semrush.com", "/cdnx"
		case strings.HasPrefix(r.URL.Path, "/abx/"):
			targetBase, prefix = "https://ab.chatgpt.com", "/abx"
		default:
			targetBase = "https://chatgpt.com"
		}

		path := strings.TrimPrefix(r.URL.Path, prefix)
		fullUrl := targetBase + path
		fmt.Printf("Proxying: %s\n", fullUrl)

		proxyReq, _ := http.NewRequest(r.Method, fullUrl, r.Body)
		proxyReq.Header = r.Header
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, "Error proxying", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// copy headers…
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		

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

		//header delete or set
		w.Header().Del("Content-Encoding")
		w.Header().Del("Content-Length")
		w.Header().Del("content-security-policy")
		w.Header().Set("Transfer-Encoding", "chunked") // optional, Go handles this automatically usually
		w.Header().Set("X-Accel-Buffering", "no")      // for NGINX: disable buffering
		w.Header().Set("access-control-allow-origin", "*")
		w.Header().Set("access-control-allow-credentials", "true")
		w.WriteHeader(resp.StatusCode)


		// bodyBytes, err := ioutil.ReadAll(reader)
		// if err != nil {
		// 	http.Error(w, "Error reading body", http.StatusInternalServerError)
		// 	return
		// }

		// bodyStr := string(bodyBytes)
		

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				chunk := buf[:n]
				chunkStr := string(chunk)

				// Optional: rewrite domains in the stream too
				chunkStr = strings.ReplaceAll(chunkStr, "cdn.oaistatic.com", r.Host+"/assetsx")
				chunkStr = strings.ReplaceAll(chunkStr, "ab.chatgpt.com", r.Host+"/abx")
				chunkStr = strings.ReplaceAll(chunkStr, "chatgpt.com", r.Host)

				_, writeErr := w.Write([]byte(chunkStr))
				if writeErr != nil {
					break // client probably disconnected
				}
				flusher.Flush() // 🔁 force the chunk to the client
			}

			if err != nil {
				if err == io.EOF {
					break
				}
				http.Error(w, "Error streaming response", http.StatusInternalServerError)
				return
			}
		}

		// rewrite URLs
		if strings.Contains(contentType, "text/") || strings.Contains(contentType, "javascript") || strings.Contains(contentType, "json") || strings.Contains(contentType, "event-stream") {
			// bodyStr = strings.ReplaceAll(bodyStr, "cdn.oaistatic.com", r.Host+"/assetsx")
			// bodyStr = strings.ReplaceAll(bodyStr, "ab.chatgpt.com", r.Host+"/abx")
			// bodyStr = strings.ReplaceAll(bodyStr, "chatgpt.com", r.Host)
		}

		// w.Header().Set("Content-Type", contentType)
		// w.Header().Set("access-control-allow-origin", "*")
		// w.Header().Set("access-control-allow-credentials", "true")
		// // w.WriteHeader(resp.StatusCode)
		// w.Write([]byte(bodyStr))
	})

	fmt.Println("Server running on port 3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
