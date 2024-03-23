# WaQuoteBot

WaQuoteBot is a WhatsApp bot that can create stickers from text messages. It is built using the Whatsmeow library, Quote API, and a combination of Golang and Node.js.

## Installation

Before getting started, make sure you have the following prerequisites installed:

- libwebp-dev: This library is required for working with WebP images, which are used for stickers.

### Installing libwebp-dev

- To install libwebp-dev, you can use the package manager for your operating system. For example, on Ubuntu, you can run the following command:

```bash
sudo apt-get install libwebp-dev
```

### Installing WaQuoteBot

- Setup and Run the quote-api server

```bash
cd quote-api
npm install
npm .
```

- Setup and Run the WaQuoteBot
    
```bash
go build -o waquotebot
./waquotebot
```

## Commands

- `!start`: check if the bot is running
- `!ping`: check the response time of the bot
- `!q`: create a sticker from a quote

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
