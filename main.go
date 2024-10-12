package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"
)

func main() {
	settings := tele.Settings{
		Token:  os.Getenv("TELEGRAM_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(settings)
	if err != nil {
		log.Fatal(err)
		return
	}

	i, _ := strconv.Atoi(os.Getenv("ADMIN_GROUP_ID"))
	adminGroupID := int64(i)

	b.Handle(tele.OnNewGroupTitle, func(c tele.Context) error {
		err = c.Delete()
		if err != nil {
			return err
		}

		user := c.Message().Sender
		who := getUserDescriptionMD(user)
		chat := getGroupLinkMD(c.Message().Chat)
		text := fmt.Sprintf("%s changed group name to %s", who, chat)

		_, err = b.Send(&tele.Chat{ID: adminGroupID}, text, tele.ModeMarkdownV2)

		return err
	})

	b.Handle(tele.OnNewGroupPhoto, func(c tele.Context) error {
		err = c.Delete()
		if err != nil {
			return err
		}

		user := c.Message().Sender
		who := getUserDescriptionMD(user)
		chat := getGroupLinkMD(c.Message().Chat)
		text := fmt.Sprintf("%s changed chat photo in %s", who, chat)

		_, err = b.Send(&tele.Chat{ID: adminGroupID}, text, tele.ModeMarkdownV2)

		return err
	})

	b.Handle(tele.OnGroupPhotoDeleted, func(c tele.Context) error {
		err = c.Delete()
		if err != nil {
			return err
		}

		user := c.Message().Sender
		who := getUserDescriptionMD(user)
		chat := getGroupLinkMD(c.Message().Chat)
		text := fmt.Sprintf("%s deleted chat photo in %s", who, chat)

		_, err = b.Send(&tele.Chat{ID: adminGroupID}, text, tele.ModeMarkdownV2)

		return err
	})

	b.Handle(tele.OnUserJoined, func(c tele.Context) error {
		err = c.Delete()
		if err != nil {
			return err
		}

		msg := c.Message()

		if msg.Sender == nil || msg.UsersJoined == nil {
			return nil
		}

		sender := msg.Sender
		chat := getGroupLinkMD(c.Message().Chat)

		log.Printf("Sender: %+v", *sender)
		log.Printf("UserJoined: %+v", *msg.UserJoined)
		log.Printf("Chat MD: %s", chat)

		if len(msg.UsersJoined) == 1 && sender.ID == msg.UsersJoined[0].ID {
			whom := getUserLinkMD(sender)
			log.Printf("Sender MD: %s", whom)
			text := fmt.Sprintf("%s joined %s", whom, chat)

			_, err = b.Send(&tele.Chat{ID: adminGroupID}, text, tele.ModeMarkdownV2)

			return err
		}

		who := getUserDescriptionMD(sender)
		log.Printf("Sender MD: %+v", who)
		whom := make([]string, 0, len(msg.UsersJoined))

		for _, user := range msg.UsersJoined {
			log.Printf("UserJoined MD: %s", getUserLinkMD(&user))
			whom = append(whom, getUserLinkMD(&user))
		}

		text := fmt.Sprintf("%s added %s to %s", who, strings.Join(whom, ", "), chat)

		_, err = b.Send(&tele.Chat{ID: adminGroupID}, text, tele.ModeMarkdownV2)

		return err
	})

	b.Handle(tele.OnUserLeft, func(c tele.Context) error {
		err = c.Delete()
		if err != nil {
			return err
		}

		msg := c.Message()

		if msg.UserLeft == nil {
			return nil
		}

		sender := msg.Sender
		leftUser := msg.UserLeft
		who := getUserDescriptionMD(sender)
		whom := getUserLinkMD(leftUser)
		chat := getGroupLinkMD(c.Message().Chat)

		log.Printf("Sender: %+v", *sender)
		log.Printf("Sender MD: %s", who)
		log.Printf("UserLeft: %+v", *leftUser)
		log.Printf("UserLeftMD: %s", whom)

		if sender.ID == leftUser.ID {
			text := fmt.Sprintf("%s left %s", whom, chat)

			_, err = b.Send(&tele.Chat{ID: adminGroupID}, text, tele.ModeMarkdownV2)

			return err
		}

		text := fmt.Sprintf("%s removed %s from %s", who, whom, chat)

		_, err = b.Send(&tele.Chat{ID: adminGroupID}, text, tele.ModeMarkdownV2)

		return err
	})

	b.Handle(tele.OnPinned, func(c tele.Context) error {
		err = c.Delete()
		if err != nil {
			return err
		}

		user := c.Message().Sender
		who := getUserDescriptionMD(user)
		chat := getGroupLinkMD(c.Message().Chat)
		text := fmt.Sprintf("%s pinned %s in %s", who, getPinnedMessageLinkMD(c.Message()), chat)

		_, err = b.Send(&tele.Chat{ID: adminGroupID}, text, tele.ModeMarkdownV2)

		return err
	})

	b.Handle("/start", func(c tele.Context) error {
		sender := c.Sender()

		err = c.Send(fmt.Sprintf(`I don't understand you, %s\. But I logged it\.`, getUserLinkMD(sender)), tele.ModeMarkdownV2)

		text := fmt.Sprintf("%s tried to /start", getUserLinkMD(sender))

		_, err = b.Send(&tele.Chat{ID: adminGroupID}, text, tele.ModeMarkdownV2)

		return err
	})

	b.Start()
}

func getUserTitle(user *tele.User) string {
	return strings.Trim(fmt.Sprintf("%s %s", user.FirstName, user.LastName), " ")
}

func getUserDescriptionMD(user *tele.User) string {
	text := getUserTitle(user)
	return escapeSpecialCharactersMD(fmt.Sprintf(`%s (%s)`, text, user.Username))
}

func getUserLinkMD(user *tele.User) string {
	text := getUserTitle(user)
	return fmt.Sprintf("[%s](tg://user?id=%d)", escapeSpecialCharactersMD(text), user.ID)
}

func getGroupLinkMD(c *tele.Chat) string {
	text := escapeSpecialCharactersMD(c.Title)
	if c.Username == "" {
		return text
	}

	return fmt.Sprintf("[%s](tg://resolve?domain=%s)", text, c.Username)
}

func getPinnedMessageLinkMD(m *tele.Message) string {
	return fmt.Sprintf("[message](tg://resolve?domain=%s&post=%d&single)", m.ReplyTo.Chat.Username, m.ReplyTo.ID)
}

// In all other places characters:
// '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'
// must be escaped with the preceding character '\'.
var specialCharactersMD = "_*[]()~`>#+-=|{}.!"

func escapeSpecialCharactersMD(text string) string {
	prevIdx := 0

	for {
		idx := strings.IndexAny(text[prevIdx:], specialCharactersMD)
		if idx == -1 {
			break
		}

		idx += prevIdx
		text = text[:idx] + `\` + text[idx:idx] + text[idx:]
		prevIdx = idx + 2
	}

	return text
}
