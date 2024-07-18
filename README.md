## One Time Share ([1ts.dev](https://1ts.dev))

A simple web service written in golang that handles sharing self-destructing messages.

[![Build](https://github.com/gameraccoon/one-time-share/actions/workflows/build.yml/badge.svg)](https://github.com/gameraccoon/one-time-share/actions/workflows/build.yml)

This service is useful when you want to share content-sensitive information in a conversation with someone, and you don't want this information to persist there.

Example:
> [me]: Hey Dave, can you unlock my computer so I can remote login into it? My pin: [URL]  
> [Dave]: Sure!

If Dave's computer gets hacked the next day, the person who gains access won't be able to see your pin from this conversation.

The service address: [1ts.dev](https://1ts.dev)

## Setting up your own server

1. Clone the repository
2. Edit `app-config.json` to set paths to your TLS certificate and key, or set `forceUnprotectedHttp` to `true` in case you enable HTTPS through a reverse proxy like nginx
3. Configure `app-config.json` with your desired port and limits.
4. `go build` to build the executable or `go run` to run it directly
6. Optionally, use tools/run_daemon.sh to start the service in the background, or configure it to run as you would any other service.

Refer to [build.yaml](https://github.com/gameraccoon/one-time-share/blob/main/.github/workflows/build.yml) for insights on how to build the project.

### Things to consider when setting up your own server
- Make sure your server runs under HTTPS and is not accessible via HTTP
  - Using HTTP is as good as broadcasting your private data to everyone in your network
- Whether you plan to deploy this web service or develop your own for your business, this service can be an easy point of entry for hackers to access other systems. You should ensure that no important information (such as access tokens or permanent passwords) is shared and the service is secured no less than other sensitive parts of your network.
  - If someone hacks into my server and finds random data without context, that's one thing. But it's a very different situation if they can understand who the data is shared by and intended for (or potentially even more context about this information if the hackers already have access to some other systems).
