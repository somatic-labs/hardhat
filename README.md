# spammy-go

## How to Install

```bash
git clone https://github.com/somatic-labs/hardhat
cd hardhat
go install ./...
```

## How to use

There is a `nodes.toml` file in the folder, it is currently configured for Sentinel.  This code supports SDK 47, but I have built sdk 50 and sdk 45 versions, you can [contact me](https://x.com/gadikian) by twitter DM and I can get you those, if you should need it.  

## Context

I've known that the spammy attack can be enhanced by maybe 100x.  This does that, and it also works with https://github.com/notional-labs/rpc-crawler.

spammy-go takes the output of the rpc-crawler, which is open RPCs.  

Then, spammy-go will blast 30kb IBC transfers into every rpc at once at top speed.  

Specificially this exploits:

* banana king
* p2p-storms

## Bonus content

* All of my emails with the interchain foundation
* reports to the interchain foundation about this issue dating back to 2021



