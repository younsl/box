# Infrastructure For Sales

세일즈와 프리 세일즈를 위한 클라우드 네이티브 인프라 강의 자료입니다.

## Usage

### How to publish

로컬 환경에 [marp-cli](https://github.com/marp-team/marp-cli)가 설치되어 있어야 합니다.

```bash
brew install marp-cli
marp --version
```

`marp` 명령어를 통해 프리젠테이션을 시연할 수 있습니다.

```bash
marp index.md --preview
```

`.pdf` 파일 확장자 형태로 현재 작성한 Marp 발표자료를 저장합니다.

> [!NOTE]
> 만약 해당 발표자료가 로컬 파일을 참조하는 경우 `--allow-local-files` 옵션을 추가해서 파일을 참조할 수 있도록 해야 합니다.

```bash
marp index.md --pdf --allow-local-files
```
