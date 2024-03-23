package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

var LOG = log.New(os.Stderr, "", log.LstdFlags)

func eventHandler(evt interface{}) {
	go StartHandler(evt)
	go PingHandler(evt)
	go QuoteHandler(evt)
}

func StartHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		if v.Message.Conversation != nil {
			if strings.Contains(strings.ToLower(*v.Message.Conversation), "!start") {
				client.SendMessage(context.Background(), v.Info.Chat, &proto.Message{Conversation: ind("Hello! I'm Valeri :D")})
				return
			}
		}
	}
}

func PingHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		if v.Message.Conversation != nil {
			if strings.Contains(strings.ToLower(*v.Message.Conversation), "!ping") {
				startTimeT := time.Now()
				msg, _ := client.SendMessage(context.Background(), types.JID{User: v.Info.Chat.User, Server: v.Info.Chat.Server}, &proto.Message{Conversation: (func(s string) *string { return &s })("Pong!")})
				elapsedTime := time.Since(startTimeT)
				client.SendMessage(context.Background(), v.Info.Chat, client.BuildEdit(v.Info.Chat, msg.ID, &proto.Message{Conversation: ind(fmt.Sprintf("Pong! (took %v)\nUptime: %v", elapsedTime, time.Since(startTime)))}))
				return
			}
		}
	}
}

func QuoteHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		isQuoteEvent := false
		if v.Message.ExtendedTextMessage != nil && strings.Contains(strings.ToLower(*v.Message.ExtendedTextMessage.Text), "!q") {
			if v.Message.ExtendedTextMessage.ContextInfo != nil && v.Message.ExtendedTextMessage.ContextInfo.QuotedMessage != nil {
				isQuoteEvent = true
			}
		}

		if !isQuoteEvent {
			return
		}

		var text string
		if v.Message.ExtendedTextMessage.ContextInfo.QuotedMessage.ExtendedTextMessage != nil {
			text = *v.Message.ExtendedTextMessage.ContextInfo.QuotedMessage.ExtendedTextMessage.Text
		} else {
			text = *v.Message.ExtendedTextMessage.ContextInfo.QuotedMessage.Conversation
		}

		if text == "" {
			client.SendMessage(context.Background(), v.Info.Chat, &proto.Message{Conversation: ind("No message to quote!")})
		}

		user, err := types.ParseJID(*v.Message.ExtendedTextMessage.ContextInfo.Participant)
		if err != nil {
			LOG.Println(err)
			return
		}

		userObj, err := client.Store.Contacts.GetContact(user)
		if err != nil {
			userObj = types.ContactInfo{
				FullName: user.User,
			}
		}

		profile, err := client.GetProfilePictureInfo(user, &whatsmeow.GetProfilePictureParams{})
		if err != nil {
			profile = &types.ProfilePictureInfo{
				URL: "https://via.placeholder.com/100",
			}
		}

		req, err := http.NewRequest("POST", "http://localhost:3000/generate", strings.NewReader(perparePayload(userObj, profile, text)))
		if err != nil {
			LOG.Println(err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			LOG.Println(err)
			return
		}
		defer resp.Body.Close()
		var result struct {
			Result struct {
				Image string `json:"image"`
			} `json:"result"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			LOG.Println(err)
			return
		}
		base64Decoded, err := base64.StdEncoding.DecodeString(result.Result.Image)
		if err != nil {
			LOG.Println(err)
			return
		}
		base64Decoded, err = WebpImagePad(base64Decoded, 128, 384, 1)
		if err != nil {
			LOG.Println(err)
			return
		}
		ul, err := client.Upload(context.Background(), base64Decoded, whatsmeow.MediaImage)
		if err != nil {
			LOG.Println(err)
			return
		}

		whatsimage := &proto.StickerMessage{
			Mimetype:      ind("image/webp"),
			Url:           &ul.URL,
			DirectPath:    &ul.DirectPath,
			MediaKey:      ul.MediaKey,
			FileEncSha256: ul.FileEncSHA256,
			FileSha256:    ul.FileSHA256,
			FileLength:    &ul.FileLength,
			Width:         ini(2048),
			Height:        ini(512),
			IsAvatar:      inb(true),
			IsAiSticker:   inb(true),
		}

		_, err = client.SendMessage(context.Background(), v.Info.Chat, &proto.Message{StickerMessage: whatsimage})
		if err != nil {
			LOG.Println(err)
			return
		}
	}
}

func perparePayload(userObj types.ContactInfo, profile *types.ProfilePictureInfo, text string) string {
	// replace newlines with \n
	text = strings.ReplaceAll(text, "\n", "\\n")
	return fmt.Sprintf(`{"type":"quote","format":"webp","width":512,"height":768,"scale":2,"messages":[{"entities":[],"avatar":true,"from":{"id":1,"name":"%s","photo":{"url":"%s"}},"text":"%s","replyMessage":{}}]}`, userObj.FullName, profile.URL, text)
}
