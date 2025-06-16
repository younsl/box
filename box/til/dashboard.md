## 개요

독립적으로 Docker container를 실행하는 EC2를 위한 전용 Loki 대시보드 샘플

### 대시보드

```json
{
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": {
          "type": "grafana",
          "uid": "-- Grafana --"
        },
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "type": "dashboard"
      }
    ]
  },
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 0,
  "id": 65,
  "links": [],
  "liveNow": false,
  "panels": [
    {
      "datasource": {
        "type": "loki",
        "uid": "df49a8de-8db2-47a6-b40d-b9950a86dea9"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "bars",
            "fillOpacity": 100,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "insertNulls": false,
            "lineInterpolation": "linear",
            "lineWidth": 2,
            "pointSize": 6,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "auto",
            "spanNulls": false,
            "stacking": {
              "group": "A",
              "mode": "normal"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "red",
                "value": ""
              }
            ]
          }
        },
        "overrides": [
          {
            "matcher": {
              "id": "byName",
              "options": "{container_severity=\"DEBUG\"}"
            },
            "properties": [
              {
                "id": "color",
                "value": {
                  "fixedColor": "green",
                  "mode": "fixed"
                }
              }
            ]
          },
          {
            "matcher": {
              "id": "byName",
              "options": "{container_severity=\"INFO\"}"
            },
            "properties": [
              {
                "id": "color",
                "value": {
                  "fixedColor": "blue",
                  "mode": "fixed"
                }
              }
            ]
          },
          {
            "matcher": {
              "id": "byName",
              "options": "{container_severity=\"ERROR\"}"
            },
            "properties": [
              {
                "id": "color",
                "value": {
                  "fixedColor": "light-red",
                  "mode": "fixed"
                }
              }
            ]
          },
          {
            "matcher": {
              "id": "byName",
              "options": "{}"
            },
            "properties": [
              {
                "id": "color",
                "value": {
                  "mode": "fixed"
                }
              }
            ]
          }
        ]
      },
      "gridPos": {
        "h": 8,
        "w": 24,
        "x": 0,
        "y": 0
      },
      "id": 4,
      "interval": "1m",
      "maxDataPoints": 100,
      "options": {
        "legend": {
          "calcs": [],
          "displayMode": "list",
          "placement": "bottom",
          "showLegend": true
        },
        "timezone": [
          "browser"
        ],
        "tooltip": {
          "mode": "single",
          "sort": "none"
        }
      },
      "targets": [
        {
          "datasource": {
            "type": "loki",
            "uid": "df49a8de-8db2-47a6-b40d-b9950a86dea9"
          },
          "editorMode": "code",
          "expr": "sum by(container_severity) (count_over_time({container_name=\"$container_name\", container_port=\"$container_port\"} [$__interval]))",
          "legendFormat": "",
          "queryType": "range",
          "refId": "A"
        }
      ],
      "title": "Log level distribution",
      "type": "timeseries"
    },
    {
      "datasource": {
        "type": "loki",
        "uid": "df49a8de-8db2-47a6-b40d-b9950a86dea9"
      },
      "gridPos": {
        "h": 24,
        "w": 24,
        "x": 0,
        "y": 8
      },
      "id": 3,
      "options": {
        "dedupStrategy": "none",
        "enableLogDetails": true,
        "prettifyLogMessage": false,
        "showCommonLabels": false,
        "showLabels": false,
        "showTime": false,
        "sortOrder": "Descending",
        "wrapLogMessage": false
      },
      "pluginVersion": "10.1.5",
      "targets": [
        {
          "datasource": {
            "type": "loki",
            "uid": "df49a8de-8db2-47a6-b40d-b9950a86dea9"
          },
          "editorMode": "code",
          "expr": "{container_name=\"$container_name\", container_severity=~\"$container_severity\", container_port=\"$container_port\"} |= \"\"",
          "queryType": "range",
          "refId": "A"
        }
      ],
      "title": "Containers log",
      "type": "logs"
    }
  ],
  "refresh": "",
  "schemaVersion": 38,
  "style": "dark",
  "tags": [
    "loki",
    "ec2",
    "ec2-log"
  ],
  "templating": {
    "list": [
      {
        "current": {
          "isNone": true,
          "selected": false,
          "text": "None",
          "value": ""
        },
        "datasource": {
          "type": "loki",
          "uid": "df49a8de-8db2-47a6-b40d-b9950a86dea9"
        },
        "definition": "",
        "hide": 0,
        "includeAll": false,
        "label": "container_name",
        "multi": false,
        "name": "container_name",
        "options": [],
        "query": {
          "label": "container_name",
          "refId": "LokiVariableQueryEditor-VariableQuery",
          "stream": "",
          "type": 1
        },
        "refresh": 1,
        "regex": "",
        "skipUrlSync": false,
        "sort": 0,
        "type": "query"
      },
      {
        "current": {
          "selected": true,
          "text": [],
          "value": []
        },
        "datasource": {
          "type": "loki",
          "uid": "df49a8de-8db2-47a6-b40d-b9950a86dea9"
        },
        "definition": "",
        "description": "Container log severity",
        "hide": 0,
        "includeAll": false,
        "label": "container_severity",
        "multi": true,
        "name": "container_severity",
        "options": [],
        "query": {
          "label": "container_severity",
          "refId": "LokiVariableQueryEditor-VariableQuery",
          "stream": "{container_name=\"$container_name\"}",
          "type": 1
        },
        "refresh": 1,
        "regex": "/.*|^$/",
        "skipUrlSync": false,
        "sort": 0,
        "type": "query"
      },
      {
        "current": {
          "selected": false,
          "text": "3000",
          "value": "3000"
        },
        "datasource": {
          "type": "loki",
          "uid": "df49a8de-8db2-47a6-b40d-b9950a86dea9"
        },
        "definition": "",
        "hide": 0,
        "includeAll": false,
        "label": "container_port",
        "multi": false,
        "name": "container_port",
        "options": [],
        "query": {
          "label": "container_port",
          "refId": "LokiVariableQueryEditor-VariableQuery",
          "stream": "{container_name=\"$container_name\"}",
          "type": 1
        },
        "refresh": 1,
        "regex": "",
        "skipUrlSync": false,
        "sort": 0,
        "type": "query"
      }
    ]
  },
  "time": {
    "from": "now-6h",
    "to": "now"
  },
  "timepicker": {},
  "timezone": "",
  "title": "[Loki] EC2 container log",
  "uid": "cd9ba2a4-aabb-4370-992a-ff5757ebb4e8",
  "version": 26,
  "weekStart": ""
}
```

### 스크린샷

<img width="1906" alt="붙여넣은_이미지_2024__9__24__오후_7_06" src="https://github.com/user-attachments/assets/50414655-2a8d-4316-9f4f-46b061dc3a73">
