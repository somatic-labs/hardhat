# hardhat

## How to Install

```bash
git clone https://github.com/somatic-labs/hardhat
cd hardhat
go install ./...
```

## How to use

Pre-baked mainnet configurations are in the configurations folder.  You need a file named `seedphrase`.  You may or may not want to set up your own node with a 10 GB mempool that accepts 50,000 transactions.  You put RPC urls in nodes.toml, and configure the other settings in there.  Then you just run `hardhat` in the same folder as the nodes.toml file and the `seedphrase` file.

## Context

I've known that the spammy attack can be enhanced by maybe 100x.  This does that, and it also works with https://github.com/notional-labs/rpc-crawler.

spammy-go takes the output of the rpc-crawler, which is open RPCs.  

Then, spammy-go will blast 30kb IBC transfers into every rpc at once at top speed.  

Specificially this exploits:

* banana king
* p2p-storms



## Bonus content

* All of my [emails with the interchain foundation](./emails)
* reports to the interchain foundation about this issue dating back to 2021



