# Cursor 확장 프로그램 내보내기 및 가져오기

## 개요

이 문서는 Cursor 에디터의 확장 프로그램(extension)을 내보내고 가져오는 방법에 대해 설명합니다. 주로 다음 두 가지 상황에서 유용합니다:

1. 기존 머신에서 새로운 머신으로 Cursor 설정을 이전할 때
2. Cursor 설정을 백업하고 복원할 때

아래에서는 명령줄 인터페이스(CLI)를 사용하여 확장 프로그램을 관리하는 구체적인 단계를 제공합니다.

## 확장 프로그램 내보내기 및 가져오기

### 기존 머신에서 확장 프로그램 내보내기

다음 명령을 사용하여 현재 설치된 모든 확장 프로그램의 목록을 파일로 내보냅니다:

기존 기기에서:

```bash
cursor --list-extensions > extensions.list
```

새 기기에서:

```bash
cursor extensions.list | xargs -L 1 cursor --install-extension
```
