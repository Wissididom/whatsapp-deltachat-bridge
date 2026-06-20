package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
	"google.golang.org/protobuf/proto"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"go.mau.fi/whatsmeow/proto/waE2E"
	waLog "go.mau.fi/whatsmeow/util/log"
	"github.com/mdp/qrterminal/v3"

	_ "github.com/mattn/go-sqlite3"
)

var dcBot deltachat.Bot
var waClient *whatsmeow.Client

func logEvent(bot *deltachat.Bot, accId uint32, event deltachat.EventType) {
	switch ev := event.(type) {
	case *deltachat.EventTypeInfo:
		log.Printf("INFO: %v", ev.Msg)
	case *deltachat.EventTypeWarning:
		log.Printf("WARNING: %v", ev.Msg)
	case *deltachat.EventTypeError:
		log.Printf("ERROR: %v", ev.Msg)
	}
}

func runBridgeBot(bot *deltachat.Bot, accId uint32) {
	sysinfo, _ := bot.Rpc.GetSystemInfo()
	log.Println("Running deltachat core", sysinfo["deltachat_core_version"])

	bot.On(&deltachat.EventTypeInfo{}, logEvent)
	bot.On(&deltachat.EventTypeWarning{}, logEvent)
	bot.On(&deltachat.EventTypeError{}, logEvent)
	bot.OnNewMsg(func(bot *deltachat.Bot, accId uint32, msgId uint32) {
		msg, _ := bot.Rpc.GetMessage(accId, msgId)
		if msg.FromId > deltachat.ContactLastSpecial {
			parts := strings.SplitN(msg.Text, " ", 2)
			if (len(parts) != 2) {
				log.Println("message does not contain WhatsApp sender info")
				return
			}
			jidStr := parts[0]
			body := strings.TrimSpace(parts[1])
			var targetJid types.JID
			var err error
			if strings.Contains(jidStr, "@") {
				targetJid, err = types.ParseJID(jidStr)
				if err != nil {
					log.Println(err)
					return
				}
			} else {
				targetJid = types.NewJID(jidStr, types.DefaultUserServer)
			}
			resp, err := waClient.SendMessage(
				context.Background(),
				targetJid,
				&waE2E.Message{
					Conversation: proto.String(body),
				},
			)
			log.Printf("resp=%+v err=%v", resp, err)
		}
	})

	if isConf, _ := bot.Rpc.IsConfigured(accId); !isConf {
		log.Println("Bot not configured, configuring...")
		botFlag := "1"
		if err := bot.Rpc.SetConfig(accId, "bot", &botFlag); err != nil {
			log.Fatalln(err)
		}
		if err := bot.Rpc.AddTransportFromQr(accId, os.Args[1]); err != nil {
			log.Fatalln(err)
		}
	}

	inviteLink, _ := bot.Rpc.GetChatSecurejoinQrCode(accId, nil)
	log.Println("Listening at:", inviteLink)
	config := qrterminal.Config{
		Level: qrterminal.M,
		Writer: os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	}
	qrterminal.GenerateWithConfig(inviteLink, config)
	if err := bot.Run(); err != nil {
		log.Fatalln(err)
	}
}

// Get the first available account or create a new one if none exists.
func getAccount(rpc *deltachat.Rpc) uint32 {
	accounts, _ := rpc.GetAllAccountIds()
	var accId uint32
	if len(accounts) == 0 {
		accId, _ = rpc.AddAccount()
	} else {
		accId = accounts[0]
	}
	return accId
}

