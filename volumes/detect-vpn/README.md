# detect-vpn Service Configuration

## Workflow

Initially the container gets the `detect-vpn` volume mounted at a known path.
Afterwards the application in the container is started and it initially reads all of the files that are inside of the `manual-blacklist` folder, creates a new text file within the `blacklist` folder that should not be touched.
This newly created text file `blacklist-manual.txt` contains IPs/IP ranges of all of the individually specified text files in the `manual-blacklist` folder.
Then all of the files in the `blacklist` folder are read and added to the Redis cache via the *goripr* library.

After all of the blacklisted IPs are added to the cache, all of the text files in the `manual-whitelist` folder are read and removed from the cache, making only the blacklisted IPs trigger a ban.

In parallel, the **detect-vpn** is connected to a initially configured Discord channel that can handle ACL restricted command execution.
The discord bot connected to that channel checks every message that is typed for potential whitelist/blacklist code blocks.

```
\```blacklist
#comments
127.0.0.1-128.0.0.2 # vpn zcat.ch/bans
127.0.0.1/24 #vpn zcat.ch/bans
127.0.0.1
\```

\```whitelist
#comments
127.0.0.1-128.0.0.2 # optional comment
127.0.0.1/24 # optional comment
127.0.0.1 # optional comment
\```
```
If the command matches an IP/IP range/IP # reason/IP range #reason then those IPs are added to the Redis cache as well as written to the `manual-blacklist`.

The actual

The `blacklist` folder is supposed to contain individual text files that contain single IPs, IP ranges or custom IP ranges followed by a # and a reason string.
If no reason string is specified, the fallback reason for creating ban events is the one specified 