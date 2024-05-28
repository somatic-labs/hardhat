# Cures

Assaf Morami of Secret Network has provided me with an example transaction that adjusts the maximum block size by governance.  

I give him huge credit.  Both myself and Jehan Tremback were not aware this was possible. 

## Note

This cure doesn't seem to work.  While reducing block sizes dramatically reduces risk and I do recommend it to ensure that block gossip is not too impactful as a % of p2p traffic, it seems that this attack can be replicated with only mempool p2p traffic. 

Still probably a good idea. 


```bash
appd tx gov submit-proposal param-change ./path/to/cure.json --from key -y -b block
```

example used on Osmosis

```bash
osmosisd tx gov submit-proposal param-change cure.json --from icns --keyring-backend file --fees 2000uosmo
```



```json
{
    "title": "Reduce Maximum Block Size",
    "description": "This Proposal reduces the maximum block size pursuant to: https://github.com/cometbft/cometbft/security/advisories/GHSA-hq58-p9mv-338c",
    "changes": [
        {
            "subspace": "baseapp",
            "key": "BlockParams",
            "value": {
                "max_bytes": "1048576"
            }
        }
    ],
    "deposit": "100000000uatom"
}
```
