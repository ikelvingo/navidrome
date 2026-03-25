##  Navidrome [China Special Edition] 

## Provide scrobbling artists and albums bio from netease.

## #~~You should use it with [navichina](https://github.com/TooAndy/navichina)~~ 

# Thanks for TooAndy's great work.

# #1139840: Remove navichina dependency in navidrome-chinese.

Input new 'netease' agent for scrobbling artists, albums, similar songs, 
and artist popular songs. 
- Note1: Similar artists functionality not supported.
- Note2: Configuration: Set the ND_AGENTS environment variable to 'netease' to activate the NetEase scrobbling agent.
    ```yaml
    # docker compose modify
      environment:
        - ND_AGENTS=netease #,deezer,lastfm,listenbrainz
    ```

-----
>  [!IMPORTANT]
>
> **引入OpenCC，终于统一了Navidrome中文繁简体搜索**
> 众所周知，Navidrome在检索管理的音乐时，简体中文仅能检索简体中文，繁体中文仅能检索繁体中文，如搜索“周杰伦”，只会搜索到“周杰伦”的结果，而无法搜索到“周杰倫”的结果。
> 本次更新，将实现无论搜索“周杰伦”还是“周杰倫”，系统会将“周杰伦”+“周杰倫”的所有搜索结果返回。从此你将不会在被繁简体检索的结果而烦恼。
> 本次更新的搜索功能，无论web端还是subsonic api接口均生效。

-----
>  [!IMPORTANT]
>
> **Added the forced refresh Artist data function, providing the following features:**

##  How to use

```bash
# Refresh via artist ID
 sudo docker exec -it navidrome refresh --id "xxxxx"

# Refresh via artist name (supports fuzzy matching)
 sudo docker exec -it navidrome refresh --name "Taylor Swift"

# Clear all external information and refresh
 sudo docker exec -it navidrome refresh --id "xxxxx" --clear-all

# Clear only the artist's image URLs
 sudo docker exec -it navidrome refresh --name "Taylor Swift" --clear-images

# Refresh all albums of the artist simultaneously
 sudo docker exec -it navidrome refresh --id "xxxxx" --albums --clear-all
```

## Available parameters

| Parameters        | **Instructions**                            |
| ----------------- | ------------------------------------------- |
| `--id`            | clear artist ID                             |
| `--name`          | clear artist name (supports fuzzy matching) |
| `--clear-images`  | clear image URLs                            |
| `--clear-bio`     | clear artist bio                            |
| `--clear-similar` | clear similar artists                       |
| `--clear-all`     | clear all external infomation               |
| `--albums`        | clear all artist’s albums                   |

After clearing, the next time you visit the artist's page, information will be fetched again from external sources (Last.fm, NetEase Cloud Music, etc.).

-----