func waEventHandler(evt any) {
	switch v := evt.(type) {
		case *events.Message:
			// https://github.com/tulir/whatsmeow/blob/089932318bc2/proto/waE2E/WAWebProtobufsE2E.pb.go#L10365
			conversation := v.Message.GetConversation();
			if conversation != "" {
				/*if v.Info.IsFromMe {
					break
				}*/
				senderName := v.Info.PushName
				senderId := v.Info.Sender.User
				chatId := v.Info.Chat
				text := fmt.Sprintf(
					"From: %s (%s) in Chat %s\n\n%s",
					senderName,
					senderId,
					chatId,
					conversation,
				)
				reply := deltachat.MessageData{Text: &text}
				if _, err := dcBot.Rpc.SendMsg(
					1,  //AccountId
					10, //ChatId
					reply,
				); err != nil {
					log.Printf("ERROR: %v", err)
				}
			}
			fmt.Println("WhatsApp-Message-ImageMessage:", v.Message.GetImageMessage())
			fmt.Println("WhatsApp-Message-ContactMessage:", v.Message.GetContactMessage())
			fmt.Println("WhatsApp-Message-LocationMessage:", v.Message.GetLocationMessage())
			fmt.Println("WhatsApp-Message-ExtendedTextMessage:", v.Message.GetExtendedTextMessage())
			fmt.Println("WhatsApp-Message-DocumentMessage:", v.Message.GetDocumentMessage())
			fmt.Println("WhatsApp-Message-AudioMessage:", v.Message.GetAudioMessage())
			fmt.Println("WhatsApp-Message-VideoMessage:", v.Message.GetVideoMessage())
			fmt.Println("WhatsApp-Message-ContactsArrayMessage:", v.Message.GetContactsArrayMessage())
			fmt.Println("WhatsApp-Message-HighlyStructuredMessage:", v.Message.GetHighlyStructuredMessage())
			fmt.Println("WhatsApp-Message-LiveLocationMessage:", v.Message.GetLiveLocationMessage())
			fmt.Println("WhatsApp-Message-TemplateMessage:", v.Message.GetTemplateMessage())
			fmt.Println("WhatsApp-Message-StickerMessage:", v.Message.GetStickerMessage())
			fmt.Println("WhatsApp-Message-GroupInviteMessage:", v.Message.GetGroupInviteMessage())
			fmt.Println("WhatsApp-Message-ReactionMessage:", v.Message.GetReactionMessage())
			fmt.Println("WhatsApp-Message-PollCreationMessage:", v.Message.GetPollCreationMessage())
			fmt.Println("WhatsApp-Message-PollUpdateMessage:", v.Message.GetPollUpdateMessage())
			fmt.Println("WhatsApp-Message-PollCreationMessageV2:", v.Message.GetPollCreationMessageV2())
			fmt.Println("WhatsApp-Message-PinInChatMessage:", v.Message.GetPinInChatMessage())
			fmt.Println("WhatsApp-Message-PollCreationMessageV3:", v.Message.GetPollCreationMessageV3())
			fmt.Println("WhatsApp-Message-AlbumMessage:", v.Message.GetAlbumMessage())
			fmt.Println("WhatsApp-Message-StickerPackMessage:", v.Message.GetStickerPackMessage())
			fmt.Println("WhatsApp-Message-PollCreationMessageV4:", v.Message.GetPollCreationMessageV4())
			fmt.Println("WhatsApp-Message-PollCreationMessageV5:", v.Message.GetPollCreationMessageV5())
			fmt.Println("WhatsApp-Message-PollCreationMessageV6:", v.Message.GetPollCreationMessageV6())
	}
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:/data/whatsapp.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	waClient = whatsmeow.NewClient(deviceStore, clientLog)
	waClient.AddEventHandler(waEventHandler)
	if waClient.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := waClient.GetQRChannel(context.Background())
		err = waClient.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				//fmt.Println("QR code:", evt.Code)
				config := qrterminal.Config{
					Level: qrterminal.M,
					Writer: os.Stdout,
					BlackChar: qrterminal.BLACK,
					WhiteChar: qrterminal.WHITE,
					QuietZone: 1,
				}
				qrterminal.GenerateWithConfig(evt.Code, config)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = waClient.Connect()
		if err != nil {
			panic(err)
		}
	}
	defer waClient.Disconnect()
	trans := deltachat.NewIOTransport()
	if err := trans.Open(); err != nil {
		log.Fatalln(err)
	}
	defer trans.Close()
	rpc := &deltachat.Rpc{Context: context.Background(), Transport: trans}
	dcBot = *deltachat.NewBot(rpc);
	runBridgeBot(&dcBot, getAccount(rpc))
}
