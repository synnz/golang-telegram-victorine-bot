// Deprecated project, no goway structure

package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"log"
)

const (
	TIMEOUT       = 60
	DATABASE      = "sqlite3"
	DATABASE_NAME = "db.sqlite3"
)

var (
	bot *tgbotapi.BotAPI
	gdb *gorm.DB
)

// You must create bot_token.go file, which include TOKEN variable in global package scope
func init() {
	var err error

	gdb, err = gorm.Open(DATABASE, DATABASE_NAME)
	if err != nil {
		panic(err)
	}
	gdb.LogMode(true)

	gdb.AutoMigrate(
		&User{},
		&Group{},
		&Available{},
	)

	go startSchedule()

	bot, err = tgbotapi.NewBotAPI(TOKEN)
	if err != nil {
		log.Fatalf("cannot create bot: %s", err)
	}
}

func main() {
	defer gdb.Close()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = TIMEOUT

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		err.Error()
	}

	for update := range updates {
		update := update
		go func() {
			msg := update.Message
			if msg == nil {
				return
			}

			switch {
			case msg.IsCommand():
				handleCommand(msg, update)
			case msg.NewChatMembers != nil && len(*msg.NewChatMembers) > 0:
				handleNewMembers(msg, update)
			case msg.LeftChatMember != nil && !msg.LeftChatMember.IsBot:
				handleLeftMembers(msg, update)
			}
		}()
	}
}

func handleLeftMembers(msg *tgbotapi.Message, update tgbotapi.Update) {
	if msg != nil {
		newMsg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("нас покинул бротек @%s", update.Message.LeftChatMember.UserName))
		newMsg.ReplyToMessageID = msg.MessageID
		bot.Send(newMsg)
	}
}

func handleNewMembers(msg *tgbotapi.Message, update tgbotapi.Update) {
	newUsers := make([]string, 0, len(*msg.NewChatMembers))
	for _, user := range *update.Message.NewChatMembers {
		newUsers = append(newUsers, user.UserName)
	}

	var joinedUsers string
	for i, v := range newUsers {
		joinedUsers += v
		if i != len(newUsers)-1 {
			joinedUsers += ", "
			continue
		}

		joinedUsers += " "
	}

	template := "@%s, %s, в этом чате очко всегда сжато"
	var form string
	if msg != nil {
		if len(joinedUsers) > 1 {
			form = "поверьте"
		} else {
			form = "поверь"
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(template, joinedUsers, form))
		bot.Send(msg)
	}
}

func handleCommand(msg *tgbotapi.Message, update tgbotapi.Update) {
	command := msg.Command()
	createGroupRecord(msg)
	createAvailableRecord(msg)
	checkIfUsernameChanged(msg)
	checkIfPresenceUserNick(msg)
	switch command {
	case "regpi":
		regpi(msg, update)
	case "showpid":
		showpid(msg)
	case "pidor":
		startQuiz(msg)
	case "pidorstat":
		pidorStat(msg)
	case "unreg":
		if checkAdminAccess(msg, update) {
			unreg(msg, update)
		}
	case "kek":
		kekogen(msg)
	case "silent":
		disableNotify(msg, update)
	}
}
