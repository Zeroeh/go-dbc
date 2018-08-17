//Not intended to be ran or compiled. This is just a showcase of the go-dbc module.

package main

import (
	"fmt"
	"time"
	"os"
	"bufio"
	"strings"
	"github.com/Zeroeh/go-dbc"
)

var (
	listURL []string
)

func main() {
	fmt.Println("Started")

	c := godbc.CaptchaClient{
		Username: "username",
		Password: "password",
		SiteKey: "6LfYpC0UAABAABI7pEgdrC8R0tY7goxU_wwSi8Ia",
		PollRate: 10,
	}
	listURL = readItems()
	if len(listURL) == 0 {
		fmt.Println("List was empty")
		return
	}
	fmt.Println("Read accounts file...")
	for i := 0; i < len(listURL); i++ {
		c.SiteURL = listURL[i]
		id, err := c.Decode(180)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Captcha ID:", id)
		pollEvent := time.Tick(time.Duration(c.PollRate) * time.Second)
		finishEvent := time.After(time.Duration(c.Timeout) * time.Second)
		for {
			select {
			case <-pollEvent:
				fmt.Println("Polling...")
				err = c.PollCaptcha(id)
				if err != nil {
					fmt.Println(err)
				}
			case <-finishEvent:
				fmt.Println("Did not solve the captcha in time.")
				c.LastStatus.Text = "nil"
			}
			if c.LastStatus.Text != "" {
				doSomething(c.LastStatus.Text)
				break
			}
		}
		time.Sleep(2500 * time.Millisecond)
	}

	fmt.Println("Finished")
}

func doSomething(t string) {
	fmt.Println(t)
}

func readItems() []string {
	f, err := os.Open("verify.urls")
	if err != nil {
		fmt.Printf("error opening file: %s\n", err)
	}
	defer f.Close()
	var accounts []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		accounts = append(accounts, strings.Split(scanner.Text(), "\n")[0])
	}
	return accounts
}
