package main

import (
	"log"
	"os"
	"strconv"

	metro_tenerife "github.com/Alexsilvacodes/metro-tenerife-bot/metro_tenerife"
	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func sendMessage(msg tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) {
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func buildMessageTram(panel_trams []metro_tenerife.PanelTram, update tgbotapi.Update, f func(m tgbotapi.MessageConfig)) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "")

	statusStr := "🚊 Próximos Tranvías en *" + panel_trams[0].StopDescription + "*\n\n"

	for _, panel_tram := range panel_trams {
		statusStr += "🚏 *Destino:* " + panel_tram.DestinationStopDescription + "\n" +
			"🕒 *Tiempo Restante:* " + strconv.Itoa(panel_tram.RemainingMinutes) + " minutos\n" +
			"👫 *Ocupación:* " + strconv.Itoa(panel_tram.Load) + "%\n\n"
	}

	msg.Text = statusStr
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Actualizar", "refresh"),
		),
	)

	f(msg)
}

func buildKeyboard(stops mapset.Set[metro_tenerife.Stop]) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	input := stops.ToSlice()
	for i := 1; i < len(input); i += 2 {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(input[i-1].Name, input[i-1].Id),
			tgbotapi.NewInlineKeyboardButtonData(input[i].Name, input[i].Id),
		})
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func fetch(panels *metro_tenerife.Panels, trams *metro_tenerife.Trams, stops mapset.Set[metro_tenerife.Stop], panel_trams *[]metro_tenerife.PanelTram) {
	stops.Clear()
	*panel_trams = (*panel_trams)[:0]

	metro_tenerife.GetJson("https://tranviaonline.metrotenerife.com/api/infoStops/infoPanel", panels)

	metro_tenerife.GetJson("https://tranviaonline.metrotenerife.com/api/cargaVehiculo", trams)

	for _, panel := range *panels {
		stops.Add(metro_tenerife.Stop{Id: panel.Stop, Name: panel.StopDescription})

		panel_tram := metro_tenerife.PanelTram{
			Stop:                       panel.Stop,
			StopDescription:            panel.StopDescription,
			DestinationStopDescription: panel.DestinationStopDescription,
			RemainingMinutes:           panel.RemainingMinutes,
		}

		for _, tram := range *trams {
			if strconv.Itoa(tram.Service) == panel.Service {
				panel_tram.Load = tram.Load
			}
		}

		*panel_trams = append(*panel_trams, panel_tram)
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("TELEGRAM_TOKEN")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	panels := new(metro_tenerife.Panels)
	trams := new(metro_tenerife.Trams)
	stops := mapset.NewSet[metro_tenerife.Stop]()
	panel_trams := make([]metro_tenerife.PanelTram, 0)
	select_stop := ""

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		if update.Message != nil {
			switch update.Message.Text {
			case "/start":
				bot.Request(tgbotapi.DeleteMessageConfig{
					ChatID:    update.Message.Chat.ID,
					MessageID: update.Message.MessageID,
				})

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
				msg.Text = "Use /start para iniciar el bot.\nUse /nexttram para obtener información acerca del siguiente tranvía.por cada parada."

				sendMessage(msg, bot)
			case "/nexttram", "/nexttram@TenerifeNextTramBot":
				fetch(panels, trams, stops, &panel_trams)

				bot.Request(tgbotapi.DeleteMessageConfig{
					ChatID:    update.Message.Chat.ID,
					MessageID: update.Message.MessageID,
				})

				keyboard := buildKeyboard(stops)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Seleccionar una parada:")
				msg.ReplyMarkup = keyboard

				sendMessage(msg, bot)
			}
		} else if update.CallbackQuery != nil {
			selected_panel_trams := make([]metro_tenerife.PanelTram, 0)

			bot.Request(tgbotapi.DeleteMessageConfig{
				ChatID:    update.CallbackQuery.Message.Chat.ID,
				MessageID: update.CallbackQuery.Message.MessageID,
			})

			switch update.CallbackQuery.Data {
			case "refresh":
				fetch(panels, trams, stops, &panel_trams)
			default:
				select_stop = update.CallbackQuery.Data
			}

			for _, panel_tram := range panel_trams {
				if panel_tram.Stop == select_stop {
					selected_panel_trams = append(selected_panel_trams, panel_tram)
				}
			}

			if len(selected_panel_trams) == 0 {
				continue
			}

			buildMessageTram(selected_panel_trams, update, func(msg tgbotapi.MessageConfig) {
				sendMessage(msg, bot)
			})
		}
	}
}
