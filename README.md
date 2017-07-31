# RSS Proxy

A simple proxy server for RSS feeds. RIP Yahoo Pipes.

Supports serving protected feeds to clients without support
for basic HTTP auth by injecting into the request (Overcast on iOS).

Supports filtering RSS feed items out of the response to hide items,
matching regular expressions against specific fields on items.

## Usage

  $ cp server.conf.dist server.conf

Edit `server.conf` to add your feeds in sections of their own.
