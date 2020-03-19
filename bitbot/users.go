package bitbot

import (
	"fmt"
	gorm "github.com/jinzhu/gorm"
	"github.com/whyrusleeping/hellabot"
	log "gopkg.in/inconshreveable/log15.v2"
	"strings"
	"time"
)

type IdleUserInfo struct {
	Username string
	Time     int64
}

var TrackIdleUsers = NamedTrigger{
	ID:   "trackIdleUsers",
	Help: "Passive, non-interactive, experimental trigger. Monitors time since last activity for all users in channel. Works like /whois.",
	Condition: func(irc *hbot.Bot, m *hbot.Message) bool {
		return m.Command == "PRIVMSG"
	},
	Action: func(irc *hbot.Bot, m *hbot.Message) bool {
		err := b.TrackIdleUsers(m)
		if err != nil {
			log.Error(err.Error())
		}
		return false // keep processing triggers
	},
}

func (b Bot) TrackIdleUsers(m *hbot.Message) error {
	var lastUserMessage = IdleUserInfo{Username: m.From, Time: time.Now().Unix()}
	err := b.DB.Create(&lastUserMessage).Error

	return err
}

var ReportIdleUsers = NamedTrigger{
	ID: "reportIdleUsers",
	Condition: func(irc *hbot.Bot, m *hbot.Message) bool {
		return m.Command == "PRIVMSG" && strings.HasPrefix(m.Content, "!idle")
	},
	Action: func(irc *hbot.Bot, m *hbot.Message) bool {
		args := strings.Split(m.Content, " ")
		if len(args) < 2 {
			irc.Reply(m, "Please specify a nick to lookup")
			return true
		}
		log.Info(fmt.Sprintf("Getting idle time for %s", args[1]))
		report, err := b.GetUserIdleTime(args[1])
		if err != nil {
			irc.Reply(m, "Unable to lookup idle time for that user")
			return true
		}
		if report == "" {
			irc.Reply(m, "I have not seen that user yet.")
			return true
		}
		irc.Reply(m, fmt.Sprintf("%s has been idle for %s", args[1], report))
		return true
	},
}

func (b Bot) GetUserIdleTime(nick string) (string, error) { // get the last time a person spoke
	var user IdleUserInfo
	res := b.DB.Where("Username = ?", nick).First(&user)

	err := res.Error
	if err != nil {
		if gorm.IsRecordNotFoundError(res.Error) {
			return "", nil
		} else {
			return "", err
		}
	}
	ts := time.Unix(user.Time, 0)
	elapsed := fmtDuration(time.Since(ts))

	return elapsed, err
}
