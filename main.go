package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	tatsu_api "github.com/tatsuworks/tatsu-api-go"
)

var Store = make(map[snowflake.ID]time.Time)
var TatsuApi tatsu_api.Client
var Discord bot.Client
var AppConfig *Config

func main() {
	{
		config, err := ReadConfig()

		if err != nil {
			panic(err)
		}

		AppConfig = config
	}

	client, err := disgo.New(AppConfig.DiscordToken,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentGuildVoiceStates,
			),
		),
		bot.WithEventListenerFunc(HandleJoin),
		bot.WithEventListenerFunc(HandleLeave),
	)

	if err != nil {
		panic(err)
	}

	Discord = client

	if err = Discord.OpenGateway(context.TODO()); err != nil {
		panic(err)
	}

	slog.Info("Bot running")

	TatsuApi = *tatsu_api.New(AppConfig.TatsuAPI)

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}

func CalculateScore(minutes int) int {
	result := 0

	scoreRange := AppConfig.TatsuMaxScore - AppConfig.TatsuMinScore

	for range minutes {
		result += rand.IntN(scoreRange+1) + AppConfig.TatsuMinScore
	}

	return result
}

func HandleJoin(e *events.GuildVoiceJoin) {
	uid := e.Member.User.ID

	BotLog(e.Member, "Użytkownik dołączył na kanał głosowy")

	if _, has := Store[uid]; !has {
		Store[uid] = time.Now()
	}
}

func HandleLeave(e *events.GuildVoiceLeave) {
	uid := e.Member.User.ID

	joinTime, entryExists := Store[uid]

	delete(Store, uid)

	if !entryExists {
		BotLog(e.Member, "Użytkownik opuścił kanał ale nie było go w bazie")
		return
	}

	diff := time.Now().Sub(joinTime)

	minutesSpent := int(diff.Abs().Minutes())

	if minutesSpent < 1 {
		BotLog(e.Member, "Mniej niż minuta na VC, nie ma punktów!")
		return
	}

	score := CalculateScore(minutesSpent)

	_, err := TatsuApi.ModifyGuildMemberScore(AppConfig.DiscordGid, uid.String(), tatsu_api.ActionAdd, uint32(score))

	if err != nil {
		panic(err)
	}

	BotLog(e.Member, fmt.Sprintf("Przyznano %d punktów za siedzenie na VC", score))
}

func BotLog(member discord.Member, msg string) {
	Discord.Rest().CreateMessage(
		snowflake.MustParse(AppConfig.DiscordLogCid),
		discord.NewMessageCreateBuilder().AddEmbeds(
			discord.NewEmbedBuilder().SetAuthor(member.EffectiveName(), "", member.EffectiveAvatarURL()).SetDescription(msg).Build(),
		).Build(),
	)

	slog.Info(msg, "UserID", member.User.ID.String())
}
