# Infrastructure For Sales

세일즈와 프리 세일즈를 위한 클라우드 네이티브 인프라 강의 자료입니다.

## Usage

`.pdf` 파일 확장자 형태로 현재 작성한 Marp 발표자료를 저장합니다.

> [!NOTE]
> 만약 해당 발표자료가 로컬 파일을 참조하는 경우 `--allow-local-files` 옵션을 추가해서 파일을 참조할 수 있도록 해야 합니다.

```bash
# Convert slide deck into PDF (using Chromium in Docker)
docker run --rm --init -v $PWD:/home/marp/app/ -e LANG=$LANG marpteam/marp-cli slide-deck.md --pdf --allow-local-files
```

`-s` 옵션을 통해 서버 모드로 실행할 수 있습니다.

```bash
# Server mode (Serve current directory in http://localhost:8080/)
docker run --rm --init -v $PWD:/home/marp/app -e LANG=$LANG -p 8080:8080 -p 37717:37717 marpteam/marp-cli -s .
```

실행 후 웹 브라우저를 열어 [http://localhost:8080/](http://localhost:8080/)에서 프리젠테이션을 시연할 수 있습니다.
