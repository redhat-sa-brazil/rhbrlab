package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/fatih/set.v0"      // Package is no longer maintained, need refactoring
	"gopkg.in/tucnak/telebot.v2" // Package is actively maintained
)

/*
Required Variables. App will crash without them... no checks... =)
*/
var (
	telegramToken        = os.Getenv("TELEGRAM_TOKEN")
	telegramChatID       = os.Getenv("TELEGRAM_CHATID")
	towerURL             = os.Getenv("TOWER_URL")
	towerUser            = os.Getenv("TOWER_USER")
	towerPass            = os.Getenv("TOWER_PASS")
	towerStartTemplateID = os.Getenv("TOWER_START_TEMPLATE_ID")
	towerStopTemplateID  = os.Getenv("TOWER_STOP_TEMPLATE_ID")
	labUsers             = set.New()
)

/*
Function main. Currently megazord pattern (need to be refactored)
*/
func main() {

	log.Println("[INFO] Initializing and registering to Telegram API")
	log.Printf("[INFO] This bot will only accept command from chatID %s\n", telegramChatID)

	b, err := telebot.NewBot(telebot.Settings{
		Token:  telegramToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatalf("[ERROR] Problem connecting to Telegram API with error %s\n", err)
	}

	b.Handle("/start", func(m *telebot.Message) {
		log.Printf("[INFO] /start command from user %s at chat %s\n", m.Sender.Username, m.Chat.Title)
		if enforceChatID(b, m) {
			return
		}
		if labUsers.Size() > 0 {
			msg := fmt.Sprintf("O lab já está ligado! Check-ins registrados: %s", labUsers.List())
			b.Send(m.Chat, msg)
			log.Printf("[INFO] /start command aborted\n")
			return
		}
		token, err := towerAuthenticate(towerURL, towerUser, towerPass)
		if err {
			b.Send(m.Chat, "Ops! Tivemos problemas com o seu comando... =(")
			log.Println("[ERROR] Problem authenticating with Ansible Tower API")
			return
		}
		towerJobLaunch(token, towerStartTemplateID)
		if err {
			b.Send(m.Chat, "Ops! Tivemos problemas com o seu comando... =(")
			log.Println("[ERROR] Problem launching a job with Ansible Tower API")
			return
		}
		labUsers.Add(m.Sender.Username)
		b.Send(m.Chat, "Acionando o Ansible Tower para ligar o lab! Já fiz seu /checkin.")
	})

	b.Handle("/stop", func(m *telebot.Message) {
		log.Printf("[INFO] /stop command from user %s at chat %s\n", m.Sender.Username, m.Chat.Title)
		if enforceChatID(b, m) {
			return
		}
		if labUsers.Size() > 0 {
			msg := fmt.Sprintf("Não posso desligar o lab! Check-ins registrados: %s", labUsers.List())
			b.Send(m.Chat, msg)
			log.Printf("[INFO] /stop command aborted\n")
			return
		}
		token, err := towerAuthenticate(towerURL, towerUser, towerPass)
		if err {
			b.Send(m.Chat, "Ops! Tivemos problemas com o seu comando... =(")
			log.Println("[ERROR] Problems authenticating with Ansible Tower API")
			return
		}
		towerJobLaunch(token, towerStopTemplateID)
		if err {
			b.Send(m.Chat, "Ops! Tivemos problemas com o seu comando... =(")
			log.Println("[ERROR] Problem launching a job with Ansible Tower API")
			return
		}
		b.Send(m.Chat, "Acionando o Ansible Tower para desligar o lab!")
	})

	b.Handle("/checkin", func(m *telebot.Message) {
		log.Printf("[INFO] /checkin command from user %s at chat %s\n", m.Sender.Username, m.Chat.Title)
		if enforceChatID(b, m) {
			return
		}
		b.Send(m.Chat, "Check-in registrado!")
		labUsers.Add(m.Sender.Username)
	})

	b.Handle("/checkout", func(m *telebot.Message) {
		log.Printf("[INFO] /checkout command from user %s at chat %s\n", m.Sender.Username, m.Chat.Title)
		if enforceChatID(b, m) {
			return
		}
		b.Send(m.Chat, "Check-out registrado!")
		labUsers.Remove(m.Sender.Username)
	})

	b.Handle("/status", func(m *telebot.Message) {
		log.Printf("[INFO] /status command from user %s at chat %s\n", m.Sender.Username, m.Chat.Title)
		if enforceChatID(b, m) {
			return
		}
		log.Printf("[INFO] %d users: %s\n", labUsers.Size(), labUsers.List())
		if labUsers.Size() > 0 {
			msg := fmt.Sprintf("Temos %d check-ins registrados: %s", labUsers.Size(), labUsers.List())
			b.Send(m.Chat, msg)
		} else {
			b.Send(m.Chat, "Não temos nenhum check-in registrado!")
		}
	})

	b.Handle("/clear", func(m *telebot.Message) {
		log.Printf("[INFO] /clear command from user %s at chat %s\n", m.Sender.Username, m.Chat.Title)
		if enforceChatID(b, m) {
			return
		}
		b.Send(m.Chat, "Forçando limpeza da lista de check-ins!")
		labUsers.Clear()
	})

	log.Println("[INFO] Starting event loop")

	b.Start()
}

/*
Function to make RESTful API call to Ansible Tower and authenticate by user/pass.
*/
func towerAuthenticate(url string, user string, pass string) (string, bool) {
	jsonData := map[string]string{"username": user, "password": pass}
	jsonValue, _ := json.Marshal(jsonData)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	response, err := client.Post(url+"/authtoken/", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Printf("[ERROR] Failed executing HTTP request with error %s\n", err)
		return "", true
	}

	data, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Printf("[ERROR] Failed reading the response buffer with error %s\n", err)
		return "", true
	}

	type TowerAuthToken struct {
		Token   string
		Expires string
	}
	towerToken := &TowerAuthToken{}
	err = json.Unmarshal(data, &towerToken)
	if err != nil {
		log.Printf("[ERROR] Failed parsing response with error %s\n", err)
		return "", true
	}

	log.Printf("[INFO] Ansible Tower API authentication succeeded!")

	return towerToken.Token, false
}

/*
Function to make RESTful API call to Ansible Tower and launch a Job Template.
*/
func towerJobLaunch(token string, jobID string) bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("POST", towerURL+"/job_templates/"+jobID+"/launch/", nil)
	req.Header.Add("Authorization", "Token "+token)
	_, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Failed executing HTTP request with error %s\n", err)
		return true
	}
	log.Printf("[INFO] Ansible Tower API request for Job Template ID launch %s succeeded!", jobID)
	return false
}

/*
Function to make sure the bot will only accept commands from specific chat.
*/
func enforceChatID(b *telebot.Bot, m *telebot.Message) bool {
	chatID, _ := strconv.ParseInt(telegramChatID, 10, 64)
	if m.Chat.ID != chatID {
		log.Printf("[ERROR] This chatID %d is not authorized! Leaving!\n", m.Chat.ID)
		b.Leave(m.Chat)
		return true
	}
	return false
}
