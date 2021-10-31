// built from someones orgin script [ it was like 40 lines long LOL ] so- special shoutout to nothing
package main

//sudo apt install statgrab

import (
	"bufio"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chyeh/pubip"
	. "github.com/logrusorgru/aurora"
	ps "github.com/mitchellh/go-ps"
	"golang.org/x/net/html"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/showwin/speedtest-go/speedtest"
)

var (
	URLFILLQ        string
	filepathsizedir string
	size            int64
	root            string
	err             error
	filepathsize    string
	filepathtime    string
	filepathtype    string
	scanfiletype    string
	filehex         string
	files           []string
	UL              string
	ErrNoPath       = errors.New("path required")
	process         ps.Process
	wg              sync.WaitGroup
	urlQueue        = make(chan string)
	config          = &tls.Config{InsecureSkipVerify: true}
	transport       = &http.Transport{
		TLSClientConfig: config,
	}
	hasCrawled = make(map[string]bool)
	netClient  *http.Client
)

/// type structure

type UrlTitle struct {
	idx   int
	url   string
	title string
}

type info struct {
	Hostname string `bson:hostname`
	Platform string `bson:platform`
	CPU      string `bson:cpu`
	RAM      uint64 `bson:ram`
	Disk     uint64 `bson:disk`
}

func process_listing() {
	processList, err := ps.Processes()
	if err != nil {
		log.Println("Gatheriong has Failed, Might you be using Windows?")
		return
	}
	// map ages
	for x := range processList {
		process = processList[x]
		log.Printf("%d\t%s\n", process.Pid(), process.Executable())

		// do os.* stuff on the pid
	}
}

func inf() {
	hostStat, _ := host.Info()
	cpuStat, _ := cpu.Info()
	vmStat, _ := mem.VirtualMemory()
	diskStat, _ := disk.Usage("/")
	info := new(info)
	info.Hostname = hostStat.Hostname
	info.Platform = hostStat.Platform
	info.CPU = cpuStat[0].ModelName
	info.RAM = vmStat.Total / 1024 / 1024
	info.Disk = diskStat.Total / 1024 / 1024
	fmt.Print("%+v\n[CPU]  >>> ", info.CPU)
	fmt.Print("%+v\n[RAM]  >>> ", info.RAM)
	fmt.Print("%+v\n[DSK]  >>> ", info.Disk)
	fmt.Print("%+v\n[PLAT] >>> ", info.Platform)
	fmt.Print("%+v\n[HOST] >>> ", info.Hostname)
}

func clsa() {
	if runtime.GOOS == "windows" {
		fmt.Println(Red("[-] I Will not be able to execute this"))
	} else {
		out, err := exec.Command("clear").Output()
		if err != nil {
			log.Fatal(err)
		}
		output := string(out[:])
		fmt.Println(output)
	}
	if runtime.GOOS == "windows" {
		os := "linux"
		fmt.Println("[-] Sorry, this command is system spacific to -> ", os, "Systems")
	} else {
		out, err := exec.Command("pwd").Output()
		if err != nil {
			log.Fatal(err)
		}
		output := string(out[:])
		fmt.Println("[~] Working Directory ~> ", output)
	}
}

func hostname() {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("\033[32mHost >>> [", hostname, "\033[32m]")
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println("\n\033[32mIP   >>> [", localAddr.IP, "\033[32m]")
	return localAddr.IP
}

/////////////////////////////////////////// TREE //////////////////////////////////////////////
func tree() {
	slab, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	files, err := ioutil.ReadDir(slab)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if !f.IsDir() {
			fmt.Println("\t\t\t\033[37m┡", f.Name())
		}
	}
}

