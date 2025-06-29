GHES의 레포를 풀링해오는 미러링(Migration) 구성시 발생하는 에러 메세지:

```
You cannot import from disallowed hosts, please ask the admin to check ALLOWED_DOMAINS/ALLOW_LOCALNETWORKS/BLOCKED_DOMAINS settings.
```

차트 설정:

```yaml
# helm/gitea/values_example.yaml
gitea:
  config:
    migrations:
      ALLOWED_LOCALNETWORKS: "true"
      ALLOWED_DOMAINS: "*" 
```

실제 gitea 파드의 설정파일에 반영된 Gitea 세부 설정:

```ini
# cat /data/gitea/conf/app.ini
[migrations]
ALLOWED_LOCALNETWORKS = true
ALLOWED_DOMAINS = *
```
