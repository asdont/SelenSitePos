package main

import (
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"strings"
	"time"
)

func main() {
	pos, url := getPositions("site.ru", "keyword", "ru-ru")
	fmt.Println(pos, url)
}

func getPositions(domain string, keyword string, country string) (int, string) {
	const (
		seleniumPathChrome = `chromedriver` // ChromeDriver 91.0.4472.19
		portChrome         = 9515
	)

	// Set the option of the selium service to null. Set as needed.
	var ops []selenium.ServiceOption

	service, errNewDriver := selenium.NewChromeDriverService(seleniumPathChrome, portChrome, ops...)
	if errNewDriver != nil {
		fmt.Printf("Error starting the ChromeDriver server: %v", errNewDriver)
	}

	// Delay service shutdown
	defer func(service *selenium.Service) {
		err := service.Stop()
		if err != nil {

		}
	}(service)

	caps := selenium.Capabilities{
		"browserName": "chrome",
		"platform":    "Windows 10",
	}

	chromeCaps := chrome.Capabilities{
		Path: "",
		Args: []string{
			//"--headless",
			"--disable-extensions",
			"--disable-plugins",
			"--disable-notifications",
			"--media-cache-size=1",
			"--mute-audio",
			// TODO менять на новых прокси
			"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/604.4.7 (KHTML, like Gecko) Version/11.0.2 Safari/604.4.7",
			//"--user-agent=Mozilla/5.0 (Windows NT 6.3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36",

		},
	}

	caps.AddChrome(chromeCaps)

	wd, errNewRemote := selenium.NewRemote(caps, "http://127.0.0.1:9515/wd/hub")
	if errNewRemote != nil {
		panic(errNewRemote)
	}

	defer func(wd selenium.WebDriver) {
		if err := wd.Quit(); err != nil {
			fmt.Println(err)
		}
	}(wd)

	time.Sleep(10 * time.Second)
	errWait := wd.SetImplicitWaitTimeout(3 * time.Second) // максимальное ожидание появления элемента
	if errWait != nil {
		fmt.Println(errWait)
	}

	errTimeout := wd.SetPageLoadTimeout(1 * time.Second) // тайм-аут для загрузки страницы
	if errTimeout != nil {
		fmt.Println(errTimeout)
	}

	// запрос страницы с требуемым keyword
	targetKeyword := strings.Join(strings.Split(keyword, " "), "+")
	targetUrl := fmt.Sprintf("https://duckduckgo.com/?q=%s&kl=%s", targetKeyword, country)
	errGetUrl := wd.Get(targetUrl)
	if errGetUrl != nil {
		panic(errGetUrl)
	}

	for i := 0; i < 2; i++ {
		// поиск целевых элементов(url-ы)
		elements, errFindSelector := wd.FindElements(selenium.ByCSSSelector, "h2 .result__a")
		if errFindSelector != nil {
			fmt.Println(errFindSelector)
		}

		// поиск требуемого домена
		position, url := searchDomain(domain, elements)
		if position != -1 {
			return position, url
		}

		// если домен не найден
		// прокрутка страницы вниз
		_, errScript := wd.ExecuteScript("window.scrollTo(0, document.body.scrollHeight);", nil)
		if errScript != nil {
			fmt.Println(errScript)
		}

		// ожидание от 2 до 5 секунд
		randomNumber := time.Now().Nanosecond() % 10
		if randomNumber < 2 || randomNumber > 5 {
			randomNumber = 3
			time.Sleep(time.Duration(randomNumber) * time.Second)
		}

		// клик по кнопке "больше результатов, если результат не подгрузились скриптом"
		next, _ := wd.FindElement(selenium.ByCSSSelector, ".result--more")
		errClick := next.Click()
		if errClick != nil {
			fmt.Println(errClick)
		}
	}

	time.Sleep(60 * time.Second)

	return -1, ""
}

func searchDomain(domain string, elements []selenium.WebElement) (int, string) {
	// поиск домена в урл-ах элементов
	for position, tags := range elements {
		url, errGetHref := tags.GetAttribute("href")

		if errGetHref != nil {
			fmt.Printf("position %s - error -> %s", position, errGetHref)
		} else {
			if strings.Contains(domain, url) {
				return position, url
			}
		}

		fmt.Printf("%d - %s\n", position, url)
	}

	fmt.Println("\nобработано", len(elements), "позиций")

	return -1, ""
}