//////////////////////////////////// GET A DIRECTORY SIZE ///////////////////////////////////////////////////////
func sizedir() {
	fmt.Println(Red("------------------------------------------------------------------------------"))
	fmt.Scanf("%s", &filepathsizedir)
	directory := filepathsizedir
	err := filepath.Walk(directory, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
			fmt.Println("\033[31m[!] An error has occured while finding size -> ", err)
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	if err != nil {
		log.Fatal(err)
		fmt.Println("\033[31m[!] An error has occured while finding Directory -> ", err)
	}
	fmt.Printf("\033[32mhas a size of -> %d\n", size)
}

///////////////////////////////// FILING FINDING/DUMPING/CONTENT/TIME/ETC ///////////////////////////////////////

// hex dumping

func list() {
	ex := "ls"
	cmd := exec.Command(ex)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(Cyan(string(stdout)))
}

func hexdumper() {

	list()
	fmt.Println(Red("------------------------------------------------------------------------------"))
	fmt.Println("File to Dump | ")
	fmt.Scanf("%s", &filehex)
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("[+] Opening File....", filehex)
	time.Sleep(1000 * time.Millisecond)
	f, err := os.Open(filehex)
	if err != nil {
		log.Fatal(err)
		fmt.Println("Error Occured when loading -> ", filehex)
	}
	reader := bufio.NewReader(f)
	buf := make([]byte, 256)
	for {
		_, err := reader.Read(buf)

		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}
		fmt.Println(Red("[+] HEX FORMAT | "))
		fmt.Printf("%s", hex.Dump(buf))
	}
}

///////////////////////////////// FILING -> Finding Files based on strings //////////////////////////

func testerrso() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("WORKING FROM DIR -> ", pwd)
}

func filebystring() {
	fmt.Println(Red("------------------------------------------------------------------------------"))
	fmt.Println("File Extension : EX | txt, rb, go, rs, asm, nasm, em, r, hack, php")
	fmt.Scanf("%s", &scanfiletype)
	fmt.Println(Red("------------------------------------------------------------------------------"))
	fmt.Println(Red("File path to search from"))
	fmt.Scanf("%s", &filepathtype)
	root := filepathtype
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == scanfiletype {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fmt.Println("FILE FOUND -> ", file)
	}
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("")
	fmt.Println("**************************************************************")
	fmt.Println("*WORKING FROM DIR -> ", pwd, "   *")
	fmt.Println("**************************************************************")
}

///////////////////////////////// FILING -> Finding Files based on SIZE //////////////////////////

func VisitFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return nil
	}
	size := info.Size()
	if !info.IsDir() && size > 1024*1024 {
		files = append(files, path)
	}
	return nil
}

func findall() {
	fmt.Println(Red("------------------------------------------------------------------------------"))
	fmt.Println(Red("File path to search from"))
	fmt.Scanf("%s", &filepathsize)
	root := filepathsize
	err := filepath.Walk(root, VisitFile)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fmt.Println("\033[34m[*] File Found ~> ", file)
	}
}

///////////////////////////////// FILING -> Finding Files based on TIME //////////////////////////

func findbytime() {
	fmt.Println(Red("------------------------------------------------------------------------------"))
	fmt.Println(Red("File path to search from"))
	fmt.Scanf("%s", &filepathtime)
	root := filepathtime
	t, err := time.Parse("2006-01-02T15:04:05-07:00", "2021-05-01T00:00:00+00:00")
	if err != nil {
		log.Fatal(err)
	}
	files, err := FindFilesAfter(root, t)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fmt.Println("\033[34m[+] FILE FOUND -> ", file)
	}
}

func FindFilesAfter(dir string, t time.Time) (files []string, err error) {
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".txt" && info.ModTime().After(t) {
			files = append(files, path)
		}
		return nil
	})
	return
}

/////////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////// PARSING HTML AND URL PAGES ///////////////////////////////////////////////////////////

