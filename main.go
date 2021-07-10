package main

import (
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"log"
	"net/url"
	"strings"
	"time"
)

const (
	seleniumPathChrome = `chromedriver` // ChromeDriver 91.0.4472.19
	portChrome         = 9515
)

func main() {
	domain := "https://www.domain.com" // какой домен искать
	keyword := "keyword"         // ключевое слово
	lang := "ru-ru"                          // язык
	search := "https://search engine.com/"   // поисковик

	// конфигурация драйвера
	wd, service, err := confDriver(keyword, lang, search)
	if err != nil {
		log.Fatalf("conf driver: %v", err)
	}
	defer func(service *selenium.Service) {
		err := service.Stop()
		if err != nil {
			log.Printf("selen service stop: %v", err)
		}
	}(service)

	// получение позиции
	pos, link, text, err := getPositions(*wd, domain, 3)
	if err != nil {
		log.Fatalf("get position and url: %v", err)
	}

	// результат
	if pos >= 0 {
		fmt.Printf("Позиция - %d\nlink - %s\ntext - %s", pos+1, link, text)
	} else {
		fmt.Println("Ничего не найдено")
	}
}

func getPositions(wd selenium.WebDriver, domain string, depth int) (int, string, string, error) {
	defer func(wd selenium.WebDriver) {
		if err := wd.Quit(); err != nil {
			log.Println(err)
		}
	}(wd)

	// depth глубина сбора(страницы)
	for i := 0; i < depth; i++ {
		// поиск целевых элементов(url-ы)
		elements, err := wd.FindElements(selenium.ByCSSSelector, "h2 .result__a")
		if err != nil {
			return 0, "", "", err
		}

		// поиск требуемого домена
		pos, link, text, err := searchDomain(domain, elements)
		if err != nil {
			return 0, "", text, err
		}

		if pos != -1 {
			return pos, link, text, nil
		}

		// если домен не найден
		// прокрутка страницы вниз
		_, err = wd.ExecuteScript("window.scrollTo(0, document.body.scrollHeight);", nil)
		if err != nil {
			return 0, "", "", err
		}

		time.Sleep(1 * time.Second)

		// клик по кнопке "больше результатов", если результаты не подгрузились скриптом
		next, _ := wd.FindElement(selenium.ByCSSSelector, ".result--more")
		err = next.Click()
		if err != nil {
			return 0, "", "", err
		}
	}
	return -1, "", "", nil
}

func searchDomain(domain string, elements []selenium.WebElement) (int, string, string, error) {
	// поиск домена в урл-ах элементов
	for pos, elem := range elements {
		link, err := elem.GetAttribute("href")
		if err != nil {
			return -1, "", "", fmt.Errorf("get attribute href: %v", err)
		}

		if strings.Contains(link, domain) {
			textURL, _ := elem.Text()
			return pos, link, textURL, nil
		}
	}
	return -1, "", "", nil
}

func confDriver(keyword string, lang string, search string) (*selenium.WebDriver, *selenium.Service, error) {
	var ops []selenium.ServiceOption

	service, err := selenium.NewChromeDriverService(seleniumPathChrome, portChrome, ops...)
	if err != nil {
		return nil, nil, err
	}

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
			"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/604.4.7 (KHTML, like Gecko) Version/11.0.2 Safari/604.4.7",
		},
	}

	caps.AddChrome(chromeCaps)

	wd, err := selenium.NewRemote(caps, "http://127.0.0.1:9515/wd/hub")
	if err != nil {
		return nil, nil, err
	}

	// максимальное ожидание появления элемента
	err = wd.SetImplicitWaitTimeout(3 * time.Second)
	if err != nil {
		return nil, nil, err
	}

	// тайм-аут для загрузки страницы
	err = wd.SetPageLoadTimeout(10 * time.Second)
	if err != nil {
		return nil, nil, err
	}

	siteURL, err := url.Parse(search)
	if err != nil {
		return nil, nil, err
	}

	siteParams := url.Values{
		"q":  {keyword},
		"kl": {lang},
	}

	siteURL.RawQuery = siteParams.Encode()

	errGetUrl := wd.Get(siteURL.String())
	if errGetUrl != nil {
		panic(errGetUrl)
	}

	return &wd, service, nil
}
