package main

import (
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"time"
)

func main() {
	wd := initializationSelenium()
	getPositions(wd)
}

func initializationSelenium() selenium.WebDriver {
	const (
		seleniumPathChrome = `chromedriver` // ChromeDriver 91.0.4472.19
		portChrome         = 9515
	)

	// Set the option of the selium service to null. Set as needed.
	var ops []selenium.ServiceOption

	service, err := selenium.NewChromeDriverService(seleniumPathChrome, portChrome, ops...)
	if err != nil {
		fmt.Printf("Error starting the ChromeDriver server: %v", err)
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
			"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/604.4.7 (KHTML, like Gecko) Version/11.0.2 Safari/604.4.7",
		},
	}

	caps.AddChrome(chromeCaps)

	wd, err := selenium.NewRemote(caps, "http://127.0.0.1:9515/wd/hub")
	if err != nil {
		panic(err)
	}

	defer func(wd selenium.WebDriver) {
		if err := wd.Quit(); err != nil {
			fmt.Println(err)
		}
	}(wd)

	time.Sleep(10 * time.Second)
	errWait := wd.SetImplicitWaitTimeout(3 * time.Second) // максимальное ожидание появления элемента
	if errWait != nil {
		fmt.Println(err)
	}

	errTimeout := wd.SetPageLoadTimeout(10 * time.Second) // тайм-аут для загрузки страницы
	if errTimeout != nil {
		fmt.Println(err)
	}

	return wd
}

func getPositions(wd selenium.WebDriver) {
	if err := wd.Get("https://duckduckgo.com/?q=кинотеатры+Санкт&kl=ru-ru"); err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)

	for i := 0; i < 2; i++ {
		// прокрутка страницы вниз
		_, err := wd.ExecuteScript("window.scrollTo(0, document.body.scrollHeight);", nil)
		if err != nil {
			fmt.Println(err)
		}

		// ожидание от 3 до 5 секунд
		randomNumber := time.Now().Nanosecond() % 10
		if randomNumber < 3 || randomNumber > 5 {
			randomNumber = 3
			time.Sleep(time.Duration(randomNumber) * time.Second)
		}

		// клик по кнопке "больше результатов, если не подгрузилось"
		next, _ := wd.FindElement(selenium.ByCSSSelector, ".result--more")
		errClick := next.Click()
		if errClick != nil {
			fmt.Println(err)
		}
	}

	time.Sleep(1 * time.Second)

	result, err := wd.FindElements(selenium.ByCSSSelector, "h2 .result__a")
	if err != nil {
		fmt.Println(err)
	}

	for position, domainTag := range result {
		url, err := domainTag.GetAttribute("href")
		if err != nil {
			fmt.Printf("position %s - error -> %s", position, err)
		} else {
			fmt.Printf("%d - %s\n", position, url)
		}
	}

	fmt.Println("\nполучено", len(result), "позиций")

	time.Sleep(600 * time.Second)
}