func grab() {
	fmt.Println(Red("------------------------------------------------------------------------------"))
	fmt.Println("\033[31mEnter a stakc overflow url with a tagged topic like")
	fmt.Println("\033[31mhttps://stackoverflow.com/questions/tagged")
	fmt.Scanf("%s", &URLFILLQ)
	webPage := URLFILLQ
	resp, err := http.Get(webPage)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("Hm? Seems as if it has timed out or didnt fufil the request %d %s", resp.StatusCode, resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".question-summary .summary").Each(func(i int, s *goquery.Selection) {
		title := s.Find("h3").Text()
		fmt.Println(i+1, title)
	})
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////
func sys() {
	if runtime.GOOS == "linux" {
		fmt.Println(Cyan("\033[32m[*] Detected System -> Linux"))
	}
	if runtime.GOOS == "windows" {
		fmt.Println(Cyan("\033[32m[*] Detected System -> Windows"))
	}
}

func clcrawl() {
	if runtime.GOOS == "linux" {
		ex := "clear"
		cmd := exec.Command(ex)
		stdout, err := cmd.Output()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(Cyan(string(stdout)))
	}
	if runtime.GOOS == "windows" {
		cl := "cls"
		cd := exec.Command(cl)
		stdout, err := cd.Output()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(Cyan(string(stdout)))
	}
}

func isValidUri(uri string) bool {
	_, err := url.ParseRequestURI(uri)

	return err == nil
}

func toUrlList(input string) []string {
	list := strings.Split(strings.TrimSpace(input), "\n")
	urls := make([]string, 0)

	for _, url := range list {
		if isValidUri(url) {
			urls = append(urls, url)
			file, fileErr := os.Create("urls.txt")
			if fileErr != nil {
				fmt.Println("[!] Could not Create a File.......")
				fmt.Println(fileErr)
			}
			fmt.Fprintf(file, "%v\n", url)
		}
	}

	return urls
}

func fetchUrlTitles(urls []string) []*UrlTitle {
	ch := make(chan *UrlTitle, len(urls))
	for idx, url := range urls {
		go func(idx int, url string) {
			doc, err := goquery.NewDocument(url)

			if err != nil {
				ch <- &UrlTitle{idx, url, ""}
			} else {
				ch <- &UrlTitle{idx, url, doc.Find("title").Text()}
			}
		}(idx, url)
	}
	urlsWithTitles := make([]*UrlTitle, len(urls))
	for range urls {
		urlWithTitle := <-ch
		urlsWithTitles[urlWithTitle.idx] = urlWithTitle
	}
	return urlsWithTitles
}

func toMarkdownList(urlsWithTitles []*UrlTitle) string {
	markdown := ""
	for _, urlWithTitle := range urlsWithTitles {
		markdown += fmt.Sprintf("- [%s](%s)\n", urlWithTitle.title, urlWithTitle.url)
	}
	return strings.TrimSpace(markdown)
}

/// get URL ID's

func getHtmlPage(webPage string) (string, error) {
	resp, err := http.Get(webPage)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func parse(text string) {
	tkn := html.NewTokenizer(strings.NewReader(text))
	var isTd bool
	var n int
	for {
		tt := tkn.Next()
		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			t := tkn.Token()
			isTd = t.Data == "td"
		case tt == html.TextToken:
			t := tkn.Token()
			if isTd {

				fmt.Printf("%s ", t.Data)
				n++
			}
			if isTd && n%3 == 0 {
				fmt.Println()
			}
			isTd = false
		}
	}
}

//////////////////////////////////////// complex url shifting //////////////////

func processElement(index int, element *goquery.Selection) {
	href, exists := element.Attr("href")
	if exists {
		fmt.Println(href)
	}
}

func grabparse() {
	hardurl := "placeholder" // figure out parsing with the command line arguments
	uro := hardurl
	parsedURL, err := url.Parse(uro)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("-------------------------- URL PARSED -------------- ")
	fmt.Println("Scheme        --->  " + parsedURL.Scheme)
	fmt.Println("Hostname      --->  " + parsedURL.Host)
	fmt.Println("Path in URL   --->  " + parsedURL.Path)
	fmt.Println("Query Strings --->  " + parsedURL.RawQuery)
	fmt.Println("Fragments     --->  " + parsedURL.Fragment)
}

/////////////////////////////////////////////////////////////////////////////////

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

// this one is specified handel for the live monitor
//go handel(make(chan os.Signal, 1))
func handel(c chan os.Signal) {
	signal.Notify(c, os.Interrupt)
	for s := <-c; ; s = <-c {
		switch s {
		case os.Interrupt:
			fmt.Println("\nDetected Interupt.....")
			os.Exit(1)
		case os.Kill:
			fmt.Println("\n\n\tKILL received")
			os.Exit(1)
		}
	}
}

//go sighandel(make(chan os.Signal, 1))
func sighandel(c chan os.Signal) {
	signal.Notify(c, os.Interrupt)
	for s := <-c; ; s = <-c {
		switch s {
		case os.Interrupt:
			fmt.Println("\nDetected Interupt.....")
			helpmen()
			fmt.Println(Red("Please use the exit command to exit!"))
		case os.Kill:
			fmt.Println("\n\n\tKILL received")
			os.Exit(1)
		}
	}
}

func desk() {
	url := "https://google.com" // desk net inf
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		fmt.Println("[-] Sorry will not be able to run this command")
	} else {
		if resp.StatusCode >= 200 {
			out, err := exec.Command("notify-send", "Server responded with code 200 Connection is stable  °˖✧◝(⁰▿⁰)◜✧˖° ✔️").Output()
			if err != nil {
				log.Fatal(err)
			}
			output := string(out[:])
			fmt.Println(output)
		} else {
			out, err := exec.Command("notify-send", "Server Responded with a code that is not within the indexed list or range").Output()
			if err != nil {
				log.Fatal(err)
			}
			output := string(out[:])
			fmt.Println(output)
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// working directory

func wherecurrent() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("\033[35m#################### Working from ####################")
	fmt.Println("\033[35m~~> ", pwd)
}

// time all

func timenow() {
	t := time.Now()
	fmt.Println("Current Year        |", t.Year())
	fmt.Println("Current Month       |", t.Month())
	fmt.Println("Current Day         | ", t.Day())
	fmt.Println("Current Hour        |", t.Hour())
	fmt.Println("Current Minute      |", t.Minute())
	fmt.Println("Current Second      |", t.Second())
	fmt.Println("Current Nanosecond  |", t.Nanosecond())
}

////////////////////////////////////////////////////////////////// Logging for desktop live connection update /////////////////////////////////////////
func test() bool {
	_, err := http.Get("https://www.google.com")
	if err == nil {
		log.Fatal(err)
		return true
	} else {
		fmt.Println(Red("Device may have been disconnected from the network"))
		return false
	}
}

func desknotif() {
	if runtime.GOOS == "windows" {
		fmt.Println("hmm you seem to be on a windows system, this will not work on windows based systems, please try again later")
	} else {
		out, err := exec.Command("notify-send", "Testing Server Conn and Node every 20-30 seconds").Output()
		if err != nil {
			log.Fatal(err)
		} else {
			output := string(out[:])
			fmt.Println(output)
		}
	}
}

func maintesterlive() {
	//test()
	url := "https://google.com"
	if err != nil {
		log.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		fmt.Println("[-] Sorry will not be able to run this command")
	}
	for {
		resp, err := http.Get(url)
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print("\n\n\tRESPONSE LENGTH -> ", len(dump))
		go handel(make(chan os.Signal, 1))
		if err != nil {
			log.Fatal(err)
			fmt.Println("\033[31m hmmm? a error has been raised trying to log")
		}
		fmt.Println("\n\n\tTesting Connection every 30 seconds")
		time.Sleep(30 * time.Second)
		if resp.StatusCode >= 200 {
			out, err := exec.Command("notify-send", "Server responded with code 200 Connection is stable  °˖✧◝(⁰▿⁰)◜✧˖° ✔️").Output()
			if err != nil {
				log.Fatal(err)
			}
			output := string(out[:])
			fmt.Println(output)
			go handel(make(chan os.Signal, 1))
		} else {
			out, err := exec.Command("notify-send", "Server Responded with a code that is not within the indexed list or range").Output()
			if err != nil {
				log.Fatal(err)
			}
			output := string(out[:])
			fmt.Println(output)
			go handel(make(chan os.Signal, 1))
		}
	}
}

///////////////////////////////////////////////// NETWORKING | IP FETCHING, PRIV, PUB, INTER, ETH, /////////////////////////////////////////

func grabpubip() {
	hostStat, _ := host.Info()
	info := new(info)
	info.Hostname = hostStat.Hostname
	uli := "https://api.ipify.org?format=text"
	fmt.Print("\033[32m\n\t\tFetching Public IPA for -> ", info.Hostname, " ...\n")
	response, err := http.Get(uli)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	ip, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\033[32m\t\tPublic Internet Address came back with ~>  %s\n", ip)
}

func GetPulicIP() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	return localAddr[0:idx]
}

func grabpubip1() {
	hostStat, _ := host.Info()
	info := new(info)
	info.Hostname = hostStat.Hostname
	ip, err := pubip.Get()
	if err != nil {
		log.Fatal(err)
		fmt.Println("[-] Could not or retrive the local Ip address for user -> ", info.Hostname)
	} else {
		fmt.Print("\n[Fetching IP for HOST] >>> ", info.Hostname)
		fmt.Println("[+] IPA Found for Hostname -> ", info.Hostname)
		fmt.Println(ip)
	}
}

////////////////////////////////////////Interfaces and local addresses///////////////////////////////////////////////////

func localaddr() {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print(fmt.Errorf("localAddresses: %v\n", err.Error()))
		log.Fatal(err)
		return
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Print(fmt.Errorf("localAddresses: %v\n", err.Error()))
			continue
		}
		for _, a := range addrs {
			log.Printf("%v %v\n", i.Name, a)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
// networking speeds

func testspeed() {
	user, _ := speedtest.FetchUserInfo()
	serverList, _ := speedtest.FetchServerList(user)
	targets, _ := serverList.FindServer([]int{})

	for _, s := range targets {
		s.PingTest()
		s.DownloadTest(false)
		s.UploadTest(false)
		fmt.Println("\t\033[35m ╭─────────╮")
		fmt.Println("\t\033[35m │Latency  │", s.Latency)
		fmt.Println("\t\033[35m │Download │", s.DLSpeed)
		fmt.Println("\t\033[35m │Upload   │", s.ULSpeed)
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
//help

func helpmen() {
	prg := "cat"
	prg1 := "txt/cmd.txt"
	command := exec.Command(prg, prg1)
	out, err := command.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Print(Magenta(string(out)))
}

//banner

func banlol() {
	prgo := "ruby"
	arg1 := "txt/banner.rb"
	cmd := exec.Command(prgo, arg1)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Print(Red(string(out)))
}

func netcheck() bool {
	_, err := http.Get("https://www.google.com")
	if err == nil {
		fmt.Println(Cyan("\033[32m[+] Connection Stable...."))
		return true
	}
	fmt.Println(Cyan("[-] Interface has been disconnected from the network, please connect or set a connection "))
	os.Exit(1)
	return false
}

func clear() {
	if runtime.GOOS == "linux" {
		ex := "clear"
		cmd := exec.Command(ex)
		stdout, err := cmd.Output()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(Cyan(string(stdout)))
	}
	if runtime.GOOS == "windows" {
		cl := "cls"
		cd := exec.Command(cl)
		stdout, err := cd.Output()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(Cyan(string(stdout)))
	}
}

func rb() {
	prgo := "ruby"
	arg1 := "main.rb"
	cmd := exec.Command(prgo, arg1)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Print(Red(string(out)))
}

func execution(input string) error {
	args := strings.Split(input, " ")
	switch args[0] {
	case "cd":

		if len(args) < 2 {
			return ErrNoPath
		}

		return os.Chdir(args[1])
	case "exit":
		os.Exit(0)
	case "con":
		netcheck()
	case "inf":
		inf()
	case "processing":
		process_listing()
	case "help":
		helpmen()
	case "networking":
		mainmain()
	case "version":
		fmt.Println("Version: 1.0 Beta")
	case "command":
		helpmen()
	case "FILEBYTYPE":
		filebystring()
	case "FILEHEX":
		hexdumper()
	case "FILEBYSIZE":
		findall()
	case "FILEBYTIME":
		findbytime()
	case "cls":
		clsa()
	case "trem":
		tree()
	case "DIRBYSIZE":
		sizedir()
	case "stackquestion":
		grab()
	case "dev":
		dev()
	case "livenetworking":
		mainlarf()
	case "inter":
		localaddr()
	case "workfrom":
		wherecurrent()
	case "findmyp":
		grabpubip()
	case "timedate":
		timenow()
	case "desknot":
		maintesterlive()
	case "intest":
		testspeed()
	}
	cmd := exec.Command(args[0], args[1:]...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

////////////////////////////////////////////////////////////////////////// GO SERV //////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func msf() {
	arg1 := "msfconsole-start"
	cmd := exec.Command(arg1)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Print(Red(string(out)))
}

func clresp() {
	if runtime.GOOS == "windows" {
		fmt.Println(Red("[-] I Will not be able to execute this"))
	} else {
		out, err := exec.Command("clear").Output()
		if err != nil {
			log.Fatal(err)
		}
		output := string(out[:])
		fmt.Println(output)
	}
	if runtime.GOOS == "windows" {
		os := "linux"
		fmt.Println("[-] Sorry, this command is system spacific to -> ", os, "Systems")
	} else {
		out, err := exec.Command("pwd").Output()
		if err != nil {
			log.Fatal(err)
		}
		output := string(out[:])
		fmt.Println("[~] Working Directory ~> ", output)
	}
}

func mndesknot() {
	if runtime.GOOS == "windows" {
		fmt.Println(Cyan("[-] Sorry, but t this time i can not run this command"))
	} else {
		out, err := exec.Command("notify-send", "Testing Server Conn and Node every 20-30 seconds").Output()
		if err != nil {
			log.Fatal(err)
		} else {
			output := string(out[:])
			fmt.Println(output)
		}
	}
}

func resplog() {
	url := "https://google.com"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		fmt.Println("[-] Sorry will not be able to run this command")
	} else {
		if resp.StatusCode >= 200 {
			out, err := exec.Command("notify-send", "Server responded with code 200 Connection is stable  °˖✧◝(⁰▿⁰)◜✧˖° ✔️").Output()
			if err != nil {
				log.Fatal(err)
			}
			output := string(out[:])
			fmt.Println(output)
		} else {
			out, err := exec.Command("notify-send", "Server Responded with a code that is not within the indexed list or range").Output()
			if err != nil {
				log.Fatal(err)
			}
			output := string(out[:])
			fmt.Println(output)
		}
	}
}

func logger() {
	if runtime.GOOS == "windows" {
		fmt.Println("This appends to a linux system only command, i will not be able to run it")
	} else {
		out, err := exec.Command("notify-send", "There was an error within the response").Output()
		if err != nil {
			log.Fatal(err)
		}
		output := string(out[:])
		fmt.Println(output)
	}
}

func get() {
	clresp()
	url := "https://google.com"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(Cyan("--------------------------Server Response---------------------------"))
	fmt.Println("[+] Response Status  -> ", resp.StatusCode, http.StatusText(resp.StatusCode))
	fmt.Println("[+] Date Of Request  -> ", resp.Header.Get("date"))
	fmt.Println("[+] Content-Encoding -> ", resp.Header.Get("content-encoding"))
	fmt.Println("[+] Content-Type     -> ", resp.Header.Get("content-type"))
	fmt.Println("[+] Connected-Server -> ", resp.Header.Get("server"))
	fmt.Println("[+] X-Frame-Options  -> ", resp.Header.Get("x-frame-options"))
	fmt.Println(Cyan("--------------------------Server X-Requests-----------------------------"))
	for k, v := range resp.Header {
		fmt.Print(Cyan("[+] -> " + k))
		fmt.Print(Red(" -> "))
		fmt.Println(v)
	}
}

func mainlarf() {
	netcheck()
	clear()
	time.Sleep(10 * time.Second)
	mndesknot()
	seconds := "20"
	time.Sleep(1 * time.Second)
	fmt.Println("[~] Testing Connection Every ", seconds, "Seconds")
	time.Sleep(1 * time.Second)
	for {
		time.Sleep(30 * time.Second)
		url := "https://google.com"
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode >= 200 {
			fmt.Println("[+] Response Status Given -> ", resp.StatusCode, http.StatusText(resp.StatusCode))
			fmt.Println("[+] Response seems good")
			resplog()
			get()
			go handel(make(chan os.Signal, 1))

		}
		if resp.StatusCode >= 300 && resp.StatusCode <= 400 {
			fmt.Println("[+] Response Status Given -> ", resp.StatusCode, http.StatusText(resp.StatusCode))
			fmt.Println("[~] Response may be laggy")
			logger()
			get()
			go handel(make(chan os.Signal, 1))
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////// development TESTING OMLY ///////////////////////////////////
//
//	scanner := bufio.NewScanner(os.Stdin)
//	fmt.Print("HTTPS URL >>>  ")
//	scanner.Scan()
//	text := scanner.Text()
//	tex := text
//	///
//	scan := bufio.NewScanner(os.Stdin)
//	fmt.Print("WWW URL 	 >>> ")
//	scan.Scan()
//	fed := scan.Text()
//	tex1 := fed
//	///
//	reader := bufio.NewReader(os.Stdin)
//	fmt.Print("Enter your name: ")
//	name, _ := reader.ReadString('\n')

func dev() {
	//
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("HTTPS URL >>>  ")
	scanner.Scan()
	text := scanner.Text()
	tex := text
	///
	scan := bufio.NewScanner(os.Stdin)
	fmt.Print("WWW URL 	 >>> ")
	scan.Scan()
	fed := scan.Text()
	tex1 := fed
	///
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	name, _ := reader.ReadString('\n')
	cmd := exec.Command("go", "run", "modules/GO-Liath/user.go", tex, tex1, name)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Print(Red(string(out)))
}

///////////////////////////////////////////// URL AND LIVE SYSTEM CONNECTION/SERVER

func processE(index int, element *goquery.Selection) {
	href, exists := element.Attr("href")
	if exists {
		fmt.Println(href)
	}
}

func parsee() {
	fmt.Scanf("%s", &UL)
	parsedURL, err := url.Parse(UL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(Cyan("-------------------------- URL PARSED -------------- "))
	fmt.Println(Cyan("Scheme        --->  " + parsedURL.Scheme))
	fmt.Println(Cyan("Hostname      --->  " + parsedURL.Host))
	fmt.Println(Cyan("Path in URL   --->  " + parsedURL.Path))
	fmt.Println(Cyan("Query Strings --->  " + parsedURL.RawQuery))
	fmt.Println(Cyan("Fragments     --->  " + parsedURL.Fragment))
	fmt.Println("-------------- URL QUERY VALS ----------------------- ")
	queryMap := parsedURL.Query()
	fmt.Println(queryMap)
}

func mainmain() {
	//parse()
	fmt.Println(Red("[>] Put a complex URL down below"))
	fmt.Scanf("%s", &UL)
	parsedURL, err := url.Parse(UL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("-------------------------- URL PARSED -------------- ")
	fmt.Println("Scheme        --->  " + parsedURL.Scheme)
	fmt.Println("Hostname      --->  " + parsedURL.Host)
	fmt.Println("Path in URL   --->  " + parsedURL.Path)
	fmt.Println("Query Strings --->  " + parsedURL.RawQuery)
	fmt.Println("Fragments     --->  " + parsedURL.Fragment)
	fmt.Println("-------------- URL QUERY VALS ----------------------- ")
	time.Sleep(2 * time.Second)
	queryMap := parsedURL.Query()
	fmt.Println(queryMap)
	response, err := http.Get(UL)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("a Error has occured while loading HTTP response body ", err)
	}
	fmt.Println("[+] Scraping URLS......")
	document.Find("a").Each(processElement)
	fmt.Println(Red("----------------------------- GATHERING CODE NOTES ----------------------"))
	time.Sleep(2 * time.Second)
	response, err = http.Get(UL)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Error reading HTTP body. ", err)
	}
	re := regexp.MustCompile("<!--(.|/n)*?-->")
	comments := re.FindAllString(string(body), -1)
	if comments == nil {
		fmt.Println("No matches.")
	} else {
		for _, comment := range comments {
			fmt.Println(comment)
		}
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	clear()
	banlol()
	GetOutboundIP()
	hostname()
	for {
		fmt.Print("\n\n\033[34m(҂‾ ▵‾)︻デ═一 [>>  ")
		go sighandel(make(chan os.Signal, 1))
		//read input
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		input = strings.TrimSuffix(input, "\n")
		if input == "" {
			continue
		}
		if err = execution(input); err != nil {
			//fmt.Fprintln(os.Stderr, err) // i have it as nothing right now because of command usage and linking
		}
	}
}
