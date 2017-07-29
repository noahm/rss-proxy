# RSS Proxy

A simple proxy server for RSS feeds.

Supports serving protected feeds to clients without support
for basic HTTP auth by injecting into the request. (Overcast on iOS)

Supports filtering RSS feed items out of the respopnse to hide items,
matching regular expressions against specific fields on items.

## Useage

  $ cp server.conf.dist server.conf

Edit `server.conf` to add your feeds in sections of their own.
