# fluentd exclude

## 개요

fluentd에서 특정 컨테이너 로그 수집 제외하기

## 설정방법

`fluentd`에서 특정 컨테이너에 대한 로그 수집을 예외처리하려면 `exclude` 섹션에 추가합니다.

```xml
<filter kube.**>
  @type grep

  <exclude>
    key log
    pattern /(ELB-HealthChecker|health|actuator|statusz|healthz)/i
  </exclude>

  <exclude>
    key $["kubernetes"]["container_name"]
    pattern /(CONTAINER_NAME_A|CONTAINER_NAME_B)/i
  </exclude>
</filter>
```

`/i`는 정규 표현식에서 대소문자를 구분하지 않도록 하는 플래그입니다. 만약 대소문자를 구분하지 않고 매칭하고 싶다면 `/i`를 유지해야 합니다. 대소문자를 구분하고 싶다면 `/i`를 제거해도 됩니다. 자세한 사항은 [Specifying Modes Inside The Regular Expression](https://www.regular-expressions.info/modifiers.html) 페이지를 참고합니다.

대소문자 구분 안하는 `fluentd` 설정:

```xml
<exclude>
  key $["kubernetes"]["container_name"]
  pattern /(CONTAINER_NAME_A|CONTAINER_NAME_B)/i
</exclude>
```

예를 들어, `CONTAINER_NAME_A`와 `container_name_a`를 모두 매칭하고 싶다면 `/i`를 유지해야 하고, `CONTAINER_NAME_A`만 매칭하고 싶다면 `/i`를 제거하면 됩니다.

대소문자를 명확하게 구분하는 `fluentd` 설정:

```xml
<exclude>
  key $["kubernetes"]["container_name"]
  pattern /(CONTAINER_NAME_A|CONTAINER_NAME_B)/
</exclude>
```

완성된 `fleuntd`의 `output.conf` 설정:

```yaml
configMaps:
  output.conf: |
    <filter kube.**>
      @type grep

      <exclude>
        key log
        pattern /(ELB-HealthChecker|health|actuator|statusz|healthz)/i
      </exclude>

      <exclude>
        key $["kubernetes"]["container_name"]
        pattern /(CONTAINER_NAME_A|CONTAINER_NAME_B)/i
      </exclude>
    </filter>

    <filter kube.**>
      @type record_transformer
      remove_keys $["kubernetes"]["labels"], $["kubernetes"]["annotations"], $["kubernetes"]["pod_id"], $["kubernetes"]["docker_id"], $["kubernetes"]["container_hash"]
    </filter>

    <filter kube.**>
      @type geoip

      geoip_lookup_keys ip
      backend_library geoip2_c

      <record>
        ip_location       ${location.latitude["ip"]},${location.longitude["ip"]}
        ip_country        ${country.iso_code["ip"]}
        ip_countryname    ${country.names.en["ip"]}
        ip_cityname       ${city.names.en["ip"]}
      </record>

      skip_adding_null_record  true
    </filter>

    <match kube.**>
      @type copy

      <store>
        @id elasticsearch
        @type aws-elasticsearch-service
        @log_level error
        include_tag_key false
        tag_key fluentd_tag

        <endpoint>
          url "#{ENV['OUTPUT_ES_HOST']}"
          region "#{ENV['OUTPUT_ES_REGION']}"
        </endpoint>

        logstash_format true
        logstash_prefix "#{ENV['OUTPUT_ES_INDEX_PREFIX']}"
        type_name "#{ENV['OUTPUT_ES_TYPE_NAME']}"

        <buffer>
          @type file
          path /var/log/fluentd-buffers/es.buffer
          flush_mode interval
          retry_type exponential_backoff
          flush_thread_count 4
          flush_interval 30s
          retry_forever
          retry_max_interval 60
          compress text
          chunk_limit_size "#{ENV['OUTPUT_ES_BUFFER_CHUNK_LIMIT']}"
          queue_limit_length "#{ENV['OUTPUT_ES_BUFFER_QUEUE_LIMIT']}"
          overflow_action block
        </buffer>
      </store>

      <store>
        @id s3
        @type s3
        @log_level error
        include_tag_key false
        tag_key fluentd_tag

        <assume_role_credentials>
          role_arn          "#{ENV['OUTPUT_S3_ROLE_ARN']}"
          role_session_name "#{ENV['OUTPUT_S3_ROLE_SESSION_NAME']}"
        </assume_role_credentials>

        s3_bucket "#{ENV['OUTPUT_S3_BUCKET']}"
        s3_region "#{ENV['OUTPUT_S3_REGION']}"
        path "#{ENV['OUTPUT_S3_PATH']}"
        store_as gzip
        time_slice_format %Y%m%d
        s3_object_key_format "%{path}/%Y/%m/%d/%{time_slice}_%{index}.%{file_extension}"

        <format>
          @type json
        </format>

        slow_flush_log_threshold 300.0

        <buffer>
          @type file
          path /var/log/fluentd-buffers/s3.buffer
          flush_mode interval
          retry_type exponential_backoff
          flush_thread_count 8
          flush_interval 3600s
          retry_forever
          retry_max_interval 30
          chunk_limit_size "#{ENV['OUTPUT_S3_BUFFER_CHUNK_LIMIT']}"
          queue_limit_length "#{ENV['OUTPUT_S3_BUFFER_QUEUE_LIMIT']}"
          overflow_action block
        </buffer>
      </store>
    </match>
```

## 참고자료

- [fluentd filter.grep 공식문서](https://docs.fluentd.org/filter/grep)
