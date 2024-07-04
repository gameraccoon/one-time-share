## One Time Share ([1ts.dev](https://1ts.dev))

A simple web service written in golang that handles sharing self-destructing messages.

[![Build](https://github.com/gameraccoon/one-time-share/actions/workflows/build.yml/badge.svg)](https://github.com/gameraccoon/one-time-share/actions/workflows/build.yml)

This service is useful when you want to share content-sensitive information in a conversation with someone, and you don't want this information to persist there.

Example:
> [me]: Hey Dave, can you unlock my computer so I can remote login into it? My pin: [URL]  
> [Dave]: Sure!

If Dave's computer gets hacked the next day, the person who gains access won't be able to see your pin from this conversation.

### What does the service guarantee?

- The shared message can be accessed by the link only once; the message is removed from the server even before it reaches the recipient.
- The message will not be accessible after the set time and will be removed from the server soon after it expires.
- TLS encryption ensures that the message cannot be intercepted on the way to the server or from the server to the recipient.

### What this service does not guarantee

This service does not guarantee that a malicious user cannot access the message. If your messenger is vulnerable to a "man-in-the-middle" attack, or if you or your recipient have spyware on your machines, this service won't prevent information leaks.

Additionally, if the link falls into the hands of a malicious user before the recipient accesses it, the data can be exposed to that malicious user.

And, as with most services today, this service cannot guarantee that it will never be hacked or that my hosting provider will not access the data on the server. Below is some information on how to deal with these risks.

### What limitations does the service have?

- Since the message is removed before it is sent to the recipient, delivery is not guaranteed. If the recipient has a very poor internet connection, the message may not be delivered and will need to be resent.
- The service does not make a distinction between read, expired, and non-existing messages.
- Only text messages are supported; no files.
- Only HTTPS is supported; no HTTP.
- To avoid spam and overuse, there is a very strict limit on how often messages can be created, and this limit is shared among all users.
- The size of the message is restricted.
- Maximum possible message storage time is 30 days (on your own server, you can disable all these limits).

### What information is stored on the server

- Message (basically in plain text)
- The expiry time of the message
- The token associated with the message

### Can I set up my own server?

Yes. The server is quite easy to set up and run on your machine. It has minimal dependencies, and the code should be straightforward to review.

### What about that "man in the middle" thing?

In simple terms, if you exchange messages over a channel that allows someone to intercept and potentially replace messages, this channel is vulnerable to a "man-in-the-middle" attack.

As of 2024, most messengers and other online communication methods are vulnerable to this type of attack, except those that use end-to-end encryption.

This service as well would store the sent messages basically in plain text (until they are accessed or expire), so if someone hacks into the server, they can intercept the messages in transit.

### OK, what should I do about it?

If you care about keeping your information secure, you should either use a trusted [end-to-end encrypted](https://en.wikipedia.org/wiki/End-to-end_encryption) messenger or utilize [public-key cryptography](https://en.wikipedia.org/wiki/Email_encryption) yourself.

**If you don't use end-to-end encryption, your conversations can be accessed by other people.**

### What if the service gets hacked?

There’s a simple rule: when sending information through the 1ts service, don’t add any context to your secret data.

Bad:
> Hi, by the link you will find your login and password, I've also added the URL of the website there just in case.  
> [link]

Good:
> Hi, your login is User1447, your password is by the link below. The URL is example.com.  
> [link]

This makes a difference if our server gets hacked. In the first case, the hacker would get everything they need to access the account. In the second case, they would get just a set of symbols without any idea where it can be applied.

## Your own server set up

1. Clone the repository
2. Edit `app-config.json` to set paths to your TLS certificate and key, or set `forceUnprotectedHttp` to `true` in case you enable HTTPS through a reverse proxy like nginx
3. Configure `app-config.json` with your desired port and limits.
4. `go build` to build the executable or `go run` to run it directly
6. Optionally, use tools/run_daemon.sh to start the service in the background, or configure it to run as you would any other service.

Refer to [build.yaml](https://github.com/gameraccoon/one-time-share/blob/main/.github/workflows/build.yml) for insights on how to build the project.

### Things to think about when setting up your own server
- Make sure your server runs under HTTPS and is not accessible via HTTP
  - Using HTTP is as good as broadcasting your private data to everyone in your network
- Whether you plan to deploy this web service or develop your own for your business, this service can be an easy point of entry for hackers to access other systems. You should ensure that no important information (such as access tokens or permanent passwords) is shared and the service is secured no less than other sensitive parts of your network.
  - If someone hacks into my server and finds random data without context, that's one thing. But it's a very different situation if they can understand who the data is shared by and intended for (or potentially even more context about this information if the hackers already have access to some other systems).
