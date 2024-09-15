
# scan-schedule-workflow

모든 `schedule:` 트리거를 사용하는 Workflow를 검색하는 스크립트

> ChatGPT 4.0를 사용해서 작성된 스크립트입니다.

## 사용법

스크립트는 pygithub와 croniter를 사용하여 Github API를 호출하므로 사전에 패키지 설치가 필요합니다.

```console
$ pip install -r requirements.txt
...
[notice] A new release of pip is available: 24.0 -> 24.1.2
[notice] To update, run: python3.12 -m pip install --upgrade pip
```

이후 `repo:*` 권한이 부여된 Classic PAT를 생성한 후, 토큰 값을 기록해둡니다. 이후 스크립트 실행할 때 사용됩니다.

스크립트를 실행한 후 PAT 토큰 값을 입력합니다.

```console
$ python3 scan-schedule-workflow.py
Enter your GitHub Access Token: <ENTER_PAT_TOKEN>
```

## 사용 예시

스크립트를 실행하면 특정 Organization에 들어있는 모든 레포지터리의 `schedule:` 트리거를 사용하는 Actions Workflow를 찾고, cron 스케줄 값과 KST 시간 기준 스케줄 값을 출력합니다.

```console
$ python3 scan-schedule-workflow.py
Enter your GitHub Access Token: <ENTER_PAT_TOKEN>
Searching exampleorg/repo1 (1/921)...
Searching exampleorg/repo2 (2/921)...
Searching exampleorg/repo3 (3/921)...
exampleorg/dbt-metric - Found 'schedule:' in .github/workflows/schedule-workflow.yml
  Original cron: 0 23 * * *, KST Time: 2024-07-10 08:00:00
Searching exampleorg/repo4 (12/921)...
```
