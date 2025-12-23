# feedgen

A web service to generate Atom feeds from websites.

## Supported Sites and Examples

- 大管家房屋網
  - https://feedgen.org/chrb
- 巴哈姆特-哈拉區
  - 精靈寶可夢: https://feedgen.org/gamer_forum?bsn=1647
  - 魔物獵人 (20 推以上): https://feedgen.org/gamer_forum?bsn=5786&gp=20
- Hacker News
  - Top links: https://feedgen.org/hackernews?category=best
- HackMD
  - Published notes: https://feedgen.org/hackmd?u=@hackmd-blog
- 批踢踢實業坊
  - Steam 版: https://feedgen.org/ptt?b=Steam
  - 電影版 (30 推以上): https://feedgen.org/ptt?b=movie&q=recommend:30
- 遊戲角落
  - 最新文章: https://feedgen.org/udn_game?section=rank&by=newest
  - 最多瀏覽: https://feedgen.org/udn_game?section=rank&by=pv

## Getting Started

### Prerequisites

- Docker (version 20.10 or later)
- Docker Compose (version 2.0 or later)

### Installation

Build the Docker image:

```bash
docker compose build
```

Start the application with Docker Compose:

```bash
docker compose up -d
```

The service is available at `http://localhost:8080`.

## License

feedgen is released under the MIT license.
